// Package leader provides functions that will only be executed if the Go
// runtime process is the sensu leader.
//
// Since leadership is global, the package relies on package-level state that
// is initialized through the Initialize function. The package is therefore a
// singleton.
//
// If Override is called, then the other functions in this package will proceed
// assuming that this node is the leader.
//
// If neither Override or Initialize are called, then the functions in this
// package will return ErrNotInitialized.
//
// The exported functions in this package are goroutine-safe.
package leader
