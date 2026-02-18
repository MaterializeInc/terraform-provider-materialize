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
	"P": "CREATENETWORKPOLICY",
}

// GetObjectPermissions returns the allowed permissions for a given entity type.
// This function uses an exhaustive switch to ensure all entity types are handled at compile time.
// https://materialize.com/docs/sql/grant-privilege/#details
func GetObjectPermissions(entityType EntityType) []string {
	switch entityType {
	case Database:
		return []string{"U", "C"}
	case Schema:
		return []string{"U", "C"}
	case Table:
		return []string{"a", "r", "w", "d"}
	case View:
		return []string{"r"}
	case MaterializedView:
		return []string{"r"}
	case Index:
		return []string{}
	case BaseType:
		return []string{"U"}
	case BaseSource:
		return []string{"r"}
	case BaseSink:
		return []string{}
	case BaseConnection:
		return []string{"U"}
	case Secret:
		return []string{"U"}
	case Cluster:
		return []string{"U", "C"}
	case ClusterReplica:
		return []string{}
	case System:
		return []string{"R", "B", "N", "P"}
	case NetworkPolicy:
		return []string{}
	case Role:
		return []string{}
	case Privilege:
		return []string{}
	case Ownership:
		return []string{}
	default:
		return []string{}
	}
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

type MzAclItem struct {
	Grantee    string
	Privileges []string
	Grantor    string
}

// https://materialize.com/docs/sql/types/mz_aclitem/
// Converts a mz catalog privilege string into a struct
// "s1=arwd/s1" would become
//
//	var x = MzAclItem{
//		Grantee:    "s1",
//		Privileges: ["INSERT", "SELECT", "UPDATE", "DELETE"],
//		Grantor:    "s1",
//	}
func ParseMzAclString(aclString string) MzAclItem {
	splitEqual := strings.Split(aclString, "=")
	splitSlash := strings.Split(splitEqual[1], "/")

	var parsedPrivileges = []string{}
	for _, p := range strings.Split(splitSlash[0], "") {
		pName, _ := PrivilegeName(p)
		parsedPrivileges = append(parsedPrivileges, pName)
	}

	return MzAclItem{
		Grantee:    splitEqual[0],
		Privileges: parsedPrivileges,
		Grantor:    splitSlash[1],
	}
}

// Converts a list of MZ ACL item strings into a map
// {"s1=arwd/s1", "u3=wd/s1"} would become
// map[string][]string
//
//	{
//		"s1": ["INSERT", "SELECT", "UPDATE", "DELETE"]
//		"u3": ["UPDATE", "DELETE"]
//	}
func MapGrantPrivileges(privileges []string) (map[string][]string, error) {
	mapping := make(map[string][]string)
	for _, p := range privileges {
		f := ParseMzAclString(p)
		mapping[f.Grantee] = f.Privileges
	}
	return mapping, nil
}

// DDL
type MaterializeRole struct {
	name string
}

func (b *MaterializeRole) QualifiedName() string {
	// If role name is PUBLIC, it should not be quoted as it is a pseudo-role
	if b.name == "PUBLIC" {
		return b.name
	}
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
func objectCompatibility(objectType EntityType) EntityType {
	switch objectType {
	case BaseSource, View, MaterializedView:
		return Table
	default:
		return objectType
	}
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

func (b *PrivilegeBuilder) GrantKey(region, objectId, roleId, privilege string) string {
	return fmt.Sprintf(`%[1]s:GRANT|%[2]s|%[3]s|%[4]s|%[5]s`, region, b.object.ObjectType, objectId, roleId, privilege)
}

func ScanPrivileges(conn *sqlx.DB, objectType EntityType, objectId string) ([]string, error) {
	var p []string
	var e error

	switch objectType {
	case Database:
		params, err := ScanDatabase(conn, objectId)
		p = params.Privileges
		e = err

	case Schema:
		params, err := ScanSchema(conn, objectId, false)
		p = params.Privileges
		e = err

	case Table:
		params, err := ScanTable(conn, objectId)
		p = params.Privileges
		e = err

	case View:
		params, err := ScanView(conn, objectId)
		p = params.Privileges
		e = err

	case MaterializedView:
		params, err := ScanMaterializedView(conn, objectId)
		p = params.Privileges
		e = err

	case BaseType:
		params, err := ScanType(conn, objectId)
		p = params.Privileges
		e = err

	case BaseSource:
		params, err := ScanSource(conn, objectId)
		p = params.Privileges
		e = err

	case BaseConnection:
		params, err := ScanConnection(conn, objectId)
		p = params.Privileges
		e = err

	case Secret:
		params, err := ScanSecret(conn, objectId)
		p = params.Privileges
		e = err

	case Cluster:
		params, err := ScanCluster(conn, objectId, false)
		p = params.Privileges
		e = err

	case NetworkPolicy:
		params, err := ScanNetworkPolicy(conn, objectId)
		p = params.Privileges
		e = err
	}

	if e != nil {
		return []string{}, e
	}

	return p, nil
}
