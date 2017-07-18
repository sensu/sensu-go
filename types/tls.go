package types

// TLSConfig for Etcd and backend listeners
type TLSConfig struct {
	CertFile              string
	KeyFile               string
	TrustedCAFile         string
	InsecureSkipTLSVerify bool
}
