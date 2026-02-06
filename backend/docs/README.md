# API Documentation

This directory contains auto-generated API documentation files for the backend.

## Files

- **docs.go** - Go code file with embedded Swagger metadata
- **swagger.json** - Swagger/OpenAPI v2 specification
- **swagger.yaml** - Swagger/OpenAPI v2 specification (YAML format)
- **openapi.json** - Copy of swagger.json for compatibility
- **swagger.html** - Redirect page to Swagger UI

## Regeneration

If you modify API handlers or add new endpoints, regenerate the docs:

```bash
cd backend
go generate ./...
```

This will run `swag init` to parse your Go code comments and regenerate all documentation files.

## CI/CD

The GitHub Actions CI workflow automatically generates these files during builds to ensure they're always up-to-date. However, committing them to the repository ensures:

1. **Codecov compatibility** - Codecov upload requires docs.go to exist
2. **Local development** - Developers can immediately see API docs without running `go generate`
3. **Documentation tracking** - API changes are visible in git diffs

> [!IMPORTANT]
> **Do not edit these files manually.** They are auto-generated from code comments in `backend/handlers/*.go` and `backend/main.go`.
