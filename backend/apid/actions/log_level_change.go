package actions

import (
	"context"
	"errors"
	"fmt"

	"github.com/sensu/sensu-go/util/logging"
)

// LogLevelChangeController exposes actions to change log levels dynamically.
type LogLevelChangeController struct{}

// NewLogLevelChangeController returns a new LogLevelChangeController
func NewLogLevelChangeController() LogLevelChangeController {
	return LogLevelChangeController{}
}

// SetGlobalModuleLogLevel exposes action to change log level for all modules
func (a LogLevelChangeController) SetGlobalModuleLogLevel(ctx context.Context, logLevel string) ([]*logging.LogHistory, error) {
	// Get current levels atomically
	existingLog := logging.GetModuleLogLevels()
	var logChanges []*logging.LogHistory
	var errLog []error

	logger.Debug("Setting global log level to: ", logLevel)

	for _, logLevelArr := range existingLog {
		// Actually change the log level
		err := logging.SetLogLevel(logLevelArr.Module, logLevel)
		if err != nil {
			errLog = append(errLog, fmt.Errorf("failed to set level for module %s: %w", logLevelArr.Module, err))
			continue
		}

		if logLevelArr.Module == "etcd" {
			if err := logging.SetEtcdLogLevel(logLevel); err != nil {
				errLog = append(errLog, fmt.Errorf("failed to set level for module %s: %w", logLevelArr.Module, err))
				continue
			}
		}

		change := &logging.LogHistory{
			Module:   logLevelArr.Module,
			OldLevel: logLevelArr.Level,
			NewLevel: logLevel,
		}
		logChanges = append(logChanges, change)
	}

	// Return error if any module failed
	if len(errLog) > 0 {
		return logChanges, fmt.Errorf("some modules failed to update: %v", errLog)
	}

	return logChanges, nil
}

// Create set log level for multiple modules.
func (a LogLevelChangeController) Create(ctx context.Context, requests []logging.LogLevelRequest) ([]logging.LogHistory, error) {
	var results []logging.LogHistory
	for _, req := range requests {
		oldLevel := logging.GetModuleLogLevel(req.Module)
		err := logging.SetLogLevel(req.Module, req.Level)
		change := logging.LogHistory{
			Module: req.Module,
		}

		if err != nil {
			change.Error = err.Error()
			results = append(results, change)
		} else {

			// if Module is etcd, initialize zap logger
			if req.Module == "etcd" {
				if err := logging.SetEtcdLogLevel(req.Level); err != nil {
					change.Error = err.Error()
				}
			}

			change.NewLevel = req.Level
			change.OldLevel = oldLevel
			results = append(results, change)
		}
	}
	return results, nil
}

// List will give Module and corresponding log level
func (a LogLevelChangeController) List(ctx context.Context) ([]logging.LogLevelRequest, error) {
	existingLog := logging.GetModuleLogLevels()
	if len(existingLog) == 0 {
		return []logging.LogLevelRequest{}, errors.New("no module log levels found")
	}
	return existingLog, nil
}

func (a LogLevelChangeController) ModuleLogLevel(ctx context.Context, module string) (*logging.LogLevelRequest, error) {
	existingLog := logging.GetModuleLogLevel(module)
	if existingLog == "" {
		return nil, errors.New("no module log levels found")
	}
	logLevel := &logging.LogLevelRequest{
		Module: module,
		Level:  existingLog,
	}
	return logLevel, nil
}
