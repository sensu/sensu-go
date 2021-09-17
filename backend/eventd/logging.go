package eventd

// Logger is the logging interface for eventd.
type Logger interface {
	Stop()
	Println(v interface{})
}

type NoopLogger struct {
}

func (NoopLogger) Stop() {}

func (NoopLogger) Println(interface{}) {}
