package jsonschema

import (
	"go/ast"
	"strings"

	"golang.org/x/tools/go/analysis"
)

var NoCommas = &analysis.Analyzer{
	Name: "jsonschema",
	Doc:  "Disallow unescaped commas in JSON Schema fields",
	Run:  noCommasRun,
}

var allowedKeys = map[string]bool{
	"additionalItems":      true,
	"additionalProperties": true,
	"allOf":                true,
	"anyOf":                true,
	"const":                true,
	"contains":             true,
	"contentEncoding":      true,
	"contentMediaType":     true,
	"default":              true,
	"definitions":          true,
	"dependencies":         true,
	"description":          true,
	"else":                 true,
	"enum":                 true,
	"example":              true,
	"examples":             true,
	"exclusiveMaximum":     true,
	"exclusiveMinimum":     true,
	"format":               true,
	"if":                   true,
	"items":                true,
	"maximum":              true,
	"maxItems":             true,
	"maxLength":            true,
	"maxProperties":        true,
	"minimum":              true,
	"minItems":             true,
	"minLength":            true,
	"minProperties":        true,
	"multipleOf":           true,
	"not":                  true,
	"oneOf":                true,
	"pattern":              true,
	"patternProperties":    true,
	"properties":           true,
	"propertyNames":        true,
	"readOnly":             true,
	"required":             true,
	"then":                 true,
	"title":                true,
	"type":                 true,
	"uniqueItems":          true,
}

func noCommasRun(pass *analysis.Pass) (any, error) {
	for _, file := range pass.Files {
		ast.Inspect(file, func(n ast.Node) bool {
			field, ok := n.(*ast.Field)
			if !ok || field.Tag == nil {
				return true
			}

			// Example tag: `jsonschema:"title=Deprecation Date,description=The exact date\\, when deprecated."`
			tag := strings.Trim(field.Tag.Value, "`")
			jsonschemaTag := parseTagValue(tag, "jsonschema")
			if jsonschemaTag == "" {
				return true
			}

			// First, check for unescaped commas in values
			if hasUnescapedCommas(jsonschemaTag) {
				// Generate fix by escaping unescaped commas
				fixedTag := escapeUnescapedCommas(jsonschemaTag)
				fixedFullTag := strings.ReplaceAll(tag, jsonschemaTag, fixedTag)

				pass.Report(analysis.Diagnostic{
					Pos:     field.Tag.Pos(),
					End:     field.Tag.End(),
					Message: "jsonschema: found unescaped comma",
					SuggestedFixes: []analysis.SuggestedFix{
						{
							Message: "jsonschema: add escape for comma using \\,",
							TextEdits: []analysis.TextEdit{
								{
									Pos:     field.Tag.Pos() + 1,
									End:     field.Tag.End() - 1,
									NewText: []byte(fixedFullTag),
								},
							},
						},
					},
				})
				return true
			}

			// Check if the tags are allowed keys only
			for k := range parseKeyValuePairs(jsonschemaTag) {
				if !allowedKeys[k] {
					pass.Report(analysis.Diagnostic{
						Pos:     field.Tag.Pos(),
						End:     field.Tag.End(),
						Message: "jsonschema: unknown key: '" + k + "'",
					})
					return true
				}
			}

			return true
		})
	}
	return nil, nil
}

// parseTagValue extracts the quoted value for a tag key, e.g. jsonschema:"foo"
func parseTagValue(tag, key string) string {
	prefix := key + ":"
	start := strings.Index(tag, prefix)
	if start == -1 {
		return ""
	}
	start += len(prefix)
	if start >= len(tag) || tag[start] != '"' {
		return ""
	}
	end := strings.Index(tag[start+1:], `"`)
	if end == -1 {
		return ""
	}
	return tag[start+1 : start+1+end]
}

// hasUnescapedCommas checks if the jsonschema tag contains unescaped commas in values
func hasUnescapedCommas(s string) bool {
	pairs := parseKeyValuePairs(s)
	for _, value := range pairs {
		if containsUnescapedComma(value) {
			return true
		}
	}
	return false
}

// isLikelyKey checks if a string looks like a jsonschema key
func isLikelyKey(s string) bool {
	// A key should contain only letters (and maybe numbers)
	if len(s) == 0 {
		return false
	}
	for _, ch := range s {
		if !((ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z')) {
			return false
		}
	}
	return allowedKeys[s]
}

// containsUnescapedComma checks if a value contains an unescaped comma
func containsUnescapedComma(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			// Check if it's escaped
			if i > 0 && s[i-1] == '\\' {
				continue
			}
			return true
		}
	}
	return false
}

