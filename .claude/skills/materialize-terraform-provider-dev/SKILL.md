---
name: materialize-terraform-provider-dev
description: >-
  Contributing to the Materialize Terraform provider codebase. Covers the
  project layout (pkg/materialize, pkg/resources, pkg/datasources,
  pkg/provider), the builder pattern for SQL generation, adding new
  resources and data sources end to end, the acceptance test framework,
  unit testing with sqlmock, CI pipelines, release flow via goreleaser,
  and codebase conventions. Use this skill whenever the user wants to add
  a new Terraform resource or data source, modify an existing resource,
  write or run provider tests, understand the provider architecture,
  debug test failures, cut a release, or contribute a fix to the
  terraform-provider-materialize codebase. Also trigger when the user
  mentions pkg/materialize, pkg/resources, acceptance tests, mock scans,
  the builder pattern in this provider, or GNUmakefile targets.
---

# Materialize Terraform Provider: Development Guide

This skill covers contributing to the `terraform-provider-materialize` codebase. For using the provider as an end user, see the `materialize-terraform-provider` skill instead.

## How to Use This Skill

1. **Adding a new resource**: Follow the end-to-end walkthrough in the "Adding a Resource" section.
2. **Understanding the architecture**: Read the "Architecture" section for the three-tier layout.
3. **Running tests**: See "Testing" for unit, acceptance, and integration test workflows.
4. **Debugging a test failure**: Check "Debugging" and the common gotchas.
5. **Cutting a release**: See "Release Flow".

Key reference files in the repo: `DESIGN.md` (architecture patterns), `CONTRIBUTING.md` (build/test instructions), `GNUmakefile` (build targets).

## Build Commands

| Command | Purpose |
|---------|---------|
| `make install` | Build and install provider to `~/.terraform.d/plugins/` |
| `make test` | Run unit tests |
| `make testacc` | Run acceptance tests (requires `docker compose up -d`) |
| `make docs` | Generate docs from schema (`go generate ./...`) |
| `make fmt` | Format Go and Terraform code |

## Architecture

Three-tier design. Each tier has a clear responsibility.

### Tier 1: Materialize layer (`pkg/materialize/`)

SQL builders and query definitions. About 107 files, one per object type. Every object follows the builder pattern (see `DESIGN.md`):

```go
// Builder: constructs SQL statements
type DatabaseBuilder struct {
    ddl          Builder
    databaseName string
}

func (b *DatabaseBuilder) Create() error {
    q := fmt.Sprintf(`CREATE DATABASE %s;`, QuoteIdentifier(b.databaseName))
    return b.ddl.exec(q)
}

func (b *DatabaseBuilder) Drop() error {
    return b.ddl.drop(QuoteIdentifier(b.databaseName))
}

// Query: reads from mz_catalog
type DatabaseParams struct {
    DatabaseId   sql.NullString `db:"id"`
    DatabaseName sql.NullString `db:"database_name"`
    OwnerName    sql.NullString `db:"owner_name"`
}

var databaseQuery = NewBaseQuery(`
    SELECT mz_databases.id, mz_databases.name AS database_name, ...
    FROM mz_databases
    JOIN mz_roles ON mz_databases.owner_id = mz_roles.id
`)
```

Key utilities:
- `QuoteIdentifier(s)`: double-quote SQL identifiers
- `QuoteString(s)`: single-quote SQL string literals
- `QualifiedName(parts...)`: join as `"db"."schema"."object"`
- `NewBaseQuery(sql)` + `.QueryPredicate(map)`: build filtered queries
- `MaterializeObject`: struct with `Name`, `SchemaName`, `DatabaseName`, `ObjectType`

### Tier 2: Resources and data sources (`pkg/resources/`, `pkg/datasources/`)

Terraform resource definitions. About 151 resource files and 55 data source files. Each resource has:

- Schema definition (map of `*schema.Schema`)
- Constructor returning `*schema.Resource` with CRUD functions
- CRUD functions: `<type>Create`, `<type>Read`, `<type>Update`, `<type>Delete`
- Optional import function

