package main

import (
	"golang.org/x/tools/go/analysis/singlechecker"

	"github.com/aarif123456/safegoroutines/pkg/analyzer"
)

func main() {
	singlechecker.Main(analyzer.Analyzer)
}
