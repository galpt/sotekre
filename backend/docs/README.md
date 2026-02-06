# Generated API Documentation

This directory contains auto-generated OpenAPI (Swagger) documentation for the backend API.

## Generation

The files in this directory are produced by the `swag` tool from source code annotations.

To regenerate the documentation:

```bash
cd backend
go generate ./...
```

## Files

- `docs.go` — Generated Go package that embeds the OpenAPI specification.
- `swagger.json` — OpenAPI 3.0 JSON specification (generated).
- `openapi.json` — Copy of `swagger.json` for compatibility with consumers that expect `openapi.json`.
- `swagger.yaml` — YAML representation of the OpenAPI specification.
- `swagger.html` — Simple static page (redirect) to the Swagger UI.

> [!NOTE]
> - These files are generated; do not edit them directly. Make changes to the source annotations in the Go code (for example, in `handlers/*.go` and `main.go`) and then regenerate.
> - The project commits the generated documentation because the application imports the `docs` package and CI/coverage tooling requires the files to be present.

## CI and Coverage

The CI pipeline regenerates documentation as part of the workflow to keep committed docs synchronized with source annotations. If you update API annotations, run the generation command and commit the resulting files.

If you have questions about the generation process or CI behavior, open an issue or contact the repository maintainer.
