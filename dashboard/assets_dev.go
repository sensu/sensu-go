//+build dev

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
Sensu web UI is not included in developer builds. While developing prefer use of
webpack dev-server. If, you did not intend to use a dev build refer to build
instructions in Sensu Github repository or use a pre-built binary.
`

// Assets implements http.FileSystem returning web UI's assets.
var Assets http.FileSystem = &fallbackFS{msg: fallbackMessage}
