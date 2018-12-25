package slice

// ContainsString looks for value in the given slice
func ContainsString(slice []string, value string) bool {
	for i := range slice {
		if slice[i] == value {
			return true
		}
	}
	return false
}
