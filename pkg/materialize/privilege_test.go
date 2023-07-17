package materialize

import (
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestParsePrivileges(t *testing.T) {
	o := ParsePrivileges("{u2=r/u18,u3=r/u18,u18=arwd/u18}")
	e := map[string][]string{
		"u2":  {"SELECT"},
		"u3":  {"SELECT"},
		"u18": {"INSERT", "SELECT", "UPDATE", "DELETE"},
	}
	if !reflect.DeepEqual(o, e) {
		t.Fatalf("unexpected privilege mapping")
	}
}

func TestHasPrivilege(t *testing.T) {
	p := []string{"SELECT", "INSERT", "UPDATE"}

	if !HasPrivilege(p, "INSERT") {
		t.Fatalf("expected priviledge %s to contain 'INSERT'", p)
	}

	if HasPrivilege(p, "DELETE") {
		t.Fatalf("expected priviledge %s to not contain 'DELETE", p)
	}
}

func TestObjectCompatibility(t *testing.T) {
	if objectCompatibility("CLUSTER") != "CLUSTER" {
		t.Fatal("expected cluster object compatibility to be 'CLUSTER")
	}

	if objectCompatibility("SOURCE") != "TABLE" {
		t.Fatal("expected source object compatibility to be 'TABLE")
	}
}

func TestPrivilegeGrant(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`GRANT CREATE ON DATABASE "materialize" TO "joe";`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewPrivilegeBuilder(db, "joe", "CREATE", ObjectSchemaStruct{ObjectType: "DATABASE", Name: "materialize"})
		if err := b.Grant(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestPrivilegeRevoke(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE CREATE ON DATABASE "materialize" FROM "joe";`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewPrivilegeBuilder(db, "joe", "CREATE", ObjectSchemaStruct{ObjectType: "DATABASE", Name: "materialize"})
		if err := b.Revoke(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestScanPrivileges(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		// Query Id
		ip := `WHERE mz_databases.id = 'u1'`
		testhelpers.MockDatabaseScan(mock, ip)

		p, err := ScanPrivileges(db, "DATABASE", "u1")
		if err != nil {
			t.Fatal(err)
		}

		if p != "{u1=UC/u18}" {
			t.Fatalf("unexpected privileges %s", p)
		}
	})
}
