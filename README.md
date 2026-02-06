# Sotekre — Menu Tree (galih-jawaban)

[![CI](https://github.com/galpt/sotekre/actions/workflows/ci.yml/badge.svg?branch=main)](https://github.com/galpt/sotekre/actions/workflows/ci.yml) [![Release](https://github.com/galpt/sotekre/actions/workflows/release.yml/badge.svg?branch=main)](https://github.com/galpt/sotekre/actions/workflows/release.yml) [![Coverage](https://codecov.io/gh/galpt/sotekre/branch/main/graph/badge.svg)](https://codecov.io/gh/galpt/sotekre)

A small full‑stack app that demonstrates a hierarchical menu/tree with CRUD, transactional reorder/move, and a Next.js + Tailwind UI.

---

## Table of contents
- [Status](#status)
- [Quickstart](#quickstart)
- [Architecture & design](#architecture--design)
- [API & docs](#api--docs)
- [Database (ERD + migrations)](#database-erd--migrations)
- [Deliverables checklist](#deliverables-checklist)
- [How to run (dev & production)](#how-to-run-dev--production)
- [Tests & verification](#tests--verification)

---

## Status
- MVP complete: backend (Go + Gin + GORM), frontend (Next.js + Tailwind), Docker compose, Swagger/OpenAPI, unit & integration tests for core flows.
- Stable enough for an interview demo (local scripts for Windows included).

## Quickstart (dev)
1. Rename for XAMPP (one-liner):
   - Windows (cmd/PowerShell): `ren backend\.env.example .env`
2. With Docker (recommended):
   - Docker Compose runs an isolated MySQL instance and therefore **requires** a non-empty MySQL root password. The project includes a demo password by default for convenience — do not use that in production.

```bash
# quick (uses demo password from compose file)
docker compose up --build

# recommended: create a docker env file, edit credentials, then run
cp .env.docker.example .env.docker
# edit .env.docker (set a non-empty MYSQL_ROOT_PASSWORD)
docker compose --env-file .env.docker up --build
```

- Backend: `http://localhost:8080`  (Swagger: `http://localhost:8080/swagger/index.html`)
- Frontend (Next dev): `http://localhost:3000` (when run locally) — in Docker the frontend proxies to the backend service via `NEXT_PUBLIC_API_URL`.

### Releases (prebuilt binaries)
Prebuilt backend binaries and release artifacts are published to GitHub Releases by CI (when a tag or manual release is created). Download the appropriate archive for your platform from the repository's Releases page.

Verification (recommended):
```bash
# verify SHA256 checksum after downloading the archive
sha256sum <archive-file>
# compare against SHA256SUMS.txt shipped with the release
```

3. Without Docker (local MySQL / XAMPP):
```bash
cd backend
go mod tidy
# optional: generate docs -> go generate ./...
go run .
```
Frontend:
```bash
cd frontend
npm install
npm run dev
```
> [!INFO]
> Windows quick demo: run `compile_golang.bat` then `run_project.bat` from repo root.

---

## Architecture & design
- Backend: Go (Gin) + GORM, adjacency‑list menu model (`parent_id`), transactional move/reorder logic to keep sibling ordering consistent.
- Frontend: Next.js (TypeScript) + Tailwind — native HTML5 drag‑and‑drop wired to PATCH endpoints.
- DB: MySQL (dev via Docker/XAMPP). Tests use in‑memory SQLite.

---

## API & docs
- OpenAPI (generated): `backend/docs/swagger.json`
- Swagger UI (runtime): `http://localhost:8080/swagger/index.html`
- Core endpoints:
  - GET  /api/menus
  - POST /api/menus
  - PUT  /api/menus/:id
  - PATCH /api/menus/:id/reorder
  - PATCH /api/menus/:id/move
  - DELETE /api/menus/:id

(Generate docs locally: `cd backend && go generate ./...` or `go run github.com/swaggo/swag/cmd/swag@latest init -g main.go -o ./docs`)

---

## Database (ERD & migrations)
- ERD (Mermaid): `backend/database/ERD.md` (source of truth for reviewers)
- Migration: `backend/migrations/001_create_menus.sql`
- Model: `backend/models/menu.go` (GORM struct + `AutoMigrate` in `main.go`)

> [!INFO]
> Quick DB facts:
> - Single table `menus` with self‑referencing `parent_id` (unlimited depth)
> - Sibling ordering via `order` integer; server enforces reindexing in transactions

---

## Deliverables checklist
(links point to the repository locations)

- [x] Follow best practices (service layer, validation, error handling)
  - Evidence: `backend/services/*`, `backend/handlers/*` (clear separation)
- [x] Clear folder structure & complete source
  - Evidence: `backend/`, `frontend/`, `docker-compose.yml`
- [x] README with setup, dev, prod, Docker, API docs, design notes
  - Evidence: this file (`README.md`) — expanded; `backend/docs/` for API
- [x] Database schema / migrations
  - Evidence: `backend/migrations/001_create_menus.sql`, `backend/models/menu.go`
- [x] Environment template (`.env.example`) and XAMPP‑ready defaults
  - Evidence: `backend/.env.example`
- [x] Docker configuration (bonus)
  - Evidence: `docker-compose.yml`, `backend/Dockerfile`, `frontend/Dockerfile`
- [x] Basic test coverage (unit + integration for critical logic)
  - Evidence: `backend/services/*_test.go`, `backend/handlers/*_test.go`

> [!INFO]
> Core deliverables are implemented and well-documented.

---

## How to verify (quick)
- Run backend tests: `cd backend && go test ./... -v`
- Generate coverage locally: `cd backend && go test -coverprofile=coverage.out ./... && go tool cover -func=coverage.out`
- Typecheck frontend: `cd frontend && npx tsc --noEmit`
- Generate & view docs: `cd backend && go generate ./...` → open `/swagger/index.html`

> Coverage: the badge at the top of this README shows the latest Go coverage reported by CI (may require one CI run to appear).

---

## Tests & verification
- Unit + integration tests cover core DB/service behaviors (tree build, move/reorder, recursive delete).

---

## Contributing
- See `backend/` for Go code and `frontend/` for Next.js. Run tests locally and open a PR.

---

## License
MIT

