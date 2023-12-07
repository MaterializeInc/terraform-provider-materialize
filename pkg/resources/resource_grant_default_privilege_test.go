package resources

import (
	"testing"
)

func TestParseDefaultPrivilegeId(t *testing.T) {
	i, err := parseDefaultPrivilegeKey("aws/us-east-1:GRANT DEFAULT|TABLE|u1||||SELECT")
	if err != nil {
		t.Fatal(err)
	}
	if i.objectType != "TABLE" {
		t.Fatalf("object type %s: does not match expected TABLE", i.objectType)
	}
	if i.granteeId != "u1" {
		t.Fatalf("grantee id %s: does not match expected u1", i.granteeId)
	}
	if i.targetRoleId != "" {
		t.Fatalf("role id %s: expected to be empty string", i.targetRoleId)
	}
	if i.databaseId != "" {
		t.Fatalf("database id %s: expected to be empty string", i.databaseId)
	}
	if i.schemaId != "" {
		t.Fatalf("schema id %s: expected to be empty string", i.schemaId)
	}
}

func TestParseDefaultPrivilegeIdComplex(t *testing.T) {
	i, err := parseDefaultPrivilegeKey("aws/us-east-1:GRANT DEFAULT|TABLE|u1|u2|u3|u4|SELECT")
	if err != nil {
		t.Fatal(err)
	}
	if i.objectType != "TABLE" {
		t.Fatalf("object type %s: does not match expected TABLE", i.objectType)
	}
	if i.granteeId != "u1" {
		t.Fatalf("grantee id %s: does not match expected u1", i.granteeId)
	}
	if i.targetRoleId != "u2" {
		t.Fatalf("role id %s: expected u2", i.targetRoleId)
	}
	if i.databaseId != "u3" {
		t.Fatalf("database id %s: expected u3", i.databaseId)
	}
	if i.schemaId != "u4" {
		t.Fatalf("schema id %s: expected to u4", i.schemaId)
	}
}

func TestParseDefaultPrivilegeIdError(t *testing.T) {
	_, err := parseDefaultPrivilegeKey("incorrect|key|string")
	if err.Error() != "incorrect|key|string: cannot be parsed correctly" {
		t.Fatalf("error not handled")
	}
}

func TestParseDefaultPrivilegeIdErrorEmpty(t *testing.T) {
	_, err := parseDefaultPrivilegeKey("")
	if err.Error() != ": cannot be parsed correctly" {
		t.Fatal(err)
	}
}
