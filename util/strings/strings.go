package strings

// InArray searches 'array' for 'item' string
// Returns true 'item' is a value of 'array'
func InArray(item string, array []string) bool {
	if item == "" || len(array) == 0 {
		return false
	}

	for _, element := range array {
		if element == item {
			return true
		}
	}

	return false
}
