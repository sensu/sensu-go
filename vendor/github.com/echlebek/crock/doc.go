// Package crock provides mock implementations of functions from the stdlib
// time package.
//
// crock deliberately does not specify any interfaces or shims for "real" time.
// Consumers should create their own interfaces, and make a trivial shim to
// integrate functions from package time into their code.
//
// See the example for some stuff you can lift for this purpose.
package crock
