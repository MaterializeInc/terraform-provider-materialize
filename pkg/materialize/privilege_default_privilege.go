package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type DefaultPrivilegeBuilder struct {
	ddl          Builder
	granteeRole  MaterializeRole
	targetRole   MaterializeRole
	objectType   string
	privilege    string
	schemaName   string
	databaseName string
}

func NewDefaultPrivilegeBuilder(conn *sqlx.DB, objectType, grantee, target, privilege string) *DefaultPrivilegeBuilder {
	return &DefaultPrivilegeBuilder{
		ddl:         Builder{conn, Privilege},
		objectType:  objectType,
		privilege:   privilege,
		granteeRole: MaterializeRole{name: grantee},
		targetRole:  MaterializeRole{name: target},
	}
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
	if b.targetRole.name == "PUBLIC" {
		q.WriteString(" FOR ALL ROLES")
	} else {
		q.WriteString(fmt.Sprintf(` FOR ROLE %s`, b.targetRole.QualifiedName()))
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

	q.WriteString(fmt.Sprintf(` %[1]s %[2]s ON %[3]sS %[4]s %[5]s`, action, b.privilege, b.objectType, grantDirection, b.granteeRole.QualifiedName()))

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}

func (b *DefaultPrivilegeBuilder) Grant() error {
	return b.baseQuery("GRANT")
}

func (b *DefaultPrivilegeBuilder) Revoke() error {
	return b.baseQuery("REVOKE")
}

func (b *DefaultPrivilegeBuilder) GrantKey(objectType, granteeId, targetId, databaseId, schemaId, privilege string) string {
	return fmt.Sprintf(`GRANT DEFAULT|%[1]s|%[2]s|%[3]s|%[4]s|%[5]s|%[6]s`, objectType, granteeId, targetId, databaseId, schemaId, privilege)
}

type DefaultPrivilegeParams struct {
	ObjectType  sql.NullString `db:"object_type"`
	GranteeId   sql.NullString `db:"grantee_id"`
	GranteeName sql.NullString `db:"grantee_name"`
	TargetId    sql.NullString `db:"target_id"`
	TargetName  sql.NullString `db:"target_name"`
	DatabaseId  sql.NullString `db:"database_id"`
	SchemaId    sql.NullString `db:"schema_id"`
	Privileges  sql.NullString `db:"privileges"`
}

var defaultPrivilegeQuery = NewBaseQuery(`
	SELECT
		mz_default_privileges.object_type,
		mz_default_privileges.grantee AS grantee_id,
		(CASE WHEN mz_default_privileges.grantee = 'p' THEN 'PUBLIC' ELSE grantee.name END) AS grantee_name,
		mz_default_privileges.role_id AS target_id,
		(CASE WHEN mz_default_privileges.role_id = 'p' THEN 'PUBLIC' ELSE target.name END) AS target_name,
		mz_default_privileges.database_id AS database_id,
		mz_default_privileges.schema_id AS schema_id,
		mz_default_privileges.privileges
	FROM mz_default_privileges
	LEFT JOIN mz_roles AS grantee
		ON mz_default_privileges.grantee = grantee.id
	LEFT JOIN mz_roles AS target
		ON mz_default_privileges.role_id = target.id
	LEFT JOIN mz_schemas
		ON mz_default_privileges.schema_id = mz_schemas.id
	LEFT JOIN mz_databases
		ON mz_default_privileges.database_id = mz_databases.id`)

func ScanDefaultPrivilege(conn *sqlx.DB, objectType, granteeId, targetRoleId, databaseId, schemaId string) ([]DefaultPrivilegeParams, error) {
	p := map[string]string{
		"mz_default_privileges.object_type": strings.ToLower(objectType),
		"mz_default_privileges.grantee":     granteeId,
		"mz_default_privileges.role_id":     targetRoleId,
	}

	if databaseId != "" {
		p["mz_default_privileges.database_id"] = databaseId
	}

	if schemaId != "" {
		p["mz_default_privileges.schema_id"] = schemaId
	}

	q := defaultPrivilegeQuery.QueryPredicate(p)

	var c []DefaultPrivilegeParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}

func MapDefaultGrantPrivileges(privileges []DefaultPrivilegeParams) (map[string][]string, error) {
	mapping := make(map[string][]string)
	for _, p := range privileges {
		key := p.ObjectType.String + "|" + p.GranteeId.String + "|" + p.DatabaseId.String + "|" + p.SchemaId.String
		parsedPrivileges := []string{}
		for _, rp := range strings.Split(p.Privileges.String, "") {
			pName, _ := PrivilegeName(rp)
			parsedPrivileges = append(parsedPrivileges, pName)
		}
		mapping[key] = parsedPrivileges
	}
	return mapping, nil
}
