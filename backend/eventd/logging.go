package eventd

// Logger ...
type Logger interface {
	Write(v interface{})
}

// RawLogger ...
type RawLogger struct{}

// Write ...
func (l *RawLogger) Write(v interface{}) { return }
