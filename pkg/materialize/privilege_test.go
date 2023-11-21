package materialize

import (
	"reflect"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/jmoiron/sqlx"
)

func TestPrivilegeName(t *testing.T) {
	p, err := PrivilegeName("r")
	if err != nil {
		t.Fatal(err)
	}
	if p != "SELECT" {
		t.Fatal("unexpected privilege name mapping")
	}
}

func TestParseMzCatalogPrivileges(t *testing.T) {
	o := ParseMzCatalogPrivileges("u18=arwd/s1")
	e := MzCatalogPrivilege{
		Grantee:    "u18",
		Privileges: []string{"INSERT", "SELECT", "UPDATE", "DELETE"},
		Grantor:    "s1",
	}
	if !reflect.DeepEqual(o, e) {
		t.Log(o)
		t.Log(e)
		t.Fatalf("could not parse into expected MzCatalogPrivilege")
	}
}

func TestMapGrantPrivileges(t *testing.T) {
	o, err := MapGrantPrivileges([]string{"s1=arwd/s1", "u3=wd/s1"})
	if err != nil {
		t.Fatal(err)
	}
	e := map[string][]string{
		"s1": {"INSERT", "SELECT", "UPDATE", "DELETE"},
		"u3": {"UPDATE", "DELETE"},
	}
	if !reflect.DeepEqual(o, e) {
		t.Log(o)
		t.Log(e)
		t.Fatalf("could not parse into expected mapping")
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

		b := NewPrivilegeBuilder(db, "joe", "CREATE", MaterializeObject{ObjectType: "DATABASE", Name: "materialize"})
		if err := b.Grant(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestPrivilegeRevoke(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE CREATE ON DATABASE "materialize" FROM "joe";`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewPrivilegeBuilder(db, "joe", "CREATE", MaterializeObject{ObjectType: "DATABASE", Name: "materialize"})
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

		o, err := ScanPrivileges(db, "DATABASE", "u1")
		if err != nil {
			t.Fatal(err)
		}

		e := []string{"s1=arwd/s1", "u1=UC/u18", "u8=arw/s1"}
		if !reflect.DeepEqual(o, e) {
			t.Fatalf("unexpected privileges %s", o)
		}
	})
}
