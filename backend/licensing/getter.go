package licensing

// Getter represents an abstracted license getter.
type Getter interface {
	// Get gets the installed license file of the Sensu cluster.
	Get() string
}
