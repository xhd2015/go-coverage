package myers

// ComputeBlockMapping
// a block is represented in a string
// the function computes
// the result is a mapping from each index in new to its
// counterpart in old, -1 if new.
func ComputeBlockMapping(old []string, new []string) map[int]int {
	m := make(map[int]int, len(new))
	operationsComplex(old, new, func(x, y int) {
		m[y] = x
	})
	return m
}
