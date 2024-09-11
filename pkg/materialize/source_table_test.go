package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var sourceTable = MaterializeObject{Name: "table", SchemaName: "schema", DatabaseName: "database"}

func TestSourceTableCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		sourceTypeQuery := `WHERE mz_databases.name = 'materialize' AND mz_schemas.name = 'public' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScanWithType(mock, sourceTypeQuery, "kafka")

		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table"
			FROM SOURCE "materialize"."public"."source"
			\(REFERENCE "upstream_schema"."upstream_table"\)
			WITH \(TEXT COLUMNS \(column1, column2\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTableBuilder(db, sourceTable)
		b.Source(IdentifierSchemaStruct{Name: "source", SchemaName: "public", DatabaseName: "materialize"})
		b.UpstreamName("upstream_table")
		b.UpstreamSchemaName("upstream_schema")
		b.TextColumns([]string{"column1", "column2"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTableCreateWithMySQLSource(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		sourceTypeQuery := `WHERE mz_databases.name = 'materialize' AND mz_schemas.name = 'public' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScanWithType(mock, sourceTypeQuery, "mysql")

		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table"
			FROM SOURCE "materialize"."public"."source"
			\(REFERENCE "upstream_schema"."upstream_table"\)
			WITH \(TEXT COLUMNS \(column1, column2\), IGNORE COLUMNS \(ignore1, ignore2\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTableBuilder(db, sourceTable)
		b.Source(IdentifierSchemaStruct{Name: "source", SchemaName: "public", DatabaseName: "materialize"})
		b.UpstreamName("upstream_table")
		b.UpstreamSchemaName("upstream_schema")
		b.TextColumns([]string{"column1", "column2"})
		b.IgnoreColumns([]string{"ignore1", "ignore2"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTableLoadgenCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		sourceTypeQuery := `WHERE mz_databases.name = 'materialize' AND mz_schemas.name = 'public' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScanWithType(mock, sourceTypeQuery, "load-generator")

		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table"
			FROM SOURCE "materialize"."public"."source"
			\(REFERENCE "upstream_schema"."upstream_table"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTableBuilder(db, sourceTable)
		b.Source(IdentifierSchemaStruct{Name: "source", SchemaName: "public", DatabaseName: "materialize"})
		b.UpstreamName("upstream_table")
		b.UpstreamSchemaName("upstream_schema")
		// Text columns are not supported for load-generator sources and should be ignored in the query builder
		b.TextColumns([]string{"column1", "column2"})

		if err := b.Create(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestGetSourceType(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		sourceTypeQuery := `WHERE mz_databases.name = 'materialize' AND mz_schemas.name = 'public' AND mz_sources.name = 'source'`
		testhelpers.MockSourceScanWithType(mock, sourceTypeQuery, "KAFKA")

		b := NewSourceTableBuilder(db, sourceTable)
		b.Source(IdentifierSchemaStruct{Name: "source", SchemaName: "public", DatabaseName: "materialize"})

		sourceType, err := b.GetSourceType()
		if err != nil {
			t.Fatal(err)
		}

		if sourceType != "KAFKA" {
			t.Fatalf("Expected source type 'KAFKA', got '%s'", sourceType)
		}
	})
}

func TestSourceTableRename(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER TABLE "database"."schema"."table" RENAME TO "database"."schema"."new_table";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTableBuilder(db, sourceTable)
		if err := b.Rename("new_table"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTableDrop(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`DROP TABLE "database"."schema"."table";`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTableBuilder(db, sourceTable)
		if err := b.Drop(); err != nil {
			t.Fatal(err)
		}
	})
}
