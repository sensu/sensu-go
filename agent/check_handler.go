package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"time"

	"github.com/sensu/sensu-go/agent/transformers"
	"github.com/sensu/sensu-go/api/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/types/dynamic"
	"github.com/sensu/sensu-go/util/environment"
	"github.com/sirupsen/logrus"
)

// TODO(greg): At some point, we're going to need max parallelism.
func (a *Agent) handleCheck(payload []byte) error {
	request := &v2.CheckRequest{}
	if err := json.Unmarshal(payload, request); err != nil {
		return err
	} else if request == nil {
		return errors.New("given check configuration appears invalid")
	}

	// only schedule check execution if its not already in progress
	// ** check hooks are part of a checks execution
	a.inProgressMu.Lock()
	_, in := a.inProgress[checkKey(request)]
	a.inProgressMu.Unlock()
	if !in {
		logger.Info("scheduling check execution: ", request.Config.Name)

		if ok := a.prepareCheck(request.Config); !ok {
			// An error occured during the preparation of the check and the error has
			// been sent back to the server. At this point we should not execute the
			// check and wait for the next check request
			return nil
		}

		go a.executeCheck(request)
	} else {
		return fmt.Errorf("check execution still in progress: %s", request.Config.Name)
	}

	return nil
}

func checkKey(request *v2.CheckRequest) string {
	return request.Config.Name
}

func (a *Agent) executeCheck(request *v2.CheckRequest) {
	a.inProgressMu.Lock()
	a.inProgress[checkKey(request)] = request.Config
	a.inProgressMu.Unlock()
	defer func() {
		a.inProgressMu.Lock()
		delete(a.inProgress, request.Config.Name)
		a.inProgressMu.Unlock()
	}()

	checkConfig := request.Config
	checkAssets := request.Assets
	checkHooks := request.Hooks

	// Instantiate Event
	check := v2.NewCheck(checkConfig)
	check.Executed = time.Now().Unix()
	check.Issued = request.Issued
	event := &v2.Event{
		ObjectMeta: v2.ObjectMeta{
			Namespace: check.Namespace,
		},
		Check: check,
	}

	// Prepare log entry
	fields := logrus.Fields{
		"namespace": check.Namespace,
		"check":     check.Name,
		"assets":    check.RuntimeAssets,
	}

	// Fetch and install all assets required for check execution.
	logger.WithFields(fields).Debug("fetching assets for check")
	assets, err := asset.GetAll(a.assetGetter, checkAssets)
	if err != nil {
		a.sendFailure(event, fmt.Errorf("error getting assets for event: %s", err))
		return
	}

	// Prepare environment variables
	env := environment.MergeEnvironments(os.Environ(), assets.Env(), check.EnvVars)

	// Inject the dependencies into PATH, LD_LIBRARY_PATH & CPATH so that they
	// are availabe when when the command is executed.
	ex := command.ExecutionRequest{
		Env:          env,
		Command:      checkConfig.Command,
		Timeout:      int(checkConfig.Timeout),
		InProgress:   a.inProgress,
		InProgressMu: a.inProgressMu,
		Name:         checkConfig.Name,
	}

	// If stdin is true, add JSON event data to command execution.
	if checkConfig.Stdin {
		input, err := json.Marshal(event)
		if err != nil {
			a.sendFailure(event, fmt.Errorf("error marshaling json from event: %s", err))
			return
		}
		ex.Input = string(input)
	}

	checkExec, err := a.executor.Execute(context.Background(), ex)
	if err != nil {
		event.Check.Output = err.Error()
	} else {
		event.Check.Output = checkExec.Output
	}

	event.Check.Duration = checkExec.Duration
	event.Check.Status = uint32(checkExec.Status)

	event.Entity = a.getAgentEntity()
	event.Timestamp = time.Now().Unix()

	if len(checkHooks) != 0 {
		event.Check.Hooks = a.ExecuteHooks(request, checkExec.Status)
	}

	// Instantiate metrics in the event if the check is attempting to extract metrics
	if check.OutputMetricFormat != "" || len(check.OutputMetricHandlers) != 0 {
		event.Metrics = &v2.Metrics{}
	}

	if check.OutputMetricFormat != "" {
		event.Metrics.Points = extractMetrics(event)
	}

	if len(check.OutputMetricHandlers) != 0 {
		event.Metrics.Handlers = check.OutputMetricHandlers
	}

	msg, err := json.Marshal(event)
	if err != nil {
		logger.WithError(err).Error("error marshaling check result")
		return
	}

	a.sendMessage(transport.MessageTypeEvent, msg)
}

// prepareCheck prepares a check before its execution by validating the
// configuration and performing token substitution. A boolean value is returned,
// indicathing whether the check should be executed or not
func (a *Agent) prepareCheck(cfg *v2.CheckConfig) bool {
	// Instantiate an event in case of failure
	check := v2.NewCheck(cfg)
	check.Executed = time.Now().Unix()
	event := &v2.Event{
		ObjectMeta: v2.NewObjectMeta("", check.Namespace),
		Check:      check,
	}

	// Validate that the given check is valid.
	if err := check.Validate(); err != nil {
		a.sendFailure(event, fmt.Errorf("given check is invalid: %s", err))
		return false
	}

	// Extract the extended attributes from the entity and combine them at the
	// top-level so they can be easily accessed using token substitution
	synthesizedEntity := dynamic.Synthesize(a.getAgentEntity())

	// Substitute tokens within the check configuration with the synthesized
	// entity
	checkBytes, err := TokenSubstitution(synthesizedEntity, cfg)
	if err != nil {
		a.sendFailure(event, err)
		return false
	}

	// Unmarshal the check configuration obtained after the token substitution
	// back into the check config struct
	err = json.Unmarshal(checkBytes, cfg)
	if err != nil {
		a.sendFailure(event, fmt.Errorf("could not unmarshal the check: %s", err))
		return false
	}

	return true
}

func (a *Agent) sendFailure(event *v2.Event, err error) {
	event.Check.Output = err.Error()
	event.Check.Status = 3
	event.Entity = a.getAgentEntity()
	event.Timestamp = time.Now().Unix()

	if msg, err := json.Marshal(event); err != nil {
		logger.WithError(err).Error("error marshaling check failure")
	} else {
		a.sendMessage(transport.MessageTypeEvent, msg)
	}
}

func extractMetrics(event *v2.Event) []*v2.MetricPoint {
	var transformer Transformer
	if !event.HasCheck() {
		logger.WithError(transformers.ErrMetricExtraction).Error("event must contain a check to parse and extract metrics")
		return nil
	}

	switch event.Check.OutputMetricFormat {
	case v2.GraphiteOutputMetricFormat:
		transformer = transformers.ParseGraphite(event)
	case v2.InfluxDBOutputMetricFormat:
		transformer = transformers.ParseInflux(event)
	case v2.NagiosOutputMetricFormat:
		transformer = transformers.ParseNagios(event)
	case v2.OpenTSDBOutputMetricFormat:
		transformer = transformers.ParseOpenTSDB(event)
	}

	if transformer == nil {
		logger.WithField("format", event.Check.OutputMetricFormat).WithError(transformers.ErrMetricExtraction).Error("output metric format is not supported")
		return nil
	}

	return transformer.Transform()
}
