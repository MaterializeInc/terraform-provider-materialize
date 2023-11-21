package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

var Permissions = map[string]string{
	"r": "SELECT",
	"a": "INSERT",
	"w": "UPDATE",
	"d": "DELETE",
	"C": "CREATE",
	"U": "USAGE",
	"R": "CREATEROLE",
	"B": "CREATEDB",
	"N": "CREATECLUSTER",
}

type ObjectType struct {
	Permissions []string
}

// https://materialize.com/docs/sql/grant-privilege/#details
var ObjectPermissions = map[string]ObjectType{
	"DATABASE": {
		Permissions: []string{"U", "C"},
	},
	"SCHEMA": {
		Permissions: []string{"U", "C"},
	},
	"TABLE": {
		Permissions: []string{"a", "r", "w", "d"},
	},
	"VIEW": {
		Permissions: []string{"r"},
	},
	"MATERIALIZED VIEW": {
		Permissions: []string{"r"},
	},
	"INDEX": {
		Permissions: []string{},
	},
	"TYPE": {
		Permissions: []string{"U"},
	},
	"SOURCE": {
		Permissions: []string{"r"},
	},
	"SINK": {
		Permissions: []string{},
	},
	"CONNECTION": {
		Permissions: []string{"U"},
	},
	"SECRET": {
		Permissions: []string{"U"},
	},
	"CLUSTER": {
		Permissions: []string{"U", "C"},
	},
	"SYSTEM": {
		Permissions: []string{"R", "B", "N"},
	},
}

// Converts a privilege abbrevation to name
// "a" would become INSERT
func PrivilegeName(privilegeAbbreviation string) (string, error) {
	val, ok := Permissions[privilegeAbbreviation]
	if ok {
		return val, nil
	}
	return "", fmt.Errorf("%s is not a valid privilege", privilegeAbbreviation)
}

type MzCatalogPrivilege struct {
	Grantee    string
	Privileges []string
	Grantor    string
}

// Converts a mz catalog privilege string into a struct
// "s1=arwd/s1" would become
//
//	var x = MzCatalogPrivilege{
//		Grantee:    "s1",
//		Privileges: ["INSERT", "SELECT", "UPDATE", "DELETE"],
//		Grantor:    "s1",
//	}
func ParseMzCatalogPrivileges(mzCatalogPrivilegeString string) MzCatalogPrivilege {
	splitEqual := strings.Split(mzCatalogPrivilegeString, "=")
	splitSlash := strings.Split(splitEqual[1], "/")

	var parsedPrivileges = []string{}
	for _, p := range strings.Split(splitSlash[0], "") {
		pName, _ := PrivilegeName(p)
		parsedPrivileges = append(parsedPrivileges, pName)
	}

	return MzCatalogPrivilege{
		Grantee:    splitEqual[0],
		Privileges: parsedPrivileges,
		Grantor:    splitSlash[1],
	}
}

// Converts a list of catalog privileges into a map for easy access
// {s1=arwd/s1,u3=wd/s1} would become
// map[string][]string
//
//	{
//		"s1": ["INSERT", "SELECT", "UPDATE", "DELETE"]
//		"u3": ["UPDATE", "DELETE"]
//	}
func MapGrantPrivileges(privileges []string) (map[string][]string, error) {
	mapping := make(map[string][]string)
	for _, p := range privileges {
		f := ParseMzCatalogPrivileges(p)
		mapping[f.Grantee] = f.Privileges
	}
	return mapping, nil
}

// DDL
type MaterializeRole struct {
	name string
}

func (b *MaterializeRole) QualifiedName() string {
	return QualifiedName(b.name)
}

type PrivilegeBuilder struct {
	ddl       Builder
	role      MaterializeRole
	privilege string
	object    MaterializeObject
}

func NewPrivilegeBuilder(conn *sqlx.DB, role, privilege string, obj MaterializeObject) *PrivilegeBuilder {
	return &PrivilegeBuilder{
		ddl:       Builder{conn, Privilege},
		role:      MaterializeRole{name: role},
		privilege: privilege,
		object:    obj,
	}
}

// https://materialize.com/docs/sql/grant-privilege/#compatibility
func objectCompatibility(objectType string) string {
	compatibility := []string{"SOURCE", "VIEW", "MATERIALIZED VIEW"}

	for _, c := range compatibility {
		if c == objectType {
			return "TABLE"
		}
	}
	return objectType
}

func (b *PrivilegeBuilder) Grant() error {
	t := objectCompatibility(b.object.ObjectType)
	q := fmt.Sprintf(`GRANT %s ON %s %s TO %s;`, b.privilege, t, b.object.QualifiedName(), b.role.QualifiedName())
	return b.ddl.exec(q)
}

func (b *PrivilegeBuilder) Revoke() error {
	t := objectCompatibility(b.object.ObjectType)
	q := fmt.Sprintf(`REVOKE %s ON %s %s FROM %s;`, b.privilege, t, b.object.QualifiedName(), b.role.QualifiedName())
	return b.ddl.exec(q)
}

func (b *PrivilegeBuilder) GrantKey(objectId, roleId, privilege string) string {
	return fmt.Sprintf(`GRANT|%[1]s|%[2]s|%[3]s|%[4]s`, b.object.ObjectType, objectId, roleId, privilege)
}

func ScanPrivileges(conn *sqlx.DB, objectType, objectId string) ([]string, error) {
	var p []string
	var e error

	switch t := objectType; t {
	case "DATABASE":
		params, err := ScanDatabase(conn, objectId)
		p = params.Privileges
		e = err

	case "SCHEMA":
		params, err := ScanSchema(conn, objectId)
		p = params.Privileges
		e = err

	case "TABLE":
		params, err := ScanTable(conn, objectId)
		p = params.Privileges
		e = err

	case "VIEW":
		params, err := ScanView(conn, objectId)
		p = params.Privileges
		e = err

	case "MATERIALIZED VIEW":
		params, err := ScanMaterializedView(conn, objectId)
		p = params.Privileges
		e = err

	case "TYPE":
		params, err := ScanType(conn, objectId)
		p = params.Privileges
		e = err

	case "SOURCE":
		params, err := ScanSource(conn, objectId)
		p = params.Privileges
		e = err

	case "CONNECTION":
		params, err := ScanConnection(conn, objectId)
		p = params.Privileges
		e = err

	case "SECRET":
		params, err := ScanSecret(conn, objectId)
		p = params.Privileges
		e = err

	case "CLUSTER":
		params, err := ScanCluster(conn, objectId)
		p = params.Privileges
		e = err
	}

	if e != nil {
		return []string{}, e
	}

	return p, nil
}
