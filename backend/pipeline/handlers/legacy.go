package legacy

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

type Handler struct {
	AssetGetter            asset.Getter
	Executor               command.Executor
	LicenseGetter          licensing.Getter
	SecretsProviderManager *secrets.ProviderManager
	Store                  store.Store
	StoreTimeout           time.Duration
}

func (h *Handler) CanHandle(ctx context.Context, ref *corev2.ResourceReference) bool {
	if ref.APIVersion == "core/v2" && ref.Type == "Handler" {
		return true
	}
	return false
}

func (h *Handler) Handle(ctx context.Context, ref *corev2.ResourceReference, event *corev2.Event) error {
	tctx, cancel := context.WithTimeout(ctx, h.StoreTimeout)
	handler, err := h.Store.GetHandlerByName(tctx, ref.Name)
	cancel()
	if err != nil {
		// TODO: handle this
	}

	switch handler.Type {
	case "pipe":
		if _, err := h.pipeHandler(handler, event, eventData); err != nil {
			logger.WithFields(fields).Error(err)
			if _, ok := err.(*store.ErrInternal); ok {
				return err
			}
		}
	case "tcp", "udp":
		if _, err := h.socketHandler(handler, event, eventData); err != nil {
			logger.WithFields(fields).Error(err)
			if _, ok := err.(*store.ErrInternal); ok {
				return err
			}
		}
	default:
		return errors.New("unknown handler type")
	}
	return nil
}

// pipeHandler fork/executes a child process for a Sensu pipe handler command
// and writes the mutated eventData to it via STDIN.
func (h *Handler) pipeHandler(ctx context.Context, handler *corev2.Handler, event *corev2.Event, eventData []byte) (*command.ExecutionResponse, error) {
	ctx = corev2.SetContextFromResource(ctx, handler)

	// Prepare log entry
	fields := utillogging.EventFields(event, false)
	fields["handler_name"] = handler.Name
	fields["handler_namespace"] = handler.Namespace

	if h.LicenseGetter != nil {
		if license := h.LicenseGetter.Get(); license != "" {
			handler.EnvVars = append(handler.EnvVars, fmt.Sprintf("SENSU_LICENSE_FILE=%s", license))
		}
	}

	secrets, err := h.SecretsProviderManager.SubSecrets(ctx, handler.Secrets)
	if err != nil {
		logger.WithFields(fields).WithError(err).Error("failed to retrieve secrets for handler")
		return nil, err
	}

	// Prepare environment variables
	env := environment.MergeEnvironments(os.Environ(), handler.EnvVars, secrets)

	handlerExec := command.ExecutionRequest{}
	handlerExec.Command = handler.Command
	handlerExec.Timeout = int(handler.Timeout)
	handlerExec.Env = env
	handlerExec.Input = string(eventData[:])

	// Only add assets to execution context if handler requires them
	if len(handler.RuntimeAssets) != 0 {
		logger.WithFields(fields).Debug("fetching assets for handler")
		// Fetch and install all assets required for handler execution
		matchedAssets := asset.GetAssets(ctx, h.Store, handler.RuntimeAssets)

		assets, err := asset.GetAll(ctx, h.AssetGetter, matchedAssets)
		if err != nil {
			logger.WithFields(fields).WithError(err).Error("failed to retrieve assets for handler")
			if _, ok := err.(*store.ErrInternal); ok {
				// Fatal error
				return nil, err
			}
		} else {
			handlerExec.Env = environment.MergeEnvironments(os.Environ(), assets.Env(), handler.EnvVars, secrets)
		}
	}

	result, err := h.Executor.Execute(context.Background(), handlerExec)

	if err != nil {
		logger.WithFields(fields).WithError(err).Error("failed to execute event pipe handler")
	} else {
		fields["status"] = result.Status
		fields["output"] = result.Output
		logger.WithFields(fields).Info("event pipe handler executed")
	}

	return result, err
}

// socketHandler creates either a TCP or UDP client to write eventData
// to a socket. The provided handler Type determines the protocol.
func (h *Handler) socketHandler(handler *corev2.Handler, event *corev2.Event, eventData []byte) (conn net.Conn, err error) {
	protocol := handler.Type
	host := handler.Socket.Host
	port := handler.Socket.Port
	timeout := handler.Timeout

	// Prepare log entry
	fields := utillogging.EventFields(event, false)
	fields["handler_name"] = handler.Name
	fields["handler_namespace"] = handler.Namespace
	fields["handler_protocol"] = protocol

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

	bytes, err := conn.Write(eventData)

	if err != nil {
		logger.WithFields(fields).WithError(err).Error("failed to execute event handler")
	} else {
		fields["bytes"] = bytes
		logger.WithFields(fields).Info("event socket handler executed")
	}

	return conn, nil
}
