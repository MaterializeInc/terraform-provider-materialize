package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func ParsePriviledges(priviledges string) map[string][]string {
	o := map[string][]string{}

	priviledges = strings.TrimPrefix(priviledges, "{")
	priviledges = strings.TrimSuffix(priviledges, "}")

	for _, p := range strings.Split(priviledges, ",") {
		e := strings.Split(p, "=")

		roleId := e[0]
		rolePriviledges := strings.Split(e[1], "/")[0]

		priviledgeMap := []string{}
		for _, rp := range strings.Split(rolePriviledges, "") {
			v := Permissions[rp]
			priviledgeMap = append(priviledgeMap, v)
		}

		o[roleId] = priviledgeMap
	}

	return o
}

func HasPriviledge(priviledges []string, checkPriviledge string) bool {
	for _, v := range priviledges {
		if v == checkPriviledge {
			return true
		}
	}
	return false
}

type PriviledgeObjectStruct struct {
	Type         string
	Name         string
	SchemaName   string
	DatabaseName string
}

func GetPriviledgeObjectStruct(databaseName string, schemaName string, v interface{}) PriviledgeObjectStruct {
	var p PriviledgeObjectStruct
	u := v.([]interface{})[0].(map[string]interface{})

	if v, ok := u["type"]; ok {
		p.Type = v.(string)
	}

	if v, ok := u["name"]; ok {
		p.Name = v.(string)
	}

	if v, ok := u["schema_name"]; ok && v.(string) != "" {
		p.SchemaName = v.(string)
	}

	if v, ok := u["database_name"]; ok && v.(string) != "" {
		p.DatabaseName = v.(string)
	}

	return p
}

func (i *PriviledgeObjectStruct) QualifiedName() string {
	p := []string{}

	if i.DatabaseName != "" {
		p = append(p, i.DatabaseName)
	}

	if i.SchemaName != "" {
		p = append(p, i.SchemaName)
	}

	p = append(p, i.Name)
	return QualifiedName(p...)
}

// DDL
type PrivilegeBuilder struct {
	ddl        Builder
	role       string
	priviledge string
	object     PriviledgeObjectStruct
}

func NewPrivilegeBuilder(conn *sqlx.DB, role, priviledge string, object PriviledgeObjectStruct) *PrivilegeBuilder {
	return &PrivilegeBuilder{
		ddl:        Builder{conn, Privilege},
		role:       role,
		priviledge: priviledge,
		object:     object,
	}
}

func (b *PrivilegeBuilder) Grant() error {
	q := fmt.Sprintf(`GRANT %s ON %s %s TO %s;`, b.priviledge, b.object.Type, b.object.QualifiedName(), b.role)
	return b.ddl.exec(q)
}

func (b *PrivilegeBuilder) Revoke() error {
	q := fmt.Sprintf(`REVOKE %s ON %s %s FROM %s;`, b.priviledge, b.object.Type, b.object.QualifiedName(), b.role)
	return b.ddl.exec(q)
}

func PrivilegeId(conn *sqlx.DB, object PriviledgeObjectStruct, roleId, privilege string) (string, error) {
	var id string

	switch t := object.Type; t {
	case "DATABASE":
		i, err := DatabaseId(conn, object.Name)
		if err != nil {
			return "", err
		}
		id = i

	case "SCHEMA":
		i, err := SchemaId(conn, object.Name, object.DatabaseName)
		if err != nil {
			return "", err
		}
		id = i

	case "TABLE":
		o := ObjectSchemaStruct{Name: object.Name, SchemaName: object.SchemaName, DatabaseName: object.DatabaseName}
		i, err := TableId(conn, o)
		if err != nil {
			return "", err
		}
		id = i

	case "VIEW":
		i, err := ViewId(conn, object.Name, object.SchemaName, object.DatabaseName)
		if err != nil {
			return "", err
		}
		id = i

	case "MATERIALIZED VIEW":
		i, err := MaterializedViewId(conn, object.Name, object.SchemaName, object.DatabaseName)
		if err != nil {
			return "", err
		}
		id = i

	case "TYPE":
		i, err := TypeId(conn, object.Name, object.SchemaName, object.DatabaseName)
		if err != nil {
			return "", err
		}
		id = i

	case "SOURCE":
		i, err := SourceId(conn, object.Name, object.SchemaName, object.DatabaseName)
		if err != nil {
			return "", err
		}
		id = i

	case "CONNECTION":
		i, err := ConnectionId(conn, object.Name, object.SchemaName, object.DatabaseName)
		if err != nil {
			return "", err
		}
		id = i

	case "SECRET":
		i, err := SecretId(conn, object.Name, object.SchemaName, object.DatabaseName)
		if err != nil {
			return "", err
		}
		id = i

	case "CLUSTER":
		i, err := ClusterId(conn, object.Name)
		if err != nil {
			return "", err
		}
		id = i
	}

	f := fmt.Sprintf(`GRANT|%s|%s|%s|%s`, object.Type, id, roleId, privilege)
	return f, nil
}

func ScanPrivileges(conn *sqlx.DB, objectType, objectId string) (string, error) {
	var params string

	switch t := objectType; t {
	case "DATABASE":
		p, err := ScanDatabase(conn, objectId)
		if err != nil {
			return "", err
		}
		params = p.Privileges.String

	case "SCHEMA":
		p, err := ScanSchema(conn, objectId)
		if err != nil {
			return "", err
		}
		params = p.Privileges.String

	case "TABLE":
		p, err := ScanTable(conn, objectId)
		if err != nil {
			return "", err
		}
		params = p.Privileges.String

	case "VIEW":
		p, err := ScanView(conn, objectId)
		if err != nil {
			return "", err
		}
		params = p.Privileges.String

	case "MATERIALIZED VIEW":
		p, err := ScanMaterializedView(conn, objectId)
		if err != nil {
			return "", err
		}
		params = p.Privileges.String

	case "TYPE":
		p, err := ScanType(conn, objectId)
		if err != nil {
			return "", err
		}
		params = p.Privileges.String

	case "SOURCE":
		p, err := ScanSource(conn, objectId)
		if err != nil {
			return "", err
		}
		params = p.Privileges.String

	case "CONNECTION":
		p, err := ScanConnection(conn, objectId)
		if err != nil {
			return "", err
		}
		params = p.Privileges.String

	case "SECRET":
		p, err := ScanSecret(conn, objectId)
		if err != nil {
			return "", err
		}
		params = p.Privileges.String

	case "CLUSTER":
		p, err := ScanCluster(conn, objectId)
		if err != nil {
			return "", err
		}
		params = p.Privileges.String
	}

	return params, nil
}
