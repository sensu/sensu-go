package patch

import (
	jsonpatch "github.com/evanphx/json-patch"
)

// Patcher abstracts the patching mechanism and allows easier testing
type Patcher interface {
	Patch(original, patch []byte) ([]byte, error)
}

// Merge is a patcher for Merge Patchs as defined in RFC7396
type Merge struct{}

// Patch merges the patch into the original document
func (m *Merge) Patch(original, patch []byte) ([]byte, error) {
	return jsonpatch.MergePatch(original, patch)
}