// escapeUnescapedCommas escapes all unescaped commas in the jsonschema tag
func escapeUnescapedCommas(s string) string {
	var builder strings.Builder
	i := 0

	for i < len(s) {
		// Skip whitespace
		for i < len(s) && s[i] == ' ' {
			builder.WriteByte(' ')
			i++
		}
		if i >= len(s) {
			break
		}

		// Find the next = or , to determine if this is a standalone key or key=value
		j := i
		for j < len(s) && s[j] != '=' && s[j] != ',' {
			j++
		}

		if j >= len(s) || s[j] == ',' {
			// Standalone key - just copy it
			builder.WriteString(s[i:j])
			i = j
			if i < len(s) && s[i] == ',' {
				builder.WriteByte(',')
				i++
			}
			continue
		}

		// key=value pair
		builder.WriteString(s[i : j+1]) // Write key and =
		i = j + 1

		// Find the value end and escape commas within it
		valueStart := i
		for i < len(s) {
			if i+1 < len(s) && s[i:i+2] == "\\," {
				// Already escaped comma, skip both characters
				i += 2
				continue
			}
			if s[i] == ',' {
				// Check if this is a separator between key=value pairs
				nextEq := strings.Index(s[i+1:], "=")
				nextComma := strings.Index(s[i+1:], ",")

				// Determine what comes after this comma
				var potentialKey string
				if nextEq == -1 && nextComma == -1 {
					// Rest of string
					potentialKey = strings.TrimSpace(s[i+1:])
				} else if nextEq == -1 {
					// Only commas ahead, take until next comma
					potentialKey = strings.TrimSpace(s[i+1 : i+1+nextComma])
				} else if nextComma == -1 || nextComma > nextEq {
					// Has = before next comma (or no next comma)
					potentialKey = strings.TrimSpace(s[i+1 : i+1+nextEq])
				} else {
					// Next comma comes before =, take until that comma
					potentialKey = strings.TrimSpace(s[i+1 : i+1+nextComma])
				}

				if isLikelyKey(potentialKey) {
					// This is a separator between key=value pairs
					break
				}
				// This comma is part of the value, continue
			}
			i++
		}

		// Escape commas in the value
		value := s[valueStart:i]
		builder.WriteString(escapeCommasInValue(value))

		// Add separator comma if we stopped at one
		if i < len(s) && s[i] == ',' {
			builder.WriteByte(',')
			i++
		}
	}

	return builder.String()
}

// escapeCommasInValue escapes unescaped commas in a single value
// Note: we need to write \\, in the source (double backslash) because Go interprets escape sequences
func escapeCommasInValue(s string) string {
	var result strings.Builder
	for i := 0; i < len(s); i++ {
		if s[i] == ',' {
			// Check if already escaped
			if i > 0 && s[i-1] == '\\' {
				result.WriteByte(',')
			} else {
				// Write \\, (backslash backslash comma) for Go source code
				result.WriteString("\\\\,")
			}
		} else {
			result.WriteByte(s[i])
		}
	}
	return result.String()
}

// parseKeyValuePairs parses key=value pairs separated by commas
// Handles escaped commas (\\,) within values and standalone keys (e.g., "required")
func parseKeyValuePairs(s string) map[string]string {
	result := make(map[string]string)
	i := 0

	for i < len(s) {
		// Skip whitespace
		for i < len(s) && s[i] == ' ' {
			i++
		}
		if i >= len(s) {
			break
		}

		// Find the next = or , to determine if this is a standalone key or key=value
		j := i
		for j < len(s) && s[j] != '=' && s[j] != ',' {
			j++
		}

		if j >= len(s) || s[j] == ',' {
			// Standalone key
			key := strings.TrimSpace(s[i:j])
			if key != "" {
				result[key] = ""
			}
			i = j
			if i < len(s) && s[i] == ',' {
				i++
			}
			continue
		}

		// key=value pair
		key := strings.TrimSpace(s[i:j])
		i = j + 1 // skip the '='

		// Find the value (until unescaped comma that precedes a known key)
		valueStart := i
		for i < len(s) {
			if i+1 < len(s) && s[i:i+2] == "\\," {
				// Escaped comma, skip
				i += 2
				continue
			}
			if s[i] == ',' {
				// Check if this is a separator
				nextEq := strings.Index(s[i+1:], "=")
				nextComma := strings.Index(s[i+1:], ",")

				// Determine what comes after this comma
				var potentialKey string
				if nextEq == -1 && nextComma == -1 {
					// Rest of string
					potentialKey = strings.TrimSpace(s[i+1:])
				} else if nextEq == -1 {
					// Only commas ahead, take until next comma
					potentialKey = strings.TrimSpace(s[i+1 : i+1+nextComma])
				} else if nextComma == -1 || nextComma > nextEq {
					// Has = before next comma (or no next comma)
					potentialKey = strings.TrimSpace(s[i+1 : i+1+nextEq])
				} else {
					// Next comma comes before =, take until that comma
					potentialKey = strings.TrimSpace(s[i+1 : i+1+nextComma])
				}

				if isLikelyKey(potentialKey) {
					// This comma is a separator
					break
				}
			}
			i++
		}

		value := s[valueStart:i]
		result[key] = value

		// Skip the comma separator
		if i < len(s) && s[i] == ',' {
			i++
		}
	}

	return result
}
