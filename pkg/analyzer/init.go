package analyzer

import (
	"encoding/json"
	"os"

	"golang.org/x/mod/modfile"
)

var importPrefix = ""
var allowed []string
var mappedDepConstraint map[string]DepConstraint = make(map[string]DepConstraint)
var projectFileCache = &projectFileCacheFact{
	cache: map[string]*projectFile{},
}

func init() {
	// TODO: better find method with moving up. Probably needs to move to the runner in case of multiple mods in the workspace
	mod, err := os.ReadFile("./go.mod")
	if err != nil {
		return
	}
	importPrefix = modfile.ModulePath(mod)

	// TODO: better find method with moving up
	boundariesFile, err := os.Open("./boundaries.json")
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
