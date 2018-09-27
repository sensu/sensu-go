// Package asset provides a mechanism for installing, managing, and utilizing
// Sensu Assets.
//
// Access to assets are serialized. When an asset is first encountered,
// getting the asset from the manager blocks until the asset has been
// fetched, verified, and expanded on the host filesystem (or deemed
// unnecessary due to asset filters).
//
// The first goroutine to get an asset will cause the installation, and
// subsequent calling goroutines will simply block while installation
// completes. If the initial installation fails, the next goroutine to
// unblock will attempt reinstallation.
package asset

import (
	"github.com/sensu/sensu-go/types"
)

// A Getter is responsible for fetching (based on fitler selection), verifying,
// and expanding an asset. Calls to the Get method block until the Asset has
// fetched, verified, and expanded or it returns an error indicating why getting
// the asset failed.
type Getter interface {
	Get(*types.Asset) (*RuntimeAsset, error)
}

// An RuntimeAsset is a locally expanded Asset. After downloading, verifying,
// and expanding the Asset, the RuntimeAsset struct contains everything
// necessary to create a runtime environment composed of one or more
// RuntimeAssets.
type RuntimeAsset struct {
	// Path is the absolute path to the asset's base directory.
	Path string
}
