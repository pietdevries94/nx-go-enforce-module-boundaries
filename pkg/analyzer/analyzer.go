package analyzer

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

// TODO implement allow

var Analyzer = &analysis.Analyzer{
	Name: "nxgoenforcemoduleboundaries",
	Doc:  "Checks that nx package boundaries are followed.",
	Run:  run,
}

func run(pass *analysis.Pass) (interface{}, error) {
	// if everything is allowed, linting makes no sense
	if importPrefix == "" || includes(allowed, "*") {
		return nil, nil
	}

	for _, f := range pass.Files {
		pkgProj := getProjectFileForPath(pass.Fset.File(f.Pos()).Name())
		if pkgProj != nil {
			ast.Inspect(f, func(node ast.Node) bool {
				return inspect(pass, pkgProj, node)
			})
		}
	}
	return nil, nil
}

func inspect(pass *analysis.Pass, pkgProj *projectFile, node ast.Node) bool {
	n := newInspector(pass, node, pkgProj, node)

	importSpec := runInspector(n, getImportSpecOrDone)
	importProj := runInspector(importSpec, getImportProjectOrDone)
	runInspector(importProj, notSameProjectOrDone)
	runInspector(importProj, noDependenciesWithBannedTagsOrDone)
	runInspector(importProj, onlyDependenciesWithAllowedTagsOrDone)

	return true
}

func getImportSpecOrDone(reportf func(string, ...any), pkgProj *projectFile, node ast.Node) (*ast.ImportSpec, bool) {
	importSpec, ok := node.(*ast.ImportSpec)
	return importSpec, !ok
}

func getImportProjectOrDone(reportf func(string, ...any), pkgProj *projectFile, importSpec *ast.ImportSpec) (*projectFile, bool) {
	p := strings.Trim(importSpec.Path.Value, `"`)
	if !strings.HasPrefix(p, importPrefix+"/") {
		return nil, true
	}
	p = "./" + strings.TrimPrefix(p, importPrefix+"/")

	importProj := getProjectFileForPath(p)
	return importProj, false
}

func notSameProjectOrDone(reportf func(string, ...any), pkgProj *projectFile, importProj *projectFile) (bool, bool) {
	eq := pkgProj == importProj
	return !eq, eq
}

func noDependenciesWithBannedTagsOrDone(reportf func(string, ...any), pkgProj *projectFile, importProj *projectFile) (bool, bool) {
	for _, pkgTag := range pkgProj.tags {
		dc := mappedDepConstraint[pkgTag]

		// NotDependOnLibsWithTags
		if matching := getOverlapping(dc.NotDependOnLibsWithTags, importProj.tags); len(matching) > 0 {
			reportf(`A project tagged with "%s" can not depend on libs tagged with %s`,
				pkgTag, stringifyTags(matching))
			return false, true
		}
	}
	return true, false
}
func onlyDependenciesWithAllowedTagsOrDone(reportf func(string, ...any), pkgProj *projectFile, importProj *projectFile) (bool, bool) {
	for _, pkgTag := range pkgProj.tags {
		dc := mappedDepConstraint[pkgTag]

		// OnlyDependOnLibsWithTags
		if matching := getOverlapping(dc.OnlyDependOnLibsWithTags, importProj.tags); len(matching) == 0 {
			reportf(`A project tagged with "%s" can only depend on libs tagged with %s`,
				pkgTag, stringifyTags(dc.OnlyDependOnLibsWithTags))
			return false, true
		}
	}
	return true, false
}
