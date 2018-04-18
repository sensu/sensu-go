// Package extension provides types and functions for creating Sensu extensions
// in Go. The extensions run as servers and use gRPC as a transport.
//
// Users only need provide a net.Listener, and one or more of the extension
// methods, to have a fully working Sensu extension. See example for details.
package extension
