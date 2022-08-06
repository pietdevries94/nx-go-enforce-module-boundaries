package main

import (
	"github.com/pietdevries94/nx-go-enforce-module-boundaries/pkg/analyzer"
	"golang.org/x/tools/go/analysis/singlechecker"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
