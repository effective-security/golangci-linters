package main

import (
	"github.com/effective-security/golangci-linters/internal/jsonschema"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/multichecker"
)

func main() {
	multichecker.Main(
		jsonschema.NoCommas,
	)
}

func New(conf any) ([]*analysis.Analyzer, error) {
	return []*analysis.Analyzer{
		jsonschema.NoCommas,
	}, nil
}
