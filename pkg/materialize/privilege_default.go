package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type DefaultPrivilegeBuilder struct {
	ddl          Builder
	targetRole   string
	schemaName   string
	databaseName string
	objectType   string
	privilege    string
	grantee      string
}

func NewDefaultPrivilegeBuilder(conn *sqlx.DB, objectType, privilege, grantee string) *DefaultPrivilegeBuilder {
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
		q.WriteString(fmt.Sprintf(` IN DATABASE "%[1]s"`, b.databaseName))
	} else {

	}

	var grantDirection string
	if action == "GRANT" {
		grantDirection = "TO"
	} else {
		grantDirection = "FROM"
	}

	q.WriteString(fmt.Sprintf(` %[1]s %[2]s ON %[3]s %[4]s %[5]s`, action, b.privilege, b.objectType, grantDirection, b.grantee))

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
	GranteeId    sql.NullString `db:"grantee_id"`
	TargetRoleId sql.NullString `db:"role_id"`
	SchemaName   sql.NullString `db:"schema_name"`
	DatabaseName sql.NullString `db:"database_name"`
	ObjectType   sql.NullString `db:"object_type"`
	Privileges   sql.NullString `db:"privileges"`
}

var defaultPrivilegeQuery = NewBaseQuery(`
	SELECT
		mz_default_privileges.grantee AS grantee_id,
		mz_default_privileges.role_id,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_default_privileges.object_type,
		mz_default_privileges.privileges
	FROM mz_default_privileges
	LEFT JOIN mz_schemas
		ON mz_default_privileges.schema_id = mz_schemas.id
	LEFT JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id`)

func DefaultPrivilegeId(conn *sqlx.DB, objectType, grantee, privilege string) (string, error) {
	granteeId, err := RoleId(conn, grantee)
	if err != nil {
		return "", err
	}

	p := map[string]string{
		"mz_default_privileges.object_type": objectType,
		"mz_default_privileges.grantee":     granteeId,
	}
	q := defaultPrivilegeQuery.QueryPredicate(p)

	var c DefaultPrivilegeParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	f := fmt.Sprintf(`GRANT DEFAULT|%[1]s|%[2]s|%[3]s`, objectType, granteeId, privilege)
	return f, nil
}

func ScanDefaultPrivilege(conn *sqlx.DB, objectType, granteeId string) (DefaultPrivilegeParams, error) {
	p := map[string]string{
		"mz_default_privileges.object_type": objectType,
		"mz_default_privileges.grantee":     granteeId,
	}
	q := defaultPrivilegeQuery.QueryPredicate(p)

	var c DefaultPrivilegeParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
