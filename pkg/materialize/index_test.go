package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

// https://materialize.com/docs/sql/create-index/

func TestIndexFieldCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE INDEX index IN CLUSTER cluster ON "database"."schema"."source" \(column\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "index"}
		b := NewIndexBuilder(db, o, false, IdentifierSchemaStruct{SchemaName: "schema", Name: "source", DatabaseName: "database"})
		b.ClusterName("cluster")
		b.ColExpr([]IndexColumn{
			{Field: "column"},
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestIndexFieldLiteralCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE INDEX index IN CLUSTER cluster ON "database"."schema"."source" \(upper\(guid\), geo_id\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "index"}
		b := NewIndexBuilder(db, o, false, IdentifierSchemaStruct{SchemaName: "schema", Name: "source", DatabaseName: "database"})
		b.ClusterName("cluster")
		b.ColExpr([]IndexColumn{
			{Field: "upper(guid)"},
			{Field: "geo_id"},
		})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestIndexDefaultCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE DEFAULT INDEX IN CLUSTER cluster ON "database"."schema"."source" USING ARRANGEMENT;`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "index"}
		b := NewIndexBuilder(db, o, true, IdentifierSchemaStruct{SchemaName: "schema", Name: "source", DatabaseName: "database"})
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

		o := MaterializeObject{Name: "index"}
		b := NewIndexBuilder(db, o, false, IdentifierSchemaStruct{SchemaName: "schema", Name: "source", DatabaseName: "database"})
		if err := b.Drop(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestIndexComment(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`COMMENT ON INDEX "database"."schema"."index" IS 'comment';`).WillReturnResult(sqlmock.NewResult(1, 1))

		o := MaterializeObject{Name: "index"}
		b := NewIndexBuilder(db, o, false, IdentifierSchemaStruct{SchemaName: "schema", Name: "source", DatabaseName: "database"})
		if err := b.Comment("comment"); err != nil {
			t.Fatal(err)
		}
	})
}
