package agent

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/sensu/sensu-go/agent/transformers"
	corev2 "github.com/sensu/core/v2"
	"github.com/sensu/sensu-go/asset"
	"github.com/sensu/sensu-go/command"
	"github.com/sensu/sensu-go/token"
	"github.com/sensu/sensu-go/transport"
	"github.com/sensu/sensu-go/util/environment"
	"github.com/sirupsen/logrus"
)

const (
	allowListOnDenyStatus        = "allow_list_on_deny_status"
	allowListOnDenyOutput        = "check command denied by the agent allow list"
	undocumentedTestCheckCommand = "!sensu_test_check!"

	measureMin        = "min"
	measureMax        = "max"
	measureNullStatus = "null-status"
)

var (
	errDupCheckRequest = errors.New("check request has already been received - agent and check may have multiple matching subscriptions")
	errOldCheckRequest = errors.New("check request is older than a previously received check request")
)

type checkExecutionError struct {
	error
	// Check Name
	Check string
}

// handleCheck is the check message handler.
// TODO(greg): At some point, we're going to need max parallelism.
func (a *Agent) handleCheck(ctx context.Context, payload []byte) error {
	request := &corev2.CheckRequest{}
	if err := a.unmarshal(payload, request); err != nil {
		return err
	} else if request == nil {
		return errors.New("given check configuration appears invalid")
	}

	checkConfig := request.Config

	// only schedule check execution if the issued timestamp is newer than any
	// previous executions of the check
	lastIssued := a.getLastIssued(request)
	if lastIssued > request.Issued {
		return checkExecutionError{
			error: errOldCheckRequest,
			Check: checkConfig.Name,
		}
	}
	if lastIssued == request.Issued {
		return checkExecutionError{
			error: errDupCheckRequest,
			Check: checkConfig.Name,
		}
	}

	// only schedule check execution if its not already in progress
	// ** check hooks are part of a checks execution
	if a.checkInProgress(request) {
		return checkExecutionError{
			error: fmt.Errorf("check execution still in progress: %s", checkConfig.Name),
			Check: checkConfig.Name,
		}
	}

	sendFailure := func(err error) {
		check := corev2.NewCheck(checkConfig)
		check.Executed = time.Now().Unix()
		event := &corev2.Event{
			ObjectMeta: corev2.NewObjectMeta("", check.Namespace),
			Check:      check,
		}
		a.sendFailure(event, err)
	}

	if a.config.DisableAssets && len(request.Assets) > 0 {
		err := errors.New("check requested assets, but they are disabled on this agent")
		sendFailure(err)
		return nil
	}

	// Validate that the given check is valid.
	if err := request.Config.Validate(); err != nil {
		sendFailure(fmt.Errorf("given check is invalid: %s", err))
		return nil
	}

	logger.Info("scheduling check execution: ", checkConfig.Name)

	entity := a.getAgentEntity()
	go a.executeCheck(ctx, request, entity)

	return nil
}

// handleCheckNoop is used to discard incoming check requests
func (a *Agent) handleCheckNoop(_ context.Context, _ []byte) error {
	return nil
}

func (a *Agent) checkInProgress(req *corev2.CheckRequest) bool {
	a.inProgressMu.Lock()
	defer a.inProgressMu.Unlock()
	_, ok := a.inProgress[checkKey(req)]
	return ok
}

func checkKey(request *corev2.CheckRequest) string {
	parts := []string{request.Config.Name}
	if len(request.Config.ProxyEntityName) > 0 {
		parts = append(parts, request.Config.ProxyEntityName)
	}
	return strings.Join(parts, "/")
}

func (a *Agent) addInProgress(request *corev2.CheckRequest) {
	a.inProgressMu.Lock()
	defer a.inProgressMu.Unlock()
	a.inProgress[checkKey(request)] = request.Config
}

func (a *Agent) removeInProgress(request *corev2.CheckRequest) {
	a.inProgressMu.Lock()
	defer a.inProgressMu.Unlock()
	delete(a.inProgress, checkKey(request))
}

func (a *Agent) getLastIssued(request *corev2.CheckRequest) int64 {
	a.lastIssuedMu.Lock()
	defer a.lastIssuedMu.Unlock()
	issued, ok := a.lastIssued[checkKey(request)]
	if !ok {
		return 0
	}
	return issued
}

func (a *Agent) setLastIssued(request *corev2.CheckRequest) {
	a.lastIssuedMu.Lock()
	defer a.lastIssuedMu.Unlock()
	a.lastIssued[checkKey(request)] = request.Issued
}

