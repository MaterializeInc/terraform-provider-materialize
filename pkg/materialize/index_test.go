package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jackc/pgx/v5/pgconn"
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

// exec flattens driver errors into one message; the SQLSTATE must survive it so
// callers can tell a dependency error apart from any other failure.
func TestIndexDropKeepsSQLState(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP INDEX "database"."schema"."index" RESTRICT;`).WillReturnError(&pgconn.PgError{
			Severity: "ERROR",
			Code:     "2BP01",
			Message:  `cannot drop index "index": still depended upon by materialized view "mv"`,
		})

		err := NewIndexBuilder(db, MaterializeObject{Name: "index"}, false, IdentifierSchemaStruct{Name: "source", SchemaName: "schema", DatabaseName: "database"}).Drop()
		if err == nil {
			t.Fatal("expected an error")
		}
		if got, want := err.Error(), `ERROR: cannot drop index "index": still depended upon by materialized view "mv" (SQLSTATE 2BP01)`; got != want {
			t.Fatalf("message changed:\n got %s\nwant %s", got, want)
		}
		if !IsDependentObjectsStillExist(err) {
			t.Fatal("expected the 2BP01 SQLSTATE to be detectable")
		}
	})
}
