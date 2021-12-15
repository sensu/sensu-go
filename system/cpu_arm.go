//go:build arm
// +build arm

package system

import _ "unsafe"

//go:linkname goarm runtime.goarm
var goarm int32

func getARMVersion() int32 {
	return goarm
}
