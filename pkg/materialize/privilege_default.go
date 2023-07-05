package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type DefaultPrivilegeBuilder struct {
	ddl          Builder
	grantee      string
	objectType   string
	privilege    string
	targetRole   string
	schemaName   string
	databaseName string
}

func NewDefaultPrivilegeBuilder(conn *sqlx.DB, objectType, grantee, privilege string) *DefaultPrivilegeBuilder {
	return &DefaultPrivilegeBuilder{
		ddl:        Builder{conn, Privilege},
		objectType: objectType,
		privilege:  privilege,
		grantee:    grantee,
	}
}

func (b *DefaultPrivilegeBuilder) TargetRole(c string) *DefaultPrivilegeBuilder {
	b.targetRole = c
	return b
}

func (b *DefaultPrivilegeBuilder) SchemaName(c string) *DefaultPrivilegeBuilder {
	b.schemaName = c
	return b
}

func (b *DefaultPrivilegeBuilder) DatabaseName(c string) *DefaultPrivilegeBuilder {
	b.databaseName = c
	return b
}

func (b *DefaultPrivilegeBuilder) baseQuery(action string) error {
	q := strings.Builder{}
	q.WriteString(`ALTER DEFAULT PRIVILEGES`)

	// role
	if b.targetRole != "" && b.targetRole != "ALL" {
		q.WriteString(fmt.Sprintf(` FOR ROLE %s`, b.targetRole))
	}

	if b.targetRole == "ALL" {
		q.WriteString(" FOR ALL ROLES")
	}

	// object location
	if b.schemaName != "" && b.databaseName != "" {
		q.WriteString(fmt.Sprintf(` IN SCHEMA "%[1]s"."%[2]s"`, b.databaseName, b.schemaName))
	} else if b.databaseName != "" {
		q.WriteString(fmt.Sprintf(` IN DATABASE "%s"`, b.databaseName))
	} else {

	}

	var grantDirection string
	if action == "GRANT" {
		grantDirection = "TO"
	} else {
		grantDirection = "FROM"
	}

	q.WriteString(fmt.Sprintf(` %[1]s %[2]s ON %[3]sS %[4]s %[5]s`, action, b.privilege, b.objectType, grantDirection, b.grantee))

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}

func (b *DefaultPrivilegeBuilder) Grant() error {
	return b.baseQuery("GRANT")
}

func (b *DefaultPrivilegeBuilder) Revoke() error {
	return b.baseQuery("REVOKE")
}

type DefaultPrivilegeParams struct {
	ObjectType   sql.NullString `db:"object_type"`
	GranteeId    sql.NullString `db:"grantee_id"`
	TargetRoleId sql.NullString `db:"role_id"`
	DatabaseId   sql.NullString `db:"database_id"`
	SchemaId     sql.NullString `db:"schema_id"`
	Privileges   sql.NullString `db:"privileges"`
}

var defaultPrivilegeQuery = NewBaseQuery(`
	SELECT
		mz_default_privileges.object_type,
		mz_default_privileges.grantee AS grantee_id,
		mz_default_privileges.role_id,
		mz_default_privileges.database_id AS database_id,
		mz_default_privileges.schema_id AS schema_id,
		mz_default_privileges.privileges
	FROM mz_default_privileges
	LEFT JOIN mz_schemas
		ON mz_default_privileges.schema_id = mz_schemas.id
	LEFT JOIN mz_databases
		ON mz_default_privileges.database_id = mz_databases.id`)

func DefaultPrivilegeId(conn *sqlx.DB, objectType, granteeName, targetRoleName, databaseName, schemaName, privilege string) (string, error) {
	g, err := RoleId(conn, granteeName)
	if err != nil {
		return "", err
	}

	var t, d, s string
	if targetRoleName != "" {
		t, err = RoleId(conn, targetRoleName)
		if err != nil {
			return "", err
		}
	}

	if databaseName != "" {
		d, err = DatabaseId(conn, ObjectSchemaStruct{Name: databaseName})
		if err != nil {
			return "", err
		}
	}

	if schemaName != "" {
		s, err = SchemaId(conn, ObjectSchemaStruct{Name: schemaName, DatabaseName: databaseName})
		if err != nil {
			return "", err
		}
	}

	f := fmt.Sprintf(`GRANT DEFAULT|%[1]s|%[2]s|%[3]s|%[4]s|%[5]s|%[6]s`, objectType, g, t, d, s, privilege)
	return f, nil
}

func ScanDefaultPrivilege(conn *sqlx.DB, objectType, granteeId, targetRoleId, databaseId, schemaId string) (DefaultPrivilegeParams, error) {
	p := map[string]string{
		"mz_default_privileges.object_type": objectType,
		"mz_default_privileges.grantee":     granteeId,
	}

	if targetRoleId != "" {
		p["mz_default_privileges.role_id"] = targetRoleId
	}

	if databaseId != "" {
		p["mz_default_privileges.database_id"] = databaseId
	}

	if schemaId != "" {
		p["mz_default_privileges.schema_id"] = schemaId
	}

	q := defaultPrivilegeQuery.QueryPredicate(p)

	var c DefaultPrivilegeParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
