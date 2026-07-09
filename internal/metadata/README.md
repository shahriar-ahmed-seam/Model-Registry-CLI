# Metadata Store Implementation

## Files Created

### `schema.go`
Contains the DDL for the `version_records` table exactly as specified in the design document:
- `tag` TEXT PRIMARY KEY
- `content_hash` TEXT NOT NULL  
- `storage_key` TEXT NOT NULL
- `accuracy` REAL (nullable)
- `size_bytes` INTEGER NOT NULL
- `created_at` TEXT NOT NULL (RFC3339 UTC)

Also creates an index on `content_hash` for deduplication checks.

### `sqlite.go`
Implements the `MetadataStore` interface using the `modernc.org/sqlite` driver (CGo-free). Includes:

- `NewSQLiteStore(path string)` - Opens/Creates SQLite database
- `InitSchema()` - Creates table and indices
- `Insert(VersionRecord)` - Maps PK conflict to `registry.ErrTagExists`
- `GetByTag(string)` - Maps no rows to `registry.ErrTagNotFound`
- `List()` - Returns records ordered by `created_at` (oldest first)
- `Close()` - Releases database connection

## Error Mapping

- Primary key (tag) conflict → `registry.ErrTagExists`
- Tag not found → `registry.ErrTagNotFound`
- Other SQL errors → wrapped with context

## Design Compliance

The implementation satisfies all referenced requirements:
- 1.2: Metadata store schema creation
- 2.4: Version record insertion  
- 2.5: Accuracy storage (nullable REAL)
- 2.7: Duplicate tag error handling
- 3.1: Tag lookup
- 3.5: Tag not found error handling
- 4.1: List all records

## Note on Dependencies

The `modernc.org/sqlite` dependency is specified in `go.mod`. There may be temporary network issues with the modernc.org repository. If dependency resolution fails, try:
- Using a proxy: `GOPROXY=https://proxy.golang.org`
- Checking network connectivity
- Using an alternative version if the specified one is unavailable