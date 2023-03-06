package licensing

// Getter represents an abstracted license getter.
type Getter interface {
	// Get gets the installed license file of the Sensu cluster.
	Get() string
}

// DummyGetter is a Getter that simply returns an empty license. It can be used
// in lieu of an actual Getter implementation when we don't have access to one.
type DummyGetter struct{}

func (g *DummyGetter) Get() string {
	return ""
}
