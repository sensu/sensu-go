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

	for moduleName, oldLevel := range existingLog {
		// Actually change the log level
		err := logging.SetLogLevel(moduleName, logLevel)
		if err != nil {
			errLog = append(errLog, fmt.Errorf("failed to set level for module %s: %w", moduleName, err))
			continue
		}

		if moduleName == "etcd" {
			if err := logging.SetEtcdLogLevel(logLevel); err != nil {
				errLog = append(errLog, fmt.Errorf("failed to set level for module %s: %w", moduleName, err))
				continue
			}
		}

		change := &logging.LogHistory{
			Module:   moduleName,
			OldLevel: oldLevel,
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

// Create changes the log level for a given module.
func (a LogLevelChangeController) Create(ctx context.Context, level string, module string) (*logging.LogHistory, error) {
	// get old logger information & build object
	oldLogger := logging.GetLogger(module)
	change := &logging.LogHistory{
		Module:   module,
		OldLevel: oldLogger.Level.String(),
		NewLevel: level,
	}
	err := logging.SetLogLevel(module, level)
	if err != nil {
		return &logging.LogHistory{}, err
	}

	// if Module is etcd, initialize zap logger
	if module == "etcd" {
		if err := logging.SetEtcdLogLevel(level); err != nil {
			return &logging.LogHistory{}, err
		}
		logger.Infof("Set embedded etcd log level to %s", level)
		return change, nil
	}
	logger.Infof("Log level for [module] %s changed to %s ", module, level)

	return change, nil
}

// List will give Module and corresponding log level
func (a LogLevelChangeController) List(ctx context.Context) (map[string]string, error) {
	existingLog := logging.GetModuleLogLevels()
	logger.Debug("Module wise Log level: ")
	for moduleName, level := range existingLog {
		logger.Debugf("[Module]:%s [Level]:%s", moduleName, level)
	}
	if len(existingLog) == 0 {
		return map[string]string{}, errors.New("no module log levels found")
	}
	return existingLog, nil
}

func (a LogLevelChangeController) ModuleLogLevel(ctx context.Context, module string) (*logging.LogLevel, error) {
	existingLog := logging.GetLogger(module)
	logLevel := &logging.LogLevel{
		Module:   module,
		LogLevel: existingLog.Level.String(),
	}
	return logLevel, nil
}
