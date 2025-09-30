# Agent Guidelines for Dailies Codebase

## Build/Test Commands

- **Backend Build**: `go build -o build/api`
- **Backend Run**: `go run main.go` or `air` (hot reload)
- **Backend Test**: `go test ./...`
- **Single test**: `go test -run TestName ./package`
- **Test with coverage**: `go test -cover ./...`
- **Frontend**: `cd frontends/web && npm start` (dev server), `npm run build`, `npm test`

## Code Style & Conventions

- **Language**: Go 1.25.1 with Gin framework, GORM ORM, SQLite database
- **Frontend**: Angular 20+ in `frontends/web/` with Prettier (printWidth: 100, singleQuote: true)
- **Imports**: Standard library first, then third-party, then local packages separated by blank lines
- **Naming**: Use camelCase for JSON tags, PascalCase for Go structs/functions
- **Error handling**: Return errors explicitly, use `log.Fatal()` for startup errors, `log.Println()` for info
- **Database**: Use GORM for ORM, pointer types for optional fields (*string, *int)
- **API**: RESTful endpoints under `/api`, use Gin handlers with dependency injection pattern
- **Structure**: Models in `/models`, handlers in `/handlers`, middleware in `/middleware`, config in `/config`
- **Documentation**: ALL packages and functions MUST have proper godoc comments
- **Testing**: All new code MUST include unit tests with good coverage

## Patterns

- Handlers receive `*gorm.DB` and return `gin.HandlerFunc`
- Use UUIDs for primary keys with BeforeCreate hooks
- JSON tags with omitempty for optional fields
- CORS middleware allows all origins
