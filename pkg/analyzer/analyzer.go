package analyzer

import (
	"fmt"
	"go/ast"
	"io/ioutil"
	"os"
	"path"
	"strings"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/analysis"
)

func init() {
	// TODO: better find method wifh moving up. Probably needs to move to the runner in case of multiple mods in the workspace
	mod, err := ioutil.ReadFile("./go.mod")
	if err != nil {
		return
	}
	importPrefix = modfile.ModulePath(mod)
}

var importPrefix = ""
var projectFileCache = &projectFileCacheFact{
	cache: map[string]*projectFile{},
}

var Analyzer = &analysis.Analyzer{
	Name: "nxgoenforcemoduleboundaries",
	Doc:  "Checks that nx package boundaries are followed.",
	Run:  run,
	FactTypes: []analysis.Fact{
		projectFileCache,
	},
}

func run(pass *analysis.Pass) (interface{}, error) {
	if importPrefix == "" {
		return nil, nil
	}

	pkgProj := projectFileCache.getProjectFileForPath(pass.Pkg.Name())

	inspect := func(node ast.Node) bool {
		importSpec, ok := node.(*ast.ImportSpec)
		if !ok {
			return true
		}

		p := strings.Trim(importSpec.Path.Value, `"`)
		if !strings.HasPrefix(p, importPrefix+"/") {
			return true
		}

		importProj := projectFileCache.getProjectFileForPath(p)

		// If in the same project, no error
		if importProj == pkgProj {
			return true
		}

		pass.Reportf(node.Pos(), "has import %s",
			importSpec.Path.Value)
		return true
	}

	for _, f := range pass.Files {
		ast.Inspect(f, inspect)
	}
	return nil, nil
}

type projectFile struct {
	path string
}

type projectFileCacheFact struct {
	cache map[string]*projectFile
}

func (projectFileCacheFact) AFact() {}

func (pfc *projectFileCacheFact) getProjectFileForPath(p string) *projectFile {
	fmt.Println(p)

	res, ok := pfc.cache[p]
	if ok {
		return res
	}

	// If project file in current folder, get it and cache the current path
	if _, err := os.Stat(path.Join(p, "project.json")); err == nil {
		res = &projectFile{
			path: p,
		}
	} else {
		// TODO: If it isn't here, check if it's the top level of the module
		res = pfc.getProjectFileForPath(path.Join(p, ".."))
	}

	pfc.cache[p] = res
	return res
}
