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
			`CREATE INDEX index IN CLUSTER cluster ON "database"."schema"."source" USING ARRANGEMENT \(column\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewIndexBuilder(db, "index", false, IdentifierSchemaStruct{SchemaName: "schema", Name: "source", DatabaseName: "database"})
		b.ClusterName("cluster")
		b.Method("ARRANGEMENT")
		b.ColExpr([]IndexColumn{
			{
				Field: "column",
			},
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestIndexDefaultCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE DEFAULT INDEX IN CLUSTER cluster ON "database"."schema"."source" USING ARRANGEMENT \(\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewIndexBuilder(db, "", true, IdentifierSchemaStruct{SchemaName: "schema", Name: "source", DatabaseName: "database"})
		b.ClusterName("cluster")
		b.Method("ARRANGEMENT")

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestIndexDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP INDEX "database"."schema"."index" RESTRICT;`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewIndexBuilder(db, "index", false, IdentifierSchemaStruct{SchemaName: "schema", Name: "source", DatabaseName: "database"})
		if err := b.Drop(); err != nil {
			t.Fatal(err)
		}
	})
}
