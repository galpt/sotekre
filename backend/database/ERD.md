# ERD — Menu Tree (compact)

This diagram and notes are intended for reviewers who want a quick, accurate view of the schema and runtime expectations.

## Mermaid ERD
```mermaid
erDiagram
  MENUS {
    BIGINT id PK "auto-increment"
    VARCHAR title "visible label"
    VARCHAR url "optional"
    VARCHAR icon "optional"
    BIGINT parent_id "self reference (nullable)"
    INT `order` "sibling position"
    DATETIME created_at
    DATETIME updated_at
    DATETIME deleted_at "soft delete"
  }

  MENUS ||--o{ MENUS : "parent -> children"
```

## Key points
- Adjacency‑list model (single table) — simple and easy to reason about for CRUD and reorder operations.
- No DB-enforced FK: application logic enforces deletion/move invariants and prevents cycles.
- Sibling ordering is stable and enforced in the service layer inside transactions.

## Migration / DDL
Authoritative DDL: `backend/migrations/001_create_menus.sql` (contains indexes used in queries and tests).

## Example verification queries
- Ordered root list:
```sql
SELECT id,title,`order` FROM menus WHERE parent_id IS NULL ORDER BY `order`, id;
```

- Ordered children for parent 1:
```sql
SELECT id,title,`order` FROM menus WHERE parent_id = 1 ORDER BY `order`, id;
```

- Recursive subtree (MySQL 8+):
```sql
WITH RECURSIVE tree AS (
  SELECT * FROM menus WHERE id = 1
  UNION ALL
  SELECT m.* FROM menus m JOIN tree t ON m.parent_id = t.id
)
SELECT * FROM tree ORDER BY parent_id, `order`, id;
```

## Performance & indexes
- Indexes: `idx_parent(parent_id)`, `idx_order(order)` — support fast sibling enumeration and scanning by parent.
- For very large trees, consider materialized path or closure table patterns; adjacency list is chosen here for simplicity and interview-readability.

## Changing the schema
1. Add a new migration SQL under `backend/migrations/` with an incremental filename.
2. Update `backend/models/menu.go` and add/update tests under `backend/services` and `backend/handlers`.
3. Run `go test ./...` and include a migration check in CI.

---

Files
- `backend/migrations/001_create_menus.sql`
- `backend/models/menu.go`
- `backend/services/menu_service.go`
```