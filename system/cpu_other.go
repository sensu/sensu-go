//go:build !arm
// +build !arm

package system

func getARMVersion() int32 {
	return 0
}
