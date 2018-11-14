package types

// DefaultKeepaliveTimeout specifies the default keepalive timeout
const DefaultKeepaliveTimeout = 120

// NewKeepaliveRecord initializes and returns a KeepaliveRecord from
// an entity and its expiration time.
func NewKeepaliveRecord(e *Entity, t int64) *KeepaliveRecord {
	return &KeepaliveRecord{
		ObjectMeta: ObjectMeta{
			Namespace: e.Namespace,
			Name:      e.Name,
		},
		Time: t,
	}
}
