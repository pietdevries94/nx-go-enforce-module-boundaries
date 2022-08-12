package analyzer

import "bytes"

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

func stringifyTags(tags []string) string {
	b := bytes.NewBufferString(`"`)

	for i, t := range tags {
		if i != 0 && i+1 == len(tags) {
			b.WriteString(`" or "`)
		} else if i > 0 {
			b.WriteString(`", "`)
		}
		b.WriteString(t)
	}

	b.WriteString(`"`)
	return b.String()
}
