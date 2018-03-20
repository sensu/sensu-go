//+build release
//go:generate yarn install
//go:generate yarn build
//go:generate go run httphandler_generate.go

package dashboard

import "net/http"

func init() {
	HTTPHandler = http.FileServer(httpAssets)
}
