package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestIndexCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE INDEX index IN CLUSTER cluster ON "database"."schema"."source" USING ARRANGEMENT (Column LONG);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewIndexBuilder(db, "index", false, IdentifierSchemaStruct{SchemaName: "schema", Name: "source", DatabaseName: "database"})
		b.ClusterName("cluster")
		b.Method("ARRANGEMENT")
		b.ColExpr([]IndexColumn{
			{
				Field: "Column",
				Val:   "LONG",
			},
		})

		b.Create()
	})
}

func TestIndexDefaultCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE DEFAULT INDEX IN CLUSTER cluster ON "database"."schema"."source" USING ARRANGEMENT ();`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewIndexBuilder(db, "", true, IdentifierSchemaStruct{SchemaName: "schema", Name: "source", DatabaseName: "database"})
		b.ClusterName("cluster")
		b.Method("ARRANGEMENT")

		b.Create()
	})

}
