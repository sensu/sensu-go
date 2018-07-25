package types

type ClusterHealth struct {
	MemberID uint64
	Name     string
	Err      error
	Healthy  bool
}
