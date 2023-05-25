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
