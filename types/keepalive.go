package types

// A KeepaliveRecord is a tuple of an Entity ID and the time at which the
// entity's keepalive will next expire.
type KeepaliveRecord struct {
	Environment  string
	Organization string
	EntityID     string
	Time         int64
}

// NewKeepaliveRecord initializes and returns a KeepaliveRecord from
// an entity and its expiration time.
func NewKeepaliveRecord(e *Entity, t int64) *KeepaliveRecord {
	return &KeepaliveRecord{
		Environment:  e.Environment,
		Organization: e.Organization,
		EntityID:     e.ID,
		Time:         t,
	}
}
