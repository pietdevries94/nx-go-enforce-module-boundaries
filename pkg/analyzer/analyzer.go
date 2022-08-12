package analyzer

import (
	"bytes"
	"encoding/json"
	"go/ast"
	"os"
	"path"
	"path/filepath"
	"strings"
	"sync"

	"golang.org/x/mod/modfile"
	"golang.org/x/tools/go/analysis"
)

// TODO implement allow

var Analyzer = &analysis.Analyzer{
	Name: "nxgoenforcemoduleboundaries",
	Doc:  "Checks that nx package boundaries are followed.",
	Run:  run,
}

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

func run(pass *analysis.Pass) (interface{}, error) {
	// if everything is allowed, linting makes no sense
	if importPrefix == "" || includes(allowed, "*") {
		return nil, nil
	}

	var pkgProj *projectFile

	inspect := func(node ast.Node) bool {
		importSpec, ok := node.(*ast.ImportSpec)
		if !ok {
			return true
		}

		p := strings.Trim(importSpec.Path.Value, `"`)
		if !strings.HasPrefix(p, importPrefix+"/") {
			return true
		}
		p = "./" + strings.TrimPrefix(p, importPrefix+"/")

		importProj := projectFileCache.getProjectFileForPath(p)

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

	for _, f := range pass.Files {
		pkgProj = projectFileCache.getProjectFileForPath(pass.Fset.File(f.Pos()).Name())
		if pkgProj != nil {
			ast.Inspect(f, inspect)
		}
	}
	return nil, nil
}

type projectFile struct {
	path string
	tags []string
}

type projectFileCacheFact struct {
	cache map[string]*projectFile
	m     sync.Mutex
}

func (*projectFileCacheFact) AFact() {}

func (pfc *projectFileCacheFact) getProjectFileForPath(p string) *projectFile {
	p, err := filepath.Abs(p)
	if err != nil {
		// TODO error message
		return nil
	}

	pfc.m.Lock()
	res, ok := pfc.cache[p]
	pfc.m.Unlock()
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
		res = pfc.getProjectFileForPath(filepath.Dir(p))
	}

	pfc.m.Lock()
	pfc.cache[p] = res
	pfc.m.Unlock()

	return res
}

type Boundaries struct {
	Allow          []string        `json:"allow"`
	DepConstraints []DepConstraint `json:"depConstraints"`
}

type DepConstraint struct {
	SourceTag                string   `json:"sourceTag"`
	OnlyDependOnLibsWithTags []string `json:"onlyDependOnLibsWithTags"`
	NotDependOnLibsWithTags  []string `json:"notDependOnLibsWithTags"`
}

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
