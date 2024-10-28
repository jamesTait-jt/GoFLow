package slice

func Contains[T comparable](slice []T, val T) bool {
	for i := 0; i < len(slice); i++ {
		if slice[i] == val {
			return true
		}
	}

	return false
}
