package eventd

// Logger ...
type Logger interface {
	Stop()
	Write(v interface{})
}

// RawLogger ...
type RawLogger struct{}

// Stop ...
func (l *RawLogger) Stop() { return }

// Write ...
func (l *RawLogger) Write(v interface{}) { return }
