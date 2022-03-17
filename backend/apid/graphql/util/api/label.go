package util_api

// KVPairString pair of string values
type KVPairString struct {
	Key string
	Val string
}

// MakeKVPairString produces a dictionary from a given map[string]string;
// required as GraphQL doesn't have an explicit map type
func MakeKVPairString(m map[string]string) []KVPairString {
	pairs := make([]KVPairString, 0, len(m))
	for key, val := range m {
		pair := KVPairString{Key: key, Val: val}
		pairs = append(pairs, pair)
	}
	return pairs
}
