package stringutil

// OccurrencesOf returns the number of times a string appears in the given slice
// of strings.
func OccurrencesOf(s string, in []string) int {
	o := NewOccurrenceSet(in...)
	return o.Get(s)
}

// OccurrenceSet captures of occurrences of string values.
type OccurrenceSet map[string]int

// NewOccurrenceSet returns new instance of OccurrenceSet.
func NewOccurrenceSet(s ...string) OccurrenceSet {
	o := OccurrenceSet{}
	o.Add(s...)
	return o
}

// Add entry and increment count
func (o OccurrenceSet) Add(ss ...string) {
	for _, s := range ss {
		num := o[s]
		o[s] = num + 1
	}
}

// Remove items from set
func (o OccurrenceSet) Remove(ss ...string) {
	for _, s := range ss {
		delete(o, s)
	}
}

// Get returns occurrences of given string
func (o OccurrenceSet) Get(entry string) int {
	return o[entry]
}

// Values returns all occurrences
func (o OccurrenceSet) Values() []string {
	vs := []string{}
	for v := range o {
		vs = append(vs, v)
	}
	return vs
}

// Merge given set of occurrences
func (o OccurrenceSet) Merge(b OccurrenceSet) {
	for name, bCount := range b {
		aCount := o[name]
		o[name] = aCount + bCount
	}
}

// Size of values tracked
func (o OccurrenceSet) Size() int {
	return len(o)
}
