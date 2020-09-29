package patch

import (
	jsonpatch "github.com/evanphx/json-patch/v5"
)

// Patcher abstracts the patching mechanism and allows easier testing
type Patcher interface {
	Patch(document []byte) ([]byte, error)
}

// Merge is a patcher for Merge Patchs as defined in RFC7396
type Merge struct {
	MergePatch []byte
}

// Patch merges the patch into the original document
func (m *Merge) Patch(document []byte) ([]byte, error) {
	return jsonpatch.MergePatch(document, m.MergePatch)
}