Schema helpers in `pkg/resources/schema.go`:
- `ObjectNameSchema(resource, required, forceNew)` for the `name` field
- `CommentSchema(required)`, `OwnershipRoleSchema()`, `RegionSchema()`
- `DatabaseNameSchema()`, `SchemaNameSchema()` for namespace fields
- `IdentifierSchema()` for object references
- `ValueSecretSchema()` for text-or-secret fields
- `FormatSpecSchema()` for complex format blocks (Avro, Protobuf, CSV)

### Tier 3: Provider registration (`pkg/provider/`)

`provider.go` registers all resources and data sources in `ResourcesMap` and `DataSourcesMap`. Also contains acceptance test infrastructure (`provider_test.go`).

### Supporting packages

- `pkg/clients/`: Database connection management (pgx/sqlx)
- `pkg/frontegg/`: Cloud/SaaS API client (Frontegg auth)
- `pkg/testhelpers/`: Mock database setup, shared mock scan functions
- `pkg/utils/`: ID extraction, region transforms, DB client helpers

## Adding a Resource End to End

Follow these steps when adding a new resource. The same pattern applies to adding fields to existing resources (just skip the registration step). See `CONTRIBUTING.md` for a concrete example using cluster `size`.

### Step 1: Builder in `pkg/materialize/`

Create `pkg/materialize/<type>.go`:

```go
type MyResourceBuilder struct {
    ddl  Builder
    name string
    // fields matching SQL parameters
}

func NewMyResourceBuilder(conn *sqlx.DB, obj MaterializeObject) *MyResourceBuilder {
    return &MyResourceBuilder{ddl: Builder{conn, MyResourceType}, name: obj.Name}
}

// Fluent setters
func (b *MyResourceBuilder) Size(s string) *MyResourceBuilder { b.size = s; return b }

// SQL generation
func (b *MyResourceBuilder) Create() error { ... }
func (b *MyResourceBuilder) Drop() error { ... }

// Query layer
type MyResourceParams struct { ... }
var myResourceQuery = NewBaseQuery(`SELECT ... FROM ...`)
func ScanMyResource(conn *sqlx.DB, id string) (MyResourceParams, error) { ... }
func MyResourceId(conn *sqlx.DB, obj MaterializeObject) (string, error) { ... }
```

Convention: parameter names should match Materialize SQL syntax as closely as possible (see `DESIGN.md`).

### Step 2: Resource in `pkg/resources/`

Create `pkg/resources/resource_<type>.go`:

```go
func MyResource() *schema.Resource {
    return &schema.Resource{
        CreateContext: myResourceCreate,
        ReadContext:   myResourceRead,
        UpdateContext: myResourceUpdate,
        DeleteContext: myResourceDelete,
        Importer: &schema.ResourceImporter{
            StateContext: schema.ImportStatePassthroughContext,
        },
        Schema: myResourceSchema,
    }
}
```

CRUD pattern:
- **Create**: Build object, execute SQL, query `mz_catalog` by name to get ID, call `d.SetId()` with region-prefixed ID, then call Read.
- **Read**: Scan by ID from `mz_catalog`. If `sql.ErrNoRows`, call `d.SetId("")` and return nil (resource was deleted externally). Otherwise populate all schema fields with `d.Set()`.
- **Update**: Check `d.HasChange("field")` for each mutable field. Apply changes via builder.
- **Delete**: Build and execute DROP.

ID management:
- After CREATE, query by name to get the `mz_catalog` ID
- Store as: `utils.TransformIdWithRegion(string(region), id)`
- Extract with: `utils.ExtractId(d.Id())`

### Step 3: Register in `pkg/provider/provider.go`

Add to `ResourcesMap`:
```go
"materialize_my_resource": resources.MyResource(),
```

### Step 4: Data source in `pkg/datasources/` (optional)

