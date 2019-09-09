/*
Package api provides API clients for interacting with Sensu's core datatypes,
with full support for RBAC authorization. The clients in this package can be
provided to users who wish to interact with Sensu programmatically, without
exposing the system to the risk of unauthorized access.

Usage

Most users will not construct these clients themselves, but will be given them
on behalf of a larger framework. For instance, filters, mutators, and handlers
written in Javascript could use these clients, expecting to be provided with a
context that contains their credentials.
*/
package api
