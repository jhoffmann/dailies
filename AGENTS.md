# AGENTS.md - Developer Guidelines

## Build/Test/Lint Commands

- `mise run dev` - Standard development workflow (deps, fmt, vet, test)
- `mise run test` - Run all tests with coverage: `go test -tags=coverage ./... -cover`
- `go test ./path/to/package` - Run tests for specific package
- `mise run fmt` - Format code: `go fmt ./...`
- `mise run vet` - Vet code: `go vet ./...`
- `mise run build` - Build server and client binaries
- `bun run build` - Build frontend assets with webpack
- `bun run dev` - Watch frontend assets during development

## Code Style Guidelines

- **Language**: Go 1.24.6 with GORM for database operations, HTMX for frontend
- **Imports**: Standard library first, then external packages, then internal packages
- **Naming**: Use camelCase for Go variables/functions, PascalCase for exported types
- **Comments**: Document exported functions with purpose and behavior (see handlers/task.go examples)
- **Error Handling**: Return descriptive HTTP errors with appropriate status codes
- **Database**: Use GORM models with struct tags, UUID primary keys, timestamps
- **JSON**: Use json tags on structs, set Content-Type headers in handlers
- **Validation**: Validate input data before database operations (required fields, UUID parsing)

## No Special Rules

No Cursor rules or Copilot instructions found in this repository.

