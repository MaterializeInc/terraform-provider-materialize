package materialize

import (
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/stretchr/testify/require"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/jmoiron/sqlx"
)

func TestOwnershipResourceId(t *testing.T) {
	r := require.New(t)

	table := OwnershipResourceId("TABLE", "u1")
	r.Equal(`ownership|table|u1`, table)

	mview := OwnershipResourceId("MATERIALIZED VIEW", "u1")
	r.Equal("ownership|materialized_view|u1", mview)
}

func TestOwnershipCatalogId(t *testing.T) {
	r := require.New(t)

	table := OwnershipCatalogId("ownership|table|u1")
	r.Equal("u1", table)

	mview := OwnershipCatalogId("ownership|materialized_view|u1")
	r.Equal("u1", mview)
}

func TestOwnershipAlter(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER TABLE "database"."schema"."table" OWNER TO my_role;`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewOwnershipBuilder(db, "TABLE")
		b.RoleName("my_role")
		b.Object(ObjectSchemaStruct{
			DatabaseName: "database",
			SchemaName:   "schema",
			Name:         "table",
		})

		if err := b.Alter(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestOwnershipReadId(t *testing.T) {
	r := require.New(t)

	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		query := `
			SELECT o.id
			FROM mz_tables o
			JOIN mz_schemas
				ON o.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
			WHERE o.name = 'table'
			AND mz_databases.name = 'database'
			AND mz_schemas.name = 'schema'
		`
		ir := mock.NewRows([]string{"id"}).AddRow("u1")
		mock.ExpectQuery(query).WillReturnRows(ir)

		b := NewOwnershipBuilder(db, "TABLE")
		b.Object(ObjectSchemaStruct{
			DatabaseName: "database",
			SchemaName:   "schema",
			Name:         "table",
		})

		id, err := b.ReadId()

		if err != nil {
			t.Fatal(err)
		}

		r.Equal("ownership|table|u1", id)
	})
}

func TestOwnershipParams(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		query := `
			SELECT
				o.owner_id,
				r.name AS role_name
			FROM mz_tables o
			JOIN mz_roles r
				ON o.owner_id = r.id
			WHERE o.id = 'u1'
		`
		ir := mock.NewRows([]string{"owner_id", "role_name"}).AddRow("u1", "my_role")
		mock.ExpectQuery(query).WillReturnRows(ir)

		b := NewOwnershipBuilder(db, "TABLE")

		_, err := b.Params("u1")

		if err != nil {
			t.Fatal(err)
		}
	})
}
