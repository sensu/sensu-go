package types

// A KeepaliveRecord is a tuple of an Entity ID and the time at which the
// entity's keepalive will next expire.
type KeepaliveRecord struct {
	EntityID string
	Time     int64
}
