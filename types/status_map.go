package types

// StatusMap is a map of backend component names to their current status info.
type StatusMap map[string]bool

// Healthy returns true if the StatsMap shows all healthy indicators; false
// otherwise.
func (s StatusMap) Healthy() bool {
	for _, v := range s {
		if !v {
			return false
		}
	}
	return true
}
