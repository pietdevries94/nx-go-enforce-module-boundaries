package analyzer

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"os"
	"path"
	"path/filepath"
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
			ast.Inspect(f, inspect(pass, pkgProj))
		}
	}
	return nil, nil
}

func inspect(pass *analysis.Pass, pkgProj *projectFile) func(node ast.Node) bool {
	return func(node ast.Node) bool {
		importSpec, ok := node.(*ast.ImportSpec)
		if !ok {
			return true
		}

		p := strings.Trim(importSpec.Path.Value, `"`)
		if !strings.HasPrefix(p, importPrefix+"/") {
			return true
		}
		p = "./" + strings.TrimPrefix(p, importPrefix+"/")

		importProj := getProjectFileForPath(p)

		// If in the same project, no error
		if importProj == pkgProj {
			return true
		}

		// TODO no tags
		for _, pkgTag := range pkgProj.tags {
			dc := mappedDepConstraint[pkgTag]

			// NotDependOnLibsWithTags
			if matching := getOverlapping(dc.NotDependOnLibsWithTags, importProj.tags); len(matching) > 0 {
				pass.Reportf(node.Pos(), `A project tagged with "%s" can not depend on libs tagged with %s`,
					pkgTag, stringifyTags(matching))
				return true
			}

			// OnlyDependOnLibsWithTags
			if matching := getOverlapping(dc.OnlyDependOnLibsWithTags, importProj.tags); len(matching) == 0 {
				pass.Reportf(node.Pos(), `A project tagged with "%s" can only depend on libs tagged with %s`,
					pkgTag, stringifyTags(dc.OnlyDependOnLibsWithTags))
				return true
			}
		}
		return true
	}
}

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
