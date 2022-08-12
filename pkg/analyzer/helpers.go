package analyzer

func includes[T comparable](slice []T, item T) bool {
	for _, si := range slice {
		if item == si {
			return true
		}
	}
	return false
}

func getOverlapping[T comparable](a []T, b []T) []T {
	res := []T{}
	for _, ai := range a {
		if includes(b, ai) {
			res = append(res, ai)
		}
	}
	return res
}
