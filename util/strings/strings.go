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

// Remove searches 'array' for 'item' string
// and removes the string if found
func Remove(item string, array []string) []string {
	for i, v := range array {
		if v == item {
			array = append(array[:i], array[i+1:]...)
			break
		}
	}
	return array
}
