package logging

import (
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/client/pkg/v3/logutil"
	"go.uber.org/zap/zapcore"
	"os"
	"sync"
)

const flagLogLevel = "log-level"

var (
	loggerRegistry = make(map[string]*logrus.Logger)
	registryMu     sync.RWMutex
)

type LogLevelRequest struct {
	Module string `json:"module"`
	Level  string `json:"level"`
}

type LogHistory struct {
	Module   string `json:"module"`
	OldLevel string `json:"old_level"`
	NewLevel string `json:"new_level"`
}

type LogLevel struct {
	Module   string `json:"module"`
	LogLevel string `json:"log_level"`
}

var defaultLogLevel = logrus.InfoLevel

func InitLogLevel(level string) {
	lvl, err := logrus.ParseLevel(level)
	if err == nil {
		defaultLogLevel = lvl
	}
}

// GetLogger returns a logger for the given module, creating one if it doesn't exist.
func GetLogger(module string) *logrus.Logger {
	registryMu.RLock()
	logger, ok := loggerRegistry[module]
	registryMu.RUnlock()
	if ok {
		return logger
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	logger, ok = loggerRegistry[module]
	if ok {
		return logger
	}
	logger = logrus.New()
	logger.Formatter = &logrus.JSONFormatter{}
	logger.SetLevel(defaultLogLevel)

	loggerRegistry[module] = logger
	return logger
}

// SetAllLoggersLevel sets All logger to a given loglevel
func SetAllLoggersLevel(level string) {
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	for _, logger := range loggerRegistry {
		logger.SetLevel(lvl)
	}
}

// SetEtcdLogLevel dynamically sets the log level for embedded etcd
func SetEtcdLogLevel(level string) error {
	lvl, err := zapcore.ParseLevel(level)
	if err != nil {
		return err
	}
	logutil.DefaultZapLoggerConfig.Level.SetLevel(lvl)
	return nil
}

// GetModuleLogLevels returns current log levels for all modules
func GetModuleLogLevels() map[string]string {
	registryMu.RLock()
	levels := make(map[string]string)
	for moduleName, logger := range loggerRegistry {
		levels[moduleName] = logger.GetLevel().String()
	}
	registryMu.RUnlock()
	return levels
}

// SetLogLevel sets the log level for the given module's logger.
func SetLogLevel(module, level string) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	logger, ok := loggerRegistry[module]
	if !ok {
		logger = logrus.New()
		logger.Out = os.Stdout
		logger.Formatter = &logrus.TextFormatter{}
		logger.SetLevel(logrus.InfoLevel)
		loggerRegistry[module] = logger
	}
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logger.SetLevel(lvl)
	return nil
}
