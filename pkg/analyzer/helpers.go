package analyzer

import (
	"bytes"
	"os"
	"path/filepath"
)

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

func findFile(p string) (string, bool) {
	p, err := filepath.Abs(p)
	if err != nil {
		panic(err)
	}
	_, err = os.Stat(p)
	if err == nil {
		return p, true
	}
	dir, file := filepath.Split(p)
	if dir == string(filepath.Separator) {
		return "", false
	}
	return findFile(filepath.Join(dir, "..", file))
}

func findAndReadFile(p string) ([]byte, error) {
	p, ok := findFile(p)
	if !ok {
		return nil, os.ErrNotExist
	}
	return os.ReadFile(p)
}
func findAndOpenFile(p string) (*os.File, error) {
	p, ok := findFile(p)
	if !ok {
		return nil, os.ErrNotExist
	}
	return os.Open(p)
}
