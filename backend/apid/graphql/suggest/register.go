package suggest

// Register is used to store and later find Sensu resources
type Register []*Resource

// Lookup finds a Resource given a ref
func (r Register) Lookup(ref RefComponents) *Resource {
	for _, res := range r {
		if ref.Group == res.Group && ref.Name == res.Name {
			return res
		}
	}
	return nil
}
