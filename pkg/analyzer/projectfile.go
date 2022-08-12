package analyzer

import (
	"encoding/json"
	"os"
	"path"
	"path/filepath"
)

func getProjectFileForPath(p string) *projectFile {
	p, err := filepath.Abs(p)
	if err != nil {
		// TODO error message
		return nil
	}

	projectFileCache.m.Lock()
	res, ok := projectFileCache.cache[p]
	projectFileCache.m.Unlock()
	if ok {
		return res
	}

	// If project file in current folder, get it and cache the current path
	if _, err := os.Stat(path.Join(p, "project.json")); err == nil {
		res = &projectFile{
			path: p,
			tags: loadTagsForProjectFile(path.Join(p, "project.json")),
		}
	} else if p == "." || p == "/" {
		return nil
	} else {
		res = getProjectFileForPath(filepath.Dir(p))
	}

	projectFileCache.m.Lock()
	projectFileCache.cache[p] = res
	projectFileCache.m.Unlock()

	return res
}

func loadTagsForProjectFile(p string) []string {
	f, err := os.Open(p)
	if err != nil {
		// TODO better handling
		return []string{}
	}
	defer f.Close()

	pr := Project{}
	err = json.NewDecoder(f).Decode(&pr)
	if err != nil {
		// TODO better handling
		return []string{}
	}

	return pr.Tags
}

type Project struct {
	Tags []string `json:"tags"`
}
