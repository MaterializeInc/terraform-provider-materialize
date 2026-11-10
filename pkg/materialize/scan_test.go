package materialize_test

// These tests use the shared catalog mocks in testhelpers, which import this
// package, so they live in the external test package.

import (
	"reflect"
	"testing"

	"github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestScanConnectionAws(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		expectedExternalId := "mz_12345678-1234-1234-1234-123456789012_u123"

		// Mock the scan query response
		pp := `WHERE mz_connections.id = 'u1'`
		testhelpers.MockConnectionAwsScan(mock, pp)

		// Test materialize.ScanConnectionAws
		params, err := materialize.ScanConnectionAws(db, "u1")
		if err != nil {
			t.Fatal(err)
		}

		if params.ExternalId.String != expectedExternalId {
			t.Fatalf("Expected external_id to be %s, got %s", expectedExternalId, params.ExternalId.String)
		}

		if params.AssumeRoleArn.String != "arn:aws:iam::123456789012:user/JohnDoe" {
			t.Fatalf("Expected assume_role_arn to be %s, got %s", "arn:aws:iam::123456789012:user/JohnDoe", params.AssumeRoleArn.String)
		}
	})
}

func TestObjectId(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		o := materialize.MaterializeObject{ObjectType: materialize.Database, Name: "materialize"}

		// Query Id
		ip := `WHERE mz_databases.name = 'materialize'`
		testhelpers.MockDatabaseScan(mock, ip)

		_, err := materialize.ObjectId(db, o)
		if err != nil {
			t.Fatal(err)
		}
	})
}

func TestScanSystemPrivileges(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		testhelpers.MockSystemPrivilege(mock)

		p, err := materialize.ScanSystemPrivileges(db)
		if err != nil {
			t.Fatal(err)
		}

		if p[0].Privileges != "s1=RBN/s1" {
			t.Fatalf("unexpected privileges")
		}
	})
}

func TestScanPrivileges(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Query Id
		ip := `WHERE mz_databases.id = 'u1'`
		testhelpers.MockDatabaseScan(mock, ip)

		o, err := materialize.ScanPrivileges(db, materialize.Database, "u1")
		if err != nil {
			t.Fatal(err)
		}

		e := []string{"s1=arwd/s1", "u1=UC/u18", "u8=arw/s1", "=UC/s1"}
		if !reflect.DeepEqual(o, e) {
			t.Fatalf("unexpected privileges %s", o)
		}
	})
}

func TestListRolesWithoutPattern(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		testhelpers.MockRoleScan(mock, "")

		if _, err := materialize.ListRoles(db, ""); err != nil {
			t.Fatal(err)
		}
	})
}

func TestListRolesWithPattern(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		testhelpers.MockRoleScan(mock, "WHERE mz_roles.name LIKE 'prod_%'")

		if _, err := materialize.ListRoles(db, "prod_%"); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTableId(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'table'`
		testhelpers.MockSourceTableScan(mock, ip)

		id, err := materialize.SourceTableId(db, materialize.MaterializeObject{Name: "table", SchemaName: "schema", DatabaseName: "database"})
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

		params, err := materialize.ScanSourceTable(db, "u1")
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

func TestListSourceTables(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		predicate := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema'`
		testhelpers.MockSourceTableScan(mock, predicate)

		tables, err := materialize.ListSourceTables(db, "schema", "database")
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
