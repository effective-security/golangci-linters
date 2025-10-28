# golangci-linters

Custom Go linters for enforcing code quality standards.

## Linters

### jsonschema

Disallows unescaped commas in JSON Schema struct tags and validates allowed keys.

## Usage

This tool provides three ways to run the linters:

### 1. Standalone CLI (Recommended)

The easiest way to use the linter in any project:

```bash
# Install
go install github.com/effective-security/golangci-linters/cmd/custom-linters@latest

# Run
custom-linters ./...
```

### 2. Via go vet

Use as a standard Go vet tool:

```bash
# Install
go install github.com/effective-security/golangci-linters/cmd/custom-linters-vet@latest

# Run
go vet -vettool=$(which custom-linters-vet) ./...
```

### 3. As a golangci-lint plugin

**Note:** This approach has version compatibility limitations due to Go's plugin system requirements.

```yaml
# .golangci.yml
linters-settings:
  custom:
    jsontags:
      path: bin/custom-linters-v2.5.0.so
      description: JSON Schema tag validator
      original-url: github.com/effective-security/golangci-linters
```

**Important:** The plugin must be built with the exact same version of `golang.org/x/tools` as `golangci-lint`. This makes the plugin approach less flexible for use across different projects.

## Development

```bash
# Build all versions
make build

# Run tests
make test

# Run linter on this project
make lint
```

## Binaries

- `bin/custom-linters` - Standalone CLI tool
- `bin/custom-linters-vet` - For use with `go vet`
- `bin/custom-linters-v2.5.0.so` - Plugin for golangci-lint v2.5.0
