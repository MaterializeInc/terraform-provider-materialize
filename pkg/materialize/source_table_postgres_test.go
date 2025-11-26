package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var sourceTablePostgres = MaterializeObject{Name: "table", SchemaName: "schema", DatabaseName: "database"}

func TestSourceTablePostgresCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table"
			FROM SOURCE "materialize"."public"."source"
			\(REFERENCE "upstream_schema"."upstream_table"\)
			WITH \(TEXT COLUMNS \("column1", "column2"\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTablePostgresBuilder(db, sourceTablePostgres)
		b.Source(IdentifierSchemaStruct{Name: "source", SchemaName: "public", DatabaseName: "materialize"})
		b.UpstreamName("upstream_table")
		b.UpstreamSchemaName("upstream_schema")
		b.TextColumns([]string{"column1", "column2"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTablePostgresRename(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER TABLE "database"."schema"."table" RENAME TO "database"."schema"."new_table";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTablePostgresBuilder(db, sourceTablePostgres)
		if err := b.Rename("new_table"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTablePostgresCreateWithExcludeColumns(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table"
			FROM SOURCE "materialize"."public"."source"
			\(REFERENCE "upstream_schema"."upstream_table"\)
			WITH \(EXCLUDE COLUMNS \("exclude1", "exclude2"\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTablePostgresBuilder(db, sourceTablePostgres)
		b.Source(IdentifierSchemaStruct{Name: "source", SchemaName: "public", DatabaseName: "materialize"})
		b.UpstreamName("upstream_table")
		b.UpstreamSchemaName("upstream_schema")
		b.ExcludeColumns([]string{"exclude1", "exclude2"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTablePostgresCreateWithTextAndExcludeColumns(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table"
			FROM SOURCE "materialize"."public"."source"
			\(REFERENCE "upstream_schema"."upstream_table"\)
			WITH \(TEXT COLUMNS \("column1", "column2"\), EXCLUDE COLUMNS \("exclude1", "exclude2"\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTablePostgresBuilder(db, sourceTablePostgres)
		b.Source(IdentifierSchemaStruct{Name: "source", SchemaName: "public", DatabaseName: "materialize"})
		b.UpstreamName("upstream_table")
		b.UpstreamSchemaName("upstream_schema")
		b.TextColumns([]string{"column1", "column2"})
		b.ExcludeColumns([]string{"exclude1", "exclude2"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTablePostgresDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP TABLE "database"."schema"."table";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTablePostgresBuilder(db, sourceTablePostgres)
		if err := b.Drop(); err != nil {
			t.Fatal(err)
		}
	})
}
