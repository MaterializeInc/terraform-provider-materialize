package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

var sourceTable = MaterializeObject{Name: "table", SchemaName: "schema", DatabaseName: "database"}

func TestSourceTableId(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'table'`
		testhelpers.MockSourceTableScan(mock, ip)

		id, err := SourceTableId(db, sourceTable)
		if err != nil {
			t.Fatal(err)
		}
		if id != "u1" {
			t.Errorf("Expected id 'u1', got '%s'", id)
		}
	})
}

func TestScanSourceTable(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableScan(mock, pp)

		params, err := ScanSourceTable(db, "u1")
		if err != nil {
			t.Fatal(err)
		}
		if params.TableId.String != "u1" {
			t.Errorf("Expected table id 'u1', got '%s'", params.TableId.String)
		}
		if params.TableName.String != "table" {
			t.Errorf("Expected table name 'table', got '%s'", params.TableName.String)
		}
	})
}

func TestSourceTableBuilderQualifiedName(t *testing.T) {
	b := NewSourceTableBuilder(nil, sourceTable)
	expected := `"database"."schema"."table"`
	if b.QualifiedName() != expected {
		t.Errorf("Expected qualified name '%s', got '%s'", expected, b.QualifiedName())
	}
}

func TestSourceTableBuilderSource(t *testing.T) {
	b := NewSourceTableBuilder(nil, sourceTable)
	source := IdentifierSchemaStruct{Name: "source", SchemaName: "public", DatabaseName: "materialize"}
	b.Source(source)

	if b.source.Name != "source" {
		t.Errorf("Expected source name 'source', got '%s'", b.source.Name)
	}
}

func TestSourceTableBuilderUpstreamName(t *testing.T) {
	b := NewSourceTableBuilder(nil, sourceTable)
	b.UpstreamName("upstream_table")

	if b.upstreamName != "upstream_table" {
		t.Errorf("Expected upstream name 'upstream_table', got '%s'", b.upstreamName)
	}
}

func TestSourceTableBuilderUpstreamSchemaName(t *testing.T) {
	b := NewSourceTableBuilder(nil, sourceTable)
	b.UpstreamSchemaName("upstream_schema")

	if b.upstreamSchemaName != "upstream_schema" {
		t.Errorf("Expected upstream schema name 'upstream_schema', got '%s'", b.upstreamSchemaName)
	}
}

func TestSourceTableBuilderRename(t *testing.T) {
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

func TestSourceTableBuilderDrop(t *testing.T) {
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

func TestSourceTableBuilderBaseCreate(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table" FROM SOURCE "materialize"."public"."source" \(REFERENCE "upstream_schema"."upstream_table"\) WITH \(TEXT COLUMNS \("col1", "col2"\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTableBuilder(db, sourceTable)
		b.Source(IdentifierSchemaStruct{Name: "source", SchemaName: "public", DatabaseName: "materialize"})
		b.UpstreamName("upstream_table")
		b.UpstreamSchemaName("upstream_schema")

		err := b.BaseCreate("postgres", func() string {
			return ` WITH (TEXT COLUMNS ("col1", "col2"))`
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTableBuilderBaseCreateWithoutReference(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table" FROM SOURCE "materialize"."public"."source" WITH \(TEXT COLUMNS \("col1"\)\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTableBuilder(db, sourceTable)
		b.Source(IdentifierSchemaStruct{Name: "source", SchemaName: "public", DatabaseName: "materialize"})

		err := b.BaseCreate("kafka", func() string {
			return ` WITH (TEXT COLUMNS ("col1"))`
		})
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTableBuilderBaseCreateWithoutOptions(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."table" FROM SOURCE "materialize"."public"."source" \(REFERENCE "upstream_table"\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewSourceTableBuilder(db, sourceTable)
		b.Source(IdentifierSchemaStruct{Name: "source", SchemaName: "public", DatabaseName: "materialize"})
		b.UpstreamName("upstream_table")

		err := b.BaseCreate("postgres", nil)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestListSourceTables(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		predicate := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockSourceTableScan(mock, predicate)

		tables, err := ListSourceTables(db, "schema", "database")
		if err != nil {
			t.Fatal(err)
		}
		if len(tables) != 1 {
			t.Errorf("Expected 1 table, got %d", len(tables))
		}
		if tables[0].TableName.String != "table" {
			t.Errorf("Expected table name 'table', got '%s'", tables[0].TableName.String)
		}
	})
}
