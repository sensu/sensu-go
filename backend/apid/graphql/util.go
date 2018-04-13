package graphql

// constrainInt returns int within range.
func constrainInt(num, min, max int) int {
	if num < min {
		return min
	} else if num > max {
		return max
	}
	return num
}
