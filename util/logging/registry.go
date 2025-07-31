package logging

import (
	"errors"
	corev2 "github.com/sensu/core/v2"
	"github.com/sirupsen/logrus"
	"go.etcd.io/etcd/client/pkg/v3/logutil"
	"go.uber.org/zap/zapcore"
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
	OldLevel string `json:"old_level,omitempty"`
	NewLevel string `json:"new_level,omitempty"`
	Error    string `json:"error,omitempty"`
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
func GetModuleLogLevels() []LogLevelRequest {
	var logLevels []LogLevelRequest
	registryMu.RLock()
	for moduleName, logger := range loggerRegistry {
		logLevel := LogLevelRequest{
			Module: moduleName,
			Level:  logger.GetLevel().String(),
		}
		logLevels = append(logLevels, logLevel)
	}
	registryMu.RUnlock()
	return logLevels
}

// GetModuleLogLevel returns a loggers log level
func GetModuleLogLevel(module string) string {
	registryMu.RLock()
	logger, ok := loggerRegistry[module]
	registryMu.RUnlock()
	if ok {
		return logger.Level.String()
	}
	registryMu.Lock()
	defer registryMu.Unlock()
	logger, ok = loggerRegistry[module]
	if ok {
		return logger.Level.String()
	}

	return ""
}

// SetLogLevel sets the log level for the given module's logger.
func SetLogLevel(module, level string) error {
	registryMu.Lock()
	defer registryMu.Unlock()
	logger, ok := loggerRegistry[module]
	if !ok {
		return errors.New("no module found")
	}
	lvl, err := logrus.ParseLevel(level)
	if err != nil {
		return err
	}
	logger.SetLevel(lvl)
	return nil
}

func (l *LogLevelRequest) GetObjectMeta() corev2.ObjectMeta {
	return corev2.ObjectMeta{Name: l.Module}
}

// StorePrefix returns the path prefix to this resource in the store
func (l *LogLevelRequest) StorePrefix() string {
	return ""
}

// SetNamespace sets the namespace of the resource.
func (l *LogLevelRequest) SetNamespace(namespace string) {
	// no-op
}

// SetObjectMeta sets the meta of the resource.
func (l *LogLevelRequest) SetObjectMeta(meta corev2.ObjectMeta) {
	// no-op
}

func (*LogLevelRequest) RBACName() string {
	return "loglevel"
}

func (l *LogLevelRequest) URIPath() string {
	return ""
}

func (l *LogLevelRequest) Validate() error {
	return nil
}
