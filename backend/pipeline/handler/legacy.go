package handler

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	corev2 "github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/backend/licensing"
	"github.com/sensu/sensu-go/backend/secrets"
	"github.com/sensu/sensu-go/backend/store"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/util/environment"
	utillogging "github.com/sensu/sensu-go/util/logging"
)

const (
	// DefaultSocketTimeout specifies the default socket dial
	// timeout in seconds for TCP and UDP handlers.
	DefaultSocketTimeout uint32 = 60
)

// LegacyAdapter is a handler adapter that supports the legacy core.v2/Handler
// type.
type LegacyAdapter struct {
	AssetGetter            asset.Getter
	Executor               command.Executor
	LicenseGetter          licensing.Getter
	SecretsProviderManager secrets.ProviderManagerer
	Store                  store.Store
	StoreTimeout           time.Duration
}

// Name returns the name of the handler adapter.
func (l *LegacyAdapter) Name() string {
	return "LegacyAdapter"
}

// CanHandle determines whether LegacyAdapter can handle the resource being
// referenced.
func (l *LegacyAdapter) CanHandle(ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "Handler" {
		return true
	}
	return false
}

// Handle handles a Sensu event. It will pass any mutated data along to pipe or
// tcp/udp handlers.
func (l *LegacyAdapter) Handle(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event, mutatedData []byte) error {
	// Prepare log entry
	fields := utillogging.EventFields(event, false)
	fields["pipeline"] = corev2.ContextPipeline(ctx)
	fields["pipeline_workflow"] = corev2.ContextPipelineWorkflow(ctx)

	tctx, cancel := context.WithTimeout(ctx, l.StoreTimeout)
	handler, err := l.Store.GetHandlerByName(tctx, ref.Name)
	cancel()
	if err != nil {
		return fmt.Errorf("failed to fetch handler from store: %v", err)
	}

	switch handler.Type {
	case "pipe":
		result, err := l.pipeHandler(ctx, handler, event, mutatedData)
		if err != nil {
			logger.WithFields(fields).
				WithError(err).
				Error("failed to execute event pipe handler")
			return err
		}
		fields["status"] = result.Status
		fields["output"] = result.Output
		logger.WithFields(fields).Info("event pipe handler executed")
	case "tcp", "udp":
		_, err := l.socketHandler(ctx, handler, event, mutatedData)
		if err != nil {
			logger.WithFields(fields).Error(err)
			return err
		}
	default:
		return errors.New("unknown handler type")
	}

	return nil
}

// pipeHandler fork/executes a child process for a Sensu pipe handler command
// and writes the mutated data to it via STDIN.
func (l *LegacyAdapter) pipeHandler(ctx context.Context, handler *corev2.Handler, event *corev2.Event, mutatedData []byte) (*command.ExecutionResponse, error) {
	ctx = corev2.SetContextFromResource(ctx, handler)

	// Prepare log entry
	fields := utillogging.EventFields(event, false)
	fields["handler_name"] = handler.Name
	fields["handler_namespace"] = handler.Namespace
	fields["pipeline"] = corev2.ContextPipeline(ctx)
	fields["pipeline_workflow"] = corev2.ContextPipelineWorkflow(ctx)

	if l.LicenseGetter != nil {
		if license := l.LicenseGetter.Get(); license != "" {
			handler.EnvVars = append(handler.EnvVars, fmt.Sprintf("SENSU_LICENSE_FILE=%s", license))
		}
	}

	secrets := []string{}
	if l.SecretsProviderManager != nil {
		substituted, err := l.SecretsProviderManager.SubSecrets(ctx, handler.Secrets)
		if err != nil {
			logger.WithFields(fields).WithError(err).Error("failed to retrieve secrets for handler")
			return nil, err
		}
		secrets = append(secrets, substituted...)
	}

	// Prepare environment variables
	env := environment.MergeEnvironments(os.Environ(), handler.EnvVars, secrets)

	handlerExec := command.ExecutionRequest{}
	handlerExec.Command = handler.Command
	handlerExec.Timeout = int(handler.Timeout)
	handlerExec.Env = env
	handlerExec.Input = string(mutatedData[:])

	// Only add assets to execution context if handler requires them
	if len(handler.RuntimeAssets) != 0 {
		logger.WithFields(fields).Debug("fetching assets for handler")
		// Fetch and install all assets required for handler execution
		// TODO: check for errors here once GetAssets() has been updated to
		// return errors.
		// See issue #4407: https://github.com/sensu/sensu-go/issues/4407
		matchedAssets := asset.GetAssets(ctx, l.Store, handler.RuntimeAssets)

		assets, err := asset.GetAll(ctx, l.AssetGetter, matchedAssets)
		if err != nil {
			logger.WithFields(fields).WithError(err).Error("failed to retrieve assets for handler")
			// TODO(jk): I think we should return an error here regardless of
			// the type of error.
			// See issue #4407: https://github.com/sensu/sensu-go/issues/4407
			if _, ok := err.(*store.ErrInternal); ok {
				// Fatal error
				return nil, err
			}
		} else {
			handlerExec.Env = environment.MergeEnvironments(os.Environ(), assets.Env(), handler.EnvVars, secrets)
		}
	}

	return l.Executor.Execute(ctx, handlerExec)
}

// socketHandler creates either a TCP or UDP client to write mutatedData
// to a socket. The provided handler Type determines the protocol.
func (l *LegacyAdapter) socketHandler(ctx context.Context, handler *corev2.Handler, event *corev2.Event, mutatedData []byte) (conn net.Conn, err error) {
	protocol := handler.Type
	host := handler.Socket.Host
	port := handler.Socket.Port
	timeout := handler.Timeout

	// Prepare log entry
	fields := utillogging.EventFields(event, false)
	fields["handler_name"] = handler.Name
	fields["handler_namespace"] = handler.Namespace
	fields["handler_protocol"] = protocol
	fields["pipeline"] = corev2.ContextPipeline(ctx)
	fields["pipeline_workflow"] = corev2.ContextPipelineWorkflow(ctx)

	// If Timeout is not specified, use the default.
	if timeout == 0 {
		timeout = DefaultSocketTimeout
	}

	address := fmt.Sprintf("%s:%d", host, port)
	timeoutDuration := time.Duration(timeout) * time.Second

	logger.WithFields(fields).Debug("sending event to socket handler")

	conn, err = net.DialTimeout(protocol, address, timeoutDuration)
	if err != nil {
		return nil, err
	}
	defer func() {
		e := conn.Close()
		if err == nil {
			err = e
		}
	}()

	bytes, err := conn.Write(mutatedData)
	if err != nil {
		logger.WithFields(fields).WithError(err).Error("failed to execute event handler")
		return nil, err
	}

	fields["bytes"] = bytes
	logger.WithFields(fields).Info("event socket handler executed")

	// TODO(jk): Why return the connection here if we never make use of it?
	// Perhaps we should return bytes or a result type?
	return conn, nil
}
