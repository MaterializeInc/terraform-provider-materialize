package materialize

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type RolePrivilegeBuilder struct {
	ddl    Builder
	role   MaterializeRole
	member MaterializeRole
}

func NewRolePrivilegeBuilder(conn *sqlx.DB, role, member string) *RolePrivilegeBuilder {
	return &RolePrivilegeBuilder{
		ddl:    Builder{conn, Privilege},
		role:   MaterializeRole{name: role},
		member: MaterializeRole{name: member},
	}
}

func (b *RolePrivilegeBuilder) Grant() error {
	q := fmt.Sprintf(`GRANT %s TO %s;`, b.role.QualifiedName(), b.member.QualifiedName())
	return b.ddl.exec(q)
}

func (b *RolePrivilegeBuilder) Revoke() error {
	q := fmt.Sprintf(`REVOKE %s FROM %s;`, b.role.QualifiedName(), b.member.QualifiedName())
	return b.ddl.exec(q)
}

func (b *RolePrivilegeBuilder) GrantKey(region, roleId, memberId string) string {
	return fmt.Sprintf(`%[1]s:ROLE MEMBER|%[2]s|%[3]s`, region, roleId, memberId)
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

func ScanRolePrivilege(conn *sqlx.DB, roleId, memberId string) ([]RolePrivilegeParams, error) {
	p := map[string]string{
		"mz_role_members.role_id": roleId,
		"mz_role_members.member":  memberId,
	}

	q := rolePrivilegeQuery.QueryPredicate(p)

	var c []RolePrivilegeParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}

func ParseRolePrivileges(privileges []RolePrivilegeParams) (map[string][]string, error) {
	mapping := make(map[string][]string)

	for _, p := range privileges {
		mapping[p.RoleId.String] = append(mapping[p.RoleId.String], p.Member.String)
	}

	return mapping, nil
}