func (a *Agent) executeCheck(ctx context.Context, request *corev2.CheckRequest, entity *corev2.Entity) {
	a.setLastIssued(request)
	a.addInProgress(request)
	defer a.removeInProgress(request)

	checkAssets := request.Assets
	checkConfig := request.Config
	checkHooks := request.Hooks
	hookAssets := request.HookAssets
	secrets := request.Secrets

	// Before token subsitution we retain copy of the command
	origCommand := checkConfig.Command
	createEvent := func() *corev2.Event {
		event := &corev2.Event{}
		event.Namespace = checkConfig.Namespace
		event.Check = corev2.NewCheck(checkConfig)
		event.Check.Executed = time.Now().Unix()
		event.Check.Issued = request.Issued
		event.Pipelines = checkConfig.Pipelines

		// To guard against publishing sensitive/redacted client attribute values
		// the original command value is reinstated.
		event.Check.Command = origCommand

		event.Sequence = a.nextSequence(checkConfig.Name)

		return event
	}

	if origCommand != undocumentedTestCheckCommand {
		// Perform token substitution on the check configuration, but only if
		// we aren't doing load testing with the undocumented test check
		// command.
		if err := token.SubstituteCheck(checkConfig, entity); err != nil {
			a.sendFailure(createEvent(), fmt.Errorf("error while substituting check tokens: %s", err))
			return
		}
	}

	// Instantiate event
	event := createEvent()
	check := event.Check
	event.Entity = a.getAgentEntity()

	// Prepare log entry
	fields := logrus.Fields{
		"namespace": check.Namespace,
		"check":     check.Name,
		"assets":    check.RuntimeAssets,
	}

	// Match check against allow list
	var matchedEntry allowList
	var match bool
	if len(a.allowList) != 0 {
		logger.WithFields(fields).Debug("matching check against agent allow list")
		matchedEntry, match = a.matchAllowList(checkConfig.Command)
		if !match {
			logger.WithFields(fields).Debug("check does not match agent allow list")
			a.sendFailure(event, fmt.Errorf(allowListOnDenyOutput))
			return
		}
		logger.WithFields(fields).Debug("check matches agent allow list")
	}

	// Fetch and install all assets required for check execution.
	var assets = asset.RuntimeAssetSet{}
	if len(checkAssets) == 0 {
		logger.WithFields(fields).Debug("no assets defined for this check")
	} else {
		logger.WithFields(fields).Debug("fetching assets for check")
		var err error
		assets, err = asset.GetAll(ctx, a.assetGetter, checkAssets)
		if err != nil {
			a.sendFailure(event, fmt.Errorf("error getting assets for check: %s", err))
			return
		}
	}

	// Prepare environment variables
	var env []string
	if match && !matchedEntry.EnableEnv {
		logger.WithFields(fields).Debug("disabling check env vars per the agent allow list")
		env = environment.MergeEnvironments(os.Environ(), assets.Env(), secrets)
	} else {
		env = environment.MergeEnvironments(os.Environ(), assets.Env(), secrets, checkConfig.EnvVars)
	}

	// Verify sha against the allow list
	if matchedEntry.Sha512 != "" {
		logger.WithFields(fields).Debug("matching check sha against agent allow list")
		path, err := lookPath(strings.Split(checkConfig.Command, " ")[0], env)
		if err != nil {
			logger.WithFields(fields).WithError(err).Error("unable to find the executable path")
			a.sendFailure(event, fmt.Errorf(allowListOnDenyOutput))
			return
		}
		file, err := os.Open(path)
		if err != nil {
			logger.WithFields(fields).WithError(err).Error("unable to open executable")
			a.sendFailure(event, fmt.Errorf(allowListOnDenyOutput))
			return
		}
		verifier := asset.Sha512Verifier{}
		if err := verifier.Verify(file, matchedEntry.Sha512); err != nil {
			logger.WithFields(fields).WithError(err).Error("check sha does not match agent allow list")
			a.sendFailure(event, fmt.Errorf(allowListOnDenyOutput))
			return
		}
	}

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
		checkExec.Status = 3
	} else {
		event.Check.Output = checkExec.Output
	}

	event.Check.Duration = checkExec.Duration
	event.Check.Status = uint32(checkExec.Status)
	event.Check.ProcessedBy = a.config.AgentName

	event.Timestamp = time.Now().Unix()
	id, err := uuid.NewRandom()
	if err == nil {
		event.ID = id[:]
	}

	// Instantiate metrics in the event if the check is attempting to extract metrics
	if check.OutputMetricFormat != "" || len(check.OutputMetricHandlers) != 0 {
		event.Metrics = &corev2.Metrics{}
	}

	if check.OutputMetricFormat != "" {
		event.Metrics.Points = extractMetrics(event)

		if event.Check.Status == 0 && len(event.Metrics.Points) > 0 && len(check.OutputMetricThresholds) > 0 {
			event.Check.Status = evaluateOutputMetricThresholds(event)
		}
	}

	if len(check.OutputMetricHandlers) != 0 {
		event.Metrics.Handlers = check.OutputMetricHandlers
	}

	// Execute hooks after we have a completely populated event object
	if len(checkHooks) != 0 {
		event.Check.Hooks = a.ExecuteHooks(ctx, request, event, hookAssets)
	}

	// The check requested that we discard its output before writing back
	// the result.
	if event.Check.DiscardOutput {
		event.Check.Output = ""
	}

	msg, err := a.marshal(event)
	if err != nil {
		logger.WithError(err).Error("error marshaling check result")
		return
	}

	tm := &transport.Message{
		Type:    transport.MessageTypeEvent,
		Payload: msg,
	}

	logEvent(event)

	a.sendMessage(tm)
}

