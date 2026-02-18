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
	o := ParseMzAclString("u18=arwd/s1")
	e := MzAclItem{
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
	if objectCompatibility(Cluster) != Cluster {
		t.Fatal("expected cluster object compatibility to be 'CLUSTER")
	}

	if objectCompatibility(BaseSource) != Table {
		t.Fatal("expected source object compatibility to be 'TABLE")
	}
}

func TestGetObjectPermissions(t *testing.T) {
	// Test entity types with permissions
	tests := []struct {
		entityType EntityType
		expected   []string
	}{
		{Database, []string{"U", "C"}},
		{Schema, []string{"U", "C"}},
		{Table, []string{"a", "r", "w", "d"}},
		{View, []string{"r"}},
		{MaterializedView, []string{"r"}},
		{BaseType, []string{"U"}},
		{BaseSource, []string{"r"}},
		{BaseConnection, []string{"U"}},
		{Secret, []string{"U"}},
		{Cluster, []string{"U", "C"}},
		{System, []string{"R", "B", "N", "P"}},
	}

	for _, tt := range tests {
		t.Run(string(tt.entityType), func(t *testing.T) {
			got := GetObjectPermissions(tt.entityType)
			if !reflect.DeepEqual(got, tt.expected) {
				t.Errorf("GetObjectPermissions(%s) = %v, want %v", tt.entityType, got, tt.expected)
			}
		})
	}

	// Test entity types without permissions (should return empty slice)
	emptyPermissionTypes := []EntityType{Index, BaseSink, ClusterReplica, NetworkPolicy, Role, Privilege, Ownership}
	for _, et := range emptyPermissionTypes {
		t.Run(string(et)+"_empty", func(t *testing.T) {
			got := GetObjectPermissions(et)
			if len(got) != 0 {
				t.Errorf("GetObjectPermissions(%s) = %v, want empty slice", et, got)
			}
		})
	}
}

func TestPrivilegeGrant(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`GRANT CREATE ON DATABASE "materialize" TO "joe";`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewPrivilegeBuilder(db, "joe", "CREATE", MaterializeObject{ObjectType: Database, Name: "materialize"})
		if err := b.Grant(); err != nil {
			t.Fatal(err)
		}
	})
}

func TestPrivilegeRevoke(t *testing.T) {
	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`REVOKE CREATE ON DATABASE "materialize" FROM "joe";`).WillReturnResult(sqlmock.NewResult(1, 1))

		b := NewPrivilegeBuilder(db, "joe", "CREATE", MaterializeObject{ObjectType: Database, Name: "materialize"})
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

		o, err := ScanPrivileges(db, Database, "u1")
		if err != nil {
			t.Fatal(err)
		}

		e := []string{"s1=arwd/s1", "u1=UC/u18", "u8=arw/s1"}
		if !reflect.DeepEqual(o, e) {
			t.Fatalf("unexpected privileges %s", o)
		}
	})
}