Create `pkg/datasources/datasource_<type>.go` with a Read function that lists objects and populates a TypeList schema.

Register in `DataSourcesMap`:
```go
"materialize_my_resource": datasources.MyResource(),
```

### Step 5: Mock scans in `pkg/testhelpers/mock_scans.go`

Add a `MockMyResourceScan` function that sets up `sqlmock` expectations matching your query.

### Step 6: Unit tests

Create `pkg/resources/resource_<type>_test.go` and optionally `pkg/materialize/<type>_test.go`.

Unit tests use `testhelpers.WithMockProviderMeta()` to set up a mock database:

```go
func TestResourceMyResourceCreate(t *testing.T) {
    in := map[string]interface{}{"name": "test_resource"}
    d := schema.TestResourceDataRaw(t, MyResource().Schema, in)

    testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
        mock.ExpectExec(`CREATE ...`).WillReturnResult(sqlmock.NewResult(1, 1))
        mock.ExpectQuery(`SELECT ...`).WillReturnRows(...)

        if err := myResourceCreate(context.TODO(), d, db); err != nil {
            t.Fatal(err)
        }
    })
}
```

Every `conn.Get()` or `conn.Exec()` in the code path needs a corresponding mock expectation. Missing mocks cause "unexpected query" test failures.

### Step 7: Acceptance tests

Create `pkg/provider/acceptance_<type>_test.go`:

```go
func TestAccMyResource_basic(t *testing.T) {
    name := acctest.RandStringFromCharSet(10, acctest.CharSetAlpha)

    resource.ParallelTest(t, resource.TestCase{
        PreCheck:          func() { testAccPreCheck(t) },
        ProviderFactories: testAccProviderFactories,
        CheckDestroy:      nil,
        Steps: []resource.TestStep{
            {
                Config: fmt.Sprintf(`resource "materialize_my_resource" "test" { name = "%s" }`, name),
                Check: resource.ComposeTestCheckFunc(
                    resource.TestCheckResourceAttr("materialize_my_resource.test", "name", name),
                ),
            },
            {
                ResourceName:      "materialize_my_resource.test",
                ImportState:       true,
                ImportStateVerify: true,
            },
        },
    })
}
```

Acceptance tests require `docker compose up -d` and `/etc/hosts` entries:
```
127.0.0.1 materialized frontegg cloud
```

### Step 8: Generate docs

```bash
make docs
```

This generates `docs/resources/<type>.md` and `docs/data-sources/<type>.md` from the schema.

## Testing

### Unit tests

```bash
make test
```

Run a specific test:
```bash
go test -v ./pkg/resources/ -run TestResourceMyResourceCreate
```

Coverage target: 50% minimum (enforced in CI). Unit tests should cover the breadth of SQL variations for each resource.

### Acceptance tests

```bash
docker compose up -d --build
make testacc
```

Run a specific acceptance test:
```bash
TF_ACC=1 go test -v ./pkg/provider/ -run TestAccMyResource_basic -timeout 1h
```

Acceptance tests run real Terraform commands against a Materialize instance in Docker. They should cover create, update, destroy, and import but do not need to exercise every SQL permutation.

### Integration tests

Full Terraform projects in `integration/` that apply and destroy all resources:

```bash
docker compose up -d
docker exec provider terraform init
docker exec provider terraform apply -auto-approve
docker exec provider terraform plan -detailed-exitcode
docker exec provider terraform destroy -auto-approve
```

For self-managed tests, use `--workdir /usr/src/app/integration/self_hosted`.

Clean up stale state before re-running: delete `.terraform/`, `.terraform.lock.hcl`, and `terraform.tfstate*` in the integration directory.

## Debugging

Enable Terraform debug logging:
```bash
TF_LOG=DEBUG terraform apply
```

Check database state during acceptance tests:
```bash
docker compose exec materialized psql -U materialize -d materialize -c "SELECT * FROM mz_tables LIMIT 5;"
```

## CI Pipelines

