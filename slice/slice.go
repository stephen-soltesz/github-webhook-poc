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

// IndexString returns the index of value in slice, or -1 if value is not present.
func IndexString(slice []string, value string) int {
	for i := range slice {
		if slice[i] == value {
			return i
		}
	}
	return -1
}

func FilterStrings(slice []string, values []string) []string {
	result := []string{}
	for _, current := range slice {
		if ContainsString(values, current) {
			// Skip this value since we want it removed.
			continue
		}
		result = append(result, current)
	}
	return result
}
