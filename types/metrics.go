package types

// A Metrics is an event metrics payload specification (TBD).
type Metrics struct{}

// Validate returns an error if metrics does not pass validation tests.
func (m *Metrics) Validate() error {
	return nil
}
