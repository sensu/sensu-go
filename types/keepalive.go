package types

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
