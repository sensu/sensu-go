//go:generate go run assets_generate.go
package dashboard

import "net/http"

var App http.FileSystem = app()
var Lib http.FileSystem = lib()
var Vendor http.FileSystem = vendor()
