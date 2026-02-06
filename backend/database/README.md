# Database — Menu Tree (backend)

Authoritative database documentation for reviewers and contributors. This file mirrors the main project README style (clickable TOC, quick verification, reviewer checklist) so reviewers can find DB information immediately.

---

## Table of contents
- [Status](#status)
- [Quick verification](#quick-verification)
- [Schema & authoritative sources](#schema--authoritative-sources)
- [ERD (visual)](#erd-visual)
- [Migrations & running locally](#migrations--running-locally)
- [Example queries (verification)](#example-queries-verification)
- [Indexes, constraints & rationale](#indexes-constraints--rationale)
- [Testing & CI guidance](#testing--ci-guidance)
- [Reviewer checklist](#reviewer-checklist)
- [Reference files](#reference-files)

---

## Status
- Ready for review: schema, migration, and service invariants are implemented and covered by unit + integration tests.
- Scope: single-table adjacency-list (`menus`) with transactional move/reorder and recursive delete handled by application logic.

## Quick verification
Commands reviewers typically run locally or in CI to sanity-check the database:

```bash
# 1) Start the app (AutoMigrate will create the table)
cd backend
go run .

# 2) Simple API smoke-check
curl -s http://localhost:8080/api/menus | jq .

# 3) Inspect DDL (MySQL/XAMPP)
mysql -uroot -p -e "USE sotekre_dev; SHOW CREATE TABLE menus\G"

# 4) Run tests that exercise DB logic
cd backend && go test ./... -v
```

## Schema & authoritative sources
- Migration (MySQL): `backend/migrations/001_create_menus.sql` — DDL and indexes
- Application model: `backend/models/menu.go` (GORM)
- Business logic & invariants: `backend/services/menu_service.go`

> [!NOTE]
> Always update both the migration and the GORM model when changing the schema; include tests that validate the new behavior.

## ERD (visual)
A compact Mermaid ERD is available here and in `backend/database/ERD.md`.

```mermaid
erDiagram
  MENUS {
    BIGINT id PK "auto-increment"
    VARCHAR title
    VARCHAR url
    BIGINT parent_id "nullable, self-reference"
    INT `order` "sibling position"
    DATETIME created_at
    DATETIME updated_at
    DATETIME deleted_at "soft delete"
  }

  MENUS ||--o{ MENUS : "parent -> children"
```

## Migrations & running locally
- Development (MVP): the server runs `AutoMigrate` on startup for rapid iteration.

```bash
# XAMPP/local MySQL (use backend/.env or rename backend/.env.example -> backend/.env)
cd backend
go run .
```

- Production: use an explicit migration runner (e.g. `golang-migrate`).

Migration workflow (recommended):
1. Add SQL under `backend/migrations/` with an incremental filename.
2. Update `backend/models/menu.go` to reflect the model.
3. Add tests (`backend/services` / `backend/handlers`) covering the behavior.

## Example queries (verification)
- Ordered root items:
```sql
SELECT id, title, parent_id, `order` FROM menus WHERE parent_id IS NULL ORDER BY `order`, id;
```

- Ordered children for parent = 1:
```sql
SELECT id, title, `order` FROM menus WHERE parent_id = 1 ORDER BY `order`, id;
```

- Recursive subtree (MySQL 8+):
```sql
WITH RECURSIVE subtree AS (
  SELECT * FROM menus WHERE id = 1
  UNION ALL
  SELECT m.* FROM menus m JOIN subtree s ON m.parent_id = s.id
)
SELECT * FROM subtree ORDER BY parent_id, `order`, id;
```

- Minimal seed for manual testing:
```sql
INSERT INTO menus (title, parent_id, `order`) VALUES
('A', NULL, 0), ('B', NULL, 1), ('A.1', 1, 0);
```

## Indexes, constraints & rationale
- Indexes: `idx_parent (parent_id)`, `idx_order (order)` — support fast sibling enumeration and parent-scoped scans.
- FK constraints: intentionally omitted to keep deletion/soft-delete semantics and test control in application code.
- Ordering: `order` is normalized inside a DB transaction by the service layer during move/reorder operations to ensure consistency.

## Testing & CI guidance
- Unit & integration tests use in-memory SQLite for fast, hermetic runs (`backend/*_test.go`).
- CI should:
  - run `go generate ./...` (if docs are generated),
  - boot a disposable DB (or use a service in CI),
  - apply migrations, and
  - run `go test ./... -v`.

Example (GitHub Actions snippet, high level):
```yaml
# - name: Run tests
#   run: |
#     docker run -d --name mysql -e MYSQL_ROOT_PASSWORD=$MYSQL_ROOT_PASSWORD mysql:8
#     # wait + run migrations + go test
```

## Reviewer checklist
- [ ] Model ↔ migration parity (`backend/models/menu.go` vs `backend/migrations/001_create_menus.sql`).
- [ ] Move/Reorder invariants: cycle prevention + transactional sibling reindexing (`backend/services/menu_service.go`).
- [ ] Recursive delete behavior and soft-delete semantics (`backend/handlers` integration tests).
- [ ] Indexes and query shapes for expected workloads (sibling enumeration).

---

## Reference files
- `backend/migrations/001_create_menus.sql`
- `backend/models/menu.go`
- `backend/services/menu_service.go`
- `backend/handlers/*` (integration tests)
