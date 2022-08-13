package analyzer

import (
	"encoding/json"

	"golang.org/x/mod/modfile"
)

var importPrefix = ""
var allowed []string
var mappedDepConstraint map[string]DepConstraint = make(map[string]DepConstraint)
var projectFileCache = &projectFileCacheFact{
	cache: map[string]*projectFile{},
}

func init() {
	mod, err := findAndReadFile("./go.mod")
	if err != nil {
		return
	}
	importPrefix = modfile.ModulePath(mod)

	boundariesFile, err := findAndOpenFile("./boundaries.json")
	if err != nil {
		return
	}
	defer boundariesFile.Close()

	b := Boundaries{}
	err = json.NewDecoder(boundariesFile).Decode(&b)
	if err != nil {
		return
	}
	allowed = b.Allow

	for _, dc := range b.DepConstraints {
		mappedDepConstraint[dc.SourceTag] = dc
	}
}
