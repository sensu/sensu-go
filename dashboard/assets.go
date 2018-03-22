//+build !dev
//go:generate yarn install
//go:generate yarn build
//go:generate go run assets_generate.go

package dashboard

import (
	"net/http"
)

//
// -----------------------------------------------------------------------------
// Fallback
// -----------------------------------------------------------------------------
//
// To allow for builds of the Sensu web UI without requiring the end-user to
// have node.js and yarn installed. Or, when a developer may want to quickly
// build the backend without re-building the entire dashboard.
//
// Fallback simply provides a empty filesystem implementation that returns a
// message informing the user that the dashboard is not present in the build.
//

const fallbackMessage = `
Sensu web UI was not included in this build.
Find build instructions in Sensu Github repository or use a pre-built binary.
`

// Assets implements http.FileSystem returning web UI's assets.
var Assets http.FileSystem = &fallbackFS{fallbackMessage}
