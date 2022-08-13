package analyzer

import (
	"go/ast"

	"golang.org/x/tools/go/analysis"
)

func newInspector[T any](pass *analysis.Pass, node ast.Node, pkgProj *projectFile, v T) inspector[T] {
	done := false

	return inspector[T]{
		reportf: func(format string, v ...any) {
			pass.Reportf(node.Pos(), format, v...)
		},
		pkgProj: pkgProj,
		done:    &done,
		value:   v,
	}
}

type inspector[T any] struct {
	reportf func(string, ...any)
	pkgProj *projectFile
	done    *bool
	value   T
}

func runInspector[T, U any](i inspector[T], f func(reportf func(string, ...any), pkgProj *projectFile, v T) (U, bool)) inspector[U] {
	if *i.done {
		return inspector[U]{
			reportf: i.reportf,
			pkgProj: i.pkgProj,
			done:    i.done,
		}
	}

	res, done := f(i.reportf, i.pkgProj, i.value)
	*i.done = done
	return inspector[U]{
		reportf: i.reportf,
		pkgProj: i.pkgProj,
		done:    i.done,
		value:   res,
	}
}

func runInspectorNoRes[T any](i inspector[T], f func(reportf func(string, ...any), pkgProj *projectFile, v T) bool) {
	if *i.done {
		return
	}

	done := f(i.reportf, i.pkgProj, i.value)
	*i.done = done
}