| Workflow | Trigger | What it does |
|----------|---------|-------------|
| `test.yml` | PR touching pkg/, main.go, go.mod | Unit tests with 50% coverage gate |
| `acceptance.yml` | PR | Acceptance tests against Docker Compose |
| `integration.yml` | PR | Full integration test (apply/plan/destroy) |
| `gofmt.yml` | PR | Go formatting check |
| `terraform.yml` | PR | Terraform HCL syntax check |
| `documentation.yml` | PR | Docs generation check |
| `release.yml` | Tag push (v*) | Goreleaser build, GPG sign, publish to GitHub Releases |

## Release Flow

Tag and push:
```bash
git tag -a vX.Y.Z -m vX.Y.Z
git push origin vX.Y.Z
```

This triggers `.goreleaser.yml` via GitHub Actions. Builds binaries for all platforms, signs with GPG, and publishes to GitHub Releases. The Terraform Registry picks up the release automatically.

## Codebase Conventions

**File naming:**
- Resources: `resource_<type>.go` + `resource_<type>_test.go` in `pkg/resources/`
- Data sources: `datasource_<type>.go` + `datasource_<type>_test.go` in `pkg/datasources/`
- Builders: `<type>.go` in `pkg/materialize/`
- Acceptance tests: `acceptance_<type>_test.go` in `pkg/provider/`
- Mock scans: `pkg/testhelpers/mock_scans.go`

**Resource naming:** Names must match Materialize SQL exactly. A load generator source is `materialize_source_load_generator` (matching `CREATE SOURCE ... FROM LOAD GENERATOR`). See `DESIGN.md`.

**Dividing resources:** Complex objects with many contradictory parameters get separate resources (e.g., `materialize_source_kafka` vs `materialize_source_postgres`), not one monolithic resource.

**SQL safety:** Always use `QuoteIdentifier()` for identifiers and `QuoteString()` for string literals. Never interpolate user input directly.

**Nested types:** Prefer typed structs over `map[string]interface{}` for builder parameters (see `DESIGN.md`).

**IDs from mz_catalog:** Resource IDs always come from `mz_catalog` system tables, never from SQL return values. After CREATE, query by name to get the ID.

**Ownership and comments:** Use the shared helpers `applyOwnership()` and `applyComment()` after resource creation.

**Region support:** Every resource gets a `region` schema field. Use `utils.GetDBClientFromMeta(meta, d)` which returns `(db, region, error)`.

## Common Gotchas

- **Missing mock expectations**: Every database call in the code path needs a mock. "Unexpected query" errors mean a mock is missing.
- **Forgetting region prefix on IDs**: Always wrap IDs with `utils.TransformIdWithRegion()` and extract with `utils.ExtractId()`.
- **Not handling sql.ErrNoRows**: In Read functions, check for `sql.ErrNoRows` and call `d.SetId("")` to mark the resource as deleted.
- **ForceNew on immutable fields**: Fields like `database_name` and `schema_name` that cannot be changed after creation need `ForceNew: true` in the schema.
- **Not updating mock_scans.go**: When you add or change query columns, the mock scan function must match.
- **Not updating the data source**: When adding a field to a resource, also add it to the corresponding data source schema and read function.
- **Pre-commit hooks**: Install with `pre-commit install`. They run `make docs`, `gofmt`, and `terraform fmt` automatically.

## Key Dependencies

| Package | Purpose |
|---------|---------|
| `hashicorp/terraform-plugin-sdk/v2` | Terraform plugin framework |
| `jackc/pgx/v4` | PostgreSQL driver |
| `jmoiron/sqlx` | SQL utility wrapper (named params, struct scanning) |
| `DATA-DOG/go-sqlmock` | Mock database for unit tests |
| `hashicorp/terraform-plugin-testing` | Acceptance test framework |
| `stretchr/testify` | Test assertions |
| `golang-jwt/jwt/v5` | JWT for cloud API auth |
| `hashicorp/terraform-plugin-docs` | Doc generation from schema |
