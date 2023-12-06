package materialize

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type SystemPrivilegeBuilder struct {
	ddl       Builder
	role      MaterializeRole
	privilege string
}

func NewSystemPrivilegeBuilder(conn *sqlx.DB, role, privilege string) *SystemPrivilegeBuilder {
	return &SystemPrivilegeBuilder{
		ddl:       Builder{conn, Privilege},
		role:      MaterializeRole{name: role},
		privilege: privilege,
	}
}

func (b *SystemPrivilegeBuilder) Grant() error {
	q := fmt.Sprintf(`GRANT %s ON SYSTEM TO %s;`, b.privilege, b.role.QualifiedName())
	return b.ddl.exec(q)
}

func (b *SystemPrivilegeBuilder) Revoke() error {
	q := fmt.Sprintf(`REVOKE %s ON SYSTEM FROM %s;`, b.privilege, b.role.QualifiedName())
	return b.ddl.exec(q)
}

func (b *SystemPrivilegeBuilder) GrantKey(region, roleId, privilege string) string {
	return fmt.Sprintf(`%[1]s:GRANT SYSTEM|%[2]s|%[3]s`, region, roleId, privilege)
}

type SytemPrivilegeParams struct {
	Privileges string `db:"privileges"`
}

var systemPrivilegeQuery = `SELECT privileges FROM mz_system_privileges`

func ScanSystemPrivileges(conn *sqlx.DB) ([]SytemPrivilegeParams, error) {
	var c []SytemPrivilegeParams
	if err := conn.Select(&c, systemPrivilegeQuery); err != nil {
		return c, err
	}

	return c, nil
}
