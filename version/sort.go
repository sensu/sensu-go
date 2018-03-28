package version

import "github.com/coreos/go-semver/semver"

type byVersion []string

func (v byVersion) Len() int {
	return len(v)
}

func (v byVersion) Swap(i, j int) {
	v[i], v[j] = v[j], v[i]
}

func (v byVersion) Less(i, j int) bool {
	vI := semver.New(v[i])
	vJ := semver.New(v[j])
	return !vI.LessThan(*vJ)
}