func (a *Agent) sendFailure(event *corev2.Event, err error) {
	logger.WithFields(logrus.Fields{
		"event": event,
	}).Error(err)

	event.Check.Output = err.Error()
	event.Check.Status = 3
	event.Entity = a.getAgentEntity()
	event.Timestamp = time.Now().Unix()

	// Override the default check status of 3 if an annotation is configured
	allowListStatus, ok := event.Check.Annotations[allowListOnDenyStatus]
	if ok {
		allowListValue, err := strconv.ParseUint(allowListStatus, 10, 32)
		if err == nil {
			event.Check.Status = uint32(allowListValue)
		}
	}

	if msg, err := a.marshal(event); err != nil {
		logger.WithError(err).Error("error marshaling check failure")
	} else {
		tm := &transport.Message{
			Type:    transport.MessageTypeEvent,
			Payload: msg,
		}
		a.sendMessage(tm)
	}
}

func extractMetrics(event *corev2.Event) []*corev2.MetricPoint {
	var transformer Transformer
	if !event.HasCheck() {
		logger.WithError(transformers.ErrMetricExtraction).Error("event must contain a check to parse and extract metrics")
		return nil
	}

	switch event.Check.OutputMetricFormat {
	case corev2.GraphiteOutputMetricFormat:
		transformer = transformers.ParseGraphite(event)
	case corev2.InfluxDBOutputMetricFormat:
		transformer = transformers.ParseInflux(event)
	case corev2.NagiosOutputMetricFormat:
		transformer = transformers.ParseNagios(event)
	case corev2.OpenTSDBOutputMetricFormat:
		transformer = transformers.ParseOpenTSDB(event)
	case corev2.PrometheusOutputMetricFormat:
		transformer = transformers.ParseProm(event)
	}

	if transformer == nil {
		logger.WithField("format", event.Check.OutputMetricFormat).WithError(transformers.ErrMetricExtraction).Error("output metric format is not supported")
		return nil
	}

	return transformer.Transform()
}

func evaluateOutputMetricThresholds(event *corev2.Event) uint32 {
	if event.Check.Status > 0 {
		return event.Check.Status
	}

	points := event.Metrics.Points
	thresholds := event.Check.OutputMetricThresholds

	var status uint32 = 0
	annotationValue := ""
	for _, thresholdRule := range thresholds {
		ruleMatched := false
		for _, metricPoint := range points {
			if thresholdRule.MatchesMetricPoint(metricPoint) {
				ruleMatched = true
				for _, rule := range thresholdRule.Thresholds {
					if rule.Min != "" {
						min, err := strconv.ParseFloat(rule.Min, 64)
						if err != nil {
							continue
						}
						if metricPoint.Value < min {
							addThresholdAnnotation(event, thresholdRule, measureMin, rule.Status, metricPoint.Value, rule.Min, true)
							if status < rule.Status {
								status = rule.Status
								annotationValue = getAnnotationValue(thresholdRule, measureMin, metricPoint.Value, rule.Min, true)
							}
							continue
						} else {
							addThresholdAnnotation(event, thresholdRule, measureMin, 0, metricPoint.Value, rule.Min, false)
							annotationValue = getAnnotationValue(thresholdRule, measureMin, metricPoint.Value, rule.Min, false)
						}
					}
					if rule.Max != "" {
						max, err := strconv.ParseFloat(rule.Max, 64)
						if err != nil {
							continue
						}
						if metricPoint.Value > max {
							addThresholdAnnotation(event, thresholdRule, measureMax, rule.Status, metricPoint.Value, rule.Max, true)
							if status < rule.Status {
								status = rule.Status
								annotationValue = getAnnotationValue(thresholdRule, measureMax, metricPoint.Value, rule.Max, true)
							}
						} else {
							addThresholdAnnotation(event, thresholdRule, measureMax, 0, metricPoint.Value, rule.Max, false)
							annotationValue = getAnnotationValue(thresholdRule, measureMax, metricPoint.Value, rule.Max, false)
						}
					}
				}
			}
		}
		if !ruleMatched {
			if thresholdRule.NullStatus > 0 {
				addNullStatusThresholdAnnotation(event, thresholdRule, thresholdRule.NullStatus)
				if status < thresholdRule.NullStatus {
					status = thresholdRule.NullStatus
					annotationValue = getNullStatusAnnotationValue(thresholdRule)
				}
			}
		}
	}

	if annotationValue != "" {
		event.AddAnnotation("sensu.io/notifications/"+corev2.CheckStatusToCaption(status), annotationValue)
	}

	return status
}

