package eventd

// Logger ...
type Logger interface {
	Stop()
	Println(v interface{})
}

// RawLogger ...
type RawLogger struct{}

// Println ...
func (l *RawLogger) Println(v interface{}) {}

// Stop ...
func (l *RawLogger) Stop() {}
