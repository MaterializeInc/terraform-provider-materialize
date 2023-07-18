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

func ParsePrivileges(privileges string) map[string][]string {
	o := map[string][]string{}

	privileges = strings.TrimPrefix(privileges, "{")
	privileges = strings.TrimSuffix(privileges, "}")

	for _, p := range strings.Split(privileges, ",") {
		e := strings.Split(p, "=")

		roleId := e[0]
		roleprivileges := strings.Split(e[1], "/")[0]

		privilegeMap := []string{}
		for _, rp := range strings.Split(roleprivileges, "") {
			v := Permissions[rp]
			privilegeMap = append(privilegeMap, v)
		}

		o[roleId] = privilegeMap
	}

	return o
}

func HasPrivilege(privileges []string, checkPrivilege string) bool {
	for _, v := range privileges {
		if v == checkPrivilege {
			return true
		}
	}
	return false
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
	object    ObjectSchemaStruct
}

func NewPrivilegeBuilder(conn *sqlx.DB, role, privilege string, object ObjectSchemaStruct) *PrivilegeBuilder {
	return &PrivilegeBuilder{
		ddl:       Builder{conn, Privilege},
		role:      MaterializeRole{name: role},
		privilege: privilege,
		object:    object,
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

func ScanPrivileges(conn *sqlx.DB, objectType, objectId string) (string, error) {
	var p string
	var e error

	switch t := objectType; t {
	case "DATABASE":
		params, err := ScanDatabase(conn, objectId)
		p = params.Privileges.String
		e = err

	case "SCHEMA":
		params, err := ScanSchema(conn, objectId)
		p = params.Privileges.String
		e = err

	case "TABLE":
		params, err := ScanTable(conn, objectId)
		p = params.Privileges.String
		e = err

	case "VIEW":
		params, err := ScanView(conn, objectId)
		p = params.Privileges.String
		e = err

	case "MATERIALIZED VIEW":
		params, err := ScanMaterializedView(conn, objectId)
		p = params.Privileges.String
		e = err

	case "TYPE":
		params, err := ScanType(conn, objectId)
		p = params.Privileges.String
		e = err

	case "SOURCE":
		params, err := ScanSource(conn, objectId)
		p = params.Privileges.String
		e = err

	case "CONNECTION":
		params, err := ScanConnection(conn, objectId)
		p = params.Privileges.String
		e = err

	case "SECRET":
		params, err := ScanSecret(conn, objectId)
		p = params.Privileges.String
		e = err

	case "CLUSTER":
		params, err := ScanCluster(conn, objectId)
		p = params.Privileges.String
		e = err
	}

	if e != nil {
		return "", e
	}

	return p, nil
}
