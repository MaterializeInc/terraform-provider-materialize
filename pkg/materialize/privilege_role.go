package materialize

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type RolePrivilegeBuilder struct {
	ddl    Builder
	role   string
	member string
}

func NewRolePrivilegeBuilder(conn *sqlx.DB, role, member string) *RolePrivilegeBuilder {
	return &RolePrivilegeBuilder{
		ddl:    Builder{conn, Privilege},
		role:   role,
		member: member,
	}
}

func (b *RolePrivilegeBuilder) Grant() error {
	q := fmt.Sprintf(`GRANT %s TO %s;`, b.role, b.member)
	return b.ddl.exec(q)
}

func (b *RolePrivilegeBuilder) Revoke() error {
	q := fmt.Sprintf(`REVOKE %s FROM %s;`, b.role, b.member)
	return b.ddl.exec(q)
}

type RolePrivilegeParams struct {
	RoleId  sql.NullString `db:"role_id"`
	Member  sql.NullString `db:"member"`
	Grantor sql.NullString `db:"grantor"`
}

var rolePrivilegeQuery = NewBaseQuery(`
	SELECT
		mz_role_members.role_id,
		mz_role_members.member,
		mz_role_members.grantor
	FROM mz_role_members`)

func RolePrivilegeId(conn *sqlx.DB, roleName, memberName string) (string, error) {
	r, err := RoleId(conn, roleName)
	if err != nil {
		return "", err
	}

	m, err := RoleId(conn, memberName)
	if err != nil {
		return "", err
	}

	f := fmt.Sprintf(`ROLE MEMBER|%[1]s|%[2]s`, r, m)
	return f, nil
}

func ScanRolePrivilege(conn *sqlx.DB, roleId, memberId string) (RolePrivilegeParams, error) {
	p := map[string]string{
		"mz_role_members.role_id": roleId,
		"mz_role_members.member":  memberId,
	}

	q := rolePrivilegeQuery.QueryPredicate(p)

	var c RolePrivilegeParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
