package interpolation

// UniqueStr will return a unique list of strings for a slice
func UniqueStr(input []string) []string {
	u := make([]string, 0, len(input))
	m := make(map[string]bool)

	for _, val := range input {
		if _, ok := m[val]; !ok {
			m[val] = true
			u = append(u, val)
		}
	}
	return u
}

// ChunkIterator will return a chunk of a specific limit of a list of strings
func ChunkIterator(listOfStrings []*string, limit int) func() []*string {
	i := 0
	generator := func() []*string {
		if i < len(listOfStrings) {
			batch := listOfStrings[i:min(i+limit, len(listOfStrings))]
			i += limit
			return batch
		}
		return nil
	}
	return generator
}

// min will return the minimum number
func min(a, b int) int {
	if a <= b {
		return a
	}
	return b
}
