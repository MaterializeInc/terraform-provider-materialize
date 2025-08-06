package materialize

import (
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

func TestSourceReferenceId(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectQuery(
			`SELECT sr\.source_id, sr\.namespace, sr\.name, sr\.updated_at, sr\.columns, s\.name AS source_name, ss\.name AS source_schema_name, sd\.name AS source_database_name, s\.type AS source_type
			FROM mz_internal\.mz_source_references sr
			JOIN mz_sources s ON sr\.source_id = s\.id
			JOIN mz_schemas ss ON s\.schema_id = ss\.id
			JOIN mz_databases sd ON ss\.database_id = sd\.id
			WHERE sr\.source_id = 'test-source-id'`,
		).
			WillReturnRows(sqlmock.NewRows([]string{"source_id"}).AddRow("test-source-id"))

		result, err := SourceReferenceId(db, "test-source-id")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "test-source-id" {
			t.Errorf("expected source id to be 'test-source-id', got %v", result)
		}
	})
}

func TestScanSourceReference(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectQuery(
			`SELECT sr\.source_id, sr\.namespace, sr\.name, sr\.updated_at, sr\.columns, s\.name AS source_name, ss\.name AS source_schema_name, sd\.name AS source_database_name, s\.type AS source_type
			FROM mz_internal\.mz_source_references sr
			JOIN mz_sources s ON sr\.source_id = s\.id
			JOIN mz_schemas ss ON s\.schema_id = ss\.id
			JOIN mz_databases sd ON ss\.database_id = sd\.id
			WHERE sr\.source_id = 'test-source-id'`,
		).
			WillReturnRows(sqlmock.NewRows([]string{"source_id", "namespace", "name", "updated_at", "columns", "source_name", "source_schema_name", "source_database_name", "source_type"}).
				AddRow("test-source-id", "test-namespace", "test-name", "2024-10-28", pq.StringArray{"col1", "col2"}, "source-name", "source-schema-name", "source-database-name", "source-type"))

		result, err := ScanSourceReference(db, "test-source-id")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result.SourceId.String != "test-source-id" {
			t.Errorf("expected source id to be 'test-source-id', got %v", result.SourceId.String)
		}
	})
}

func TestRefreshSourceReferences(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(
			`ALTER SOURCE "test-database"\."test-schema"\."test-source" REFRESH REFERENCES`,
		).
			WillReturnResult(sqlmock.NewResult(1, 1))

		err := refreshSourceReferences(db, "test-source", "test-schema", "test-database")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})
}

func TestListSourceReferences(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectQuery(
			`SELECT sr\.source_id, sr\.namespace, sr\.name, sr\.updated_at, sr\.columns, s\.name AS source_name, ss\.name AS source_schema_name, sd\.name AS source_database_name, s\.type AS source_type
			FROM mz_internal\.mz_source_references sr
			JOIN mz_sources s ON sr\.source_id = s\.id
			JOIN mz_schemas ss ON s\.schema_id = ss\.id
			JOIN mz_databases sd ON ss\.database_id = sd\.id
			WHERE sr\.source_id = 'test-source-id'`,
		).
			WillReturnRows(sqlmock.NewRows([]string{"source_id", "namespace", "name", "updated_at", "columns", "source_name", "source_schema_name", "source_database_name", "source_type"}).
				AddRow("test-source-id", "test-namespace", "test-name", "2024-10-28", pq.StringArray{"col1", "col2"}, "source-name", "source-schema-name", "source-database-name", "source-type"))

		result, err := ListSourceReferences(db, "test-source-id")
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if len(result) != 1 {
			t.Errorf("expected 1 result, got %d", len(result))
		}
		if result[0].SourceId.String != "test-source-id" {
			t.Errorf("expected source id to be 'test-source-id', got %v", result[0].SourceId.String)
		}
	})
}
