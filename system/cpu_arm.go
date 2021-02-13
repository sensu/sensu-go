// +build arm

package system

//go:linkname goarm runtime.goarm
var goarm int32

func getARMVersion() int32 {
	return goarm
}
