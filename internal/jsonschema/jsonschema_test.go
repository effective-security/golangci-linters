package jsonschema_test

import (
	"testing"

	"github.com/effective-security/golangci-linters/internal/jsonschema"
	"golang.org/x/tools/go/analysis/analysistest"
)

func TestJsonSchema(t *testing.T) {
	testDataDir := analysistest.TestData()

	analysistest.Run(t, testDataDir, jsonschema.NoCommas, "./src/jsonschema/")
}

func TestJsonSchemaAutoFix(t *testing.T) {
	testDataDir := analysistest.TestData()

	results := analysistest.RunWithSuggestedFixes(t, testDataDir, jsonschema.NoCommas, "./src/jsonschema/")
	suggestedFixProvided := false
	for _, result := range results {
		for _, diagnostic := range result.Diagnostics {
			for _, suggestedFix := range diagnostic.SuggestedFixes {
				if len(suggestedFix.TextEdits) != 0 {
					suggestedFixProvided = true
				}
			}
		}
	}

	if !suggestedFixProvided {
		t.Errorf("expected a suggested fix to be provided, but didn't have any in %+v", results)
	}
}
