//go:build arm
// +build arm

package system

import _ "unsafe"

//go:linkname goarm runtime.goarm
var goarm uint8

func getARMVersion() int32 {
	return int32(goarm)
}