func addThresholdAnnotation(event *corev2.Event, metricThreshold *corev2.MetricThreshold, measure string, status uint32, value float64, threshold string, isExceeded bool) {
	event.AddAnnotation(getAnnotationKey(metricThreshold, measure, status), getAnnotationValue(metricThreshold, measure, value, threshold, isExceeded))
}

func addNullStatusThresholdAnnotation(event *corev2.Event, metricThreshold *corev2.MetricThreshold, status uint32) {
	event.AddAnnotation(getAnnotationKey(metricThreshold, measureNullStatus, status), getNullStatusAnnotationValue(metricThreshold))
}

func getAnnotationKey(metricThreshold *corev2.MetricThreshold, measure string, status uint32) string {
	var key strings.Builder

	key.WriteString("sensu.io/output_metric_thresholds/")
	key.WriteString(metricThreshold.Name)
	for _, tag := range metricThreshold.Tags {
		key.WriteString(".")
		key.WriteString(tag.Value)
	}
	key.WriteString("/")
	key.WriteString(measure)
	key.WriteString("/")
	key.WriteString(corev2.CheckStatusToCaption(status))

	return key.String()
}

func getAnnotationValue(metricThreshold *corev2.MetricThreshold, measure string, value float64, threshold string, isExceeded bool) string {
	var val strings.Builder
	var tagsKeyVal strings.Builder

	for tagIdx, tag := range metricThreshold.Tags {
		if tagIdx > 0 {
			tagsKeyVal.WriteString(",")
		}
		tagsKeyVal.WriteString(tag.Name)
		tagsKeyVal.WriteString("=")
		tagsKeyVal.WriteString(tag.Value)
	}

	val.WriteString("The value of ")
	val.WriteString(metricThreshold.Name)
	if tagsKeyVal.Len() > 0 {
		val.WriteString(" (")
		val.WriteString(tagsKeyVal.String())
		val.WriteString(")")
	}
	if isExceeded {
		val.WriteString(" exceeded the configured threshold (")
	} else {
		val.WriteString(" is within the configured threshold (")
	}
	val.WriteString(measure)
	val.WriteString(": ")
	val.WriteString(threshold)
	val.WriteString(", actual: ")
	val.WriteString(strconv.FormatFloat(value, 'f', -1, 64))
	val.WriteString(").")

	return val.String()
}

func getNullStatusAnnotationValue(metricThreshold *corev2.MetricThreshold) string {
	var val strings.Builder
	var tagsKeyVal strings.Builder

	for tagIdx, tag := range metricThreshold.Tags {
		if tagIdx > 0 {
			tagsKeyVal.WriteString(", ")
		}
		tagsKeyVal.WriteString(tag.Name)
		tagsKeyVal.WriteString("=\"")
		tagsKeyVal.WriteString(tag.Value)
		tagsKeyVal.WriteString("\"")
	}

	val.WriteString(strings.ToUpper(corev2.CheckStatusToCaption(metricThreshold.NullStatus)))
	val.WriteString(" : no metric matching \"")
	val.WriteString(metricThreshold.Name)
	val.WriteString("\"")
	if tagsKeyVal.Len() > 0 {
		val.WriteString(" (")
		val.WriteString(tagsKeyVal.String())
		val.WriteString(")")
	}
	val.WriteString(" was found")

	for _, t := range metricThreshold.Thresholds {
		hasMin := len(t.Min) > 0
		hasMax := len(t.Max) > 0
		val.WriteString("; expected ")
		if hasMin {
			val.WriteString("min: ")
			val.WriteString(t.Min)
		}
		if hasMin && hasMax {
			val.WriteString(" - ")
		}
		if hasMax {
			val.WriteString("max: ")
			val.WriteString(t.Max)
		}
		val.WriteString(" (status: ")
		val.WriteString(corev2.CheckStatusToCaption(t.Status))
		val.WriteString(")")
	}

	return val.String()
}
