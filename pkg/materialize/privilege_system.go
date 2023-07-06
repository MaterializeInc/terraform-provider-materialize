package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

func ParseSystemPrivileges(privileges []SytemPrivilegeParams) map[string][]string {
	o := map[string][]string{}
	for _, p := range privileges {
		s := strings.Split(p.Privileges.String, "=")

		roleId := s[0]
		rolePrivileges := strings.Split(s[1], "/")[0]

		parsedPrivileges := []string{}
		for _, rp := range strings.Split(rolePrivileges, "") {
			v := Permissions[rp]
			parsedPrivileges = append(parsedPrivileges, v)
		}

		o[roleId] = parsedPrivileges
	}
	return o
}

type SystemPrivilegeBuilder struct {
	ddl       Builder
	role      string
	privilege string
}

func NewSystemPrivilegeBuilder(conn *sqlx.DB, role, privilege string) *SystemPrivilegeBuilder {
	return &SystemPrivilegeBuilder{
		ddl:       Builder{conn, Privilege},
		role:      role,
		privilege: privilege,
	}
}

func (b *SystemPrivilegeBuilder) Grant() error {
	q := fmt.Sprintf(`GRANT %s ON SYSTEM TO %s;`, b.privilege, b.role)
	return b.ddl.exec(q)
}

func (b *SystemPrivilegeBuilder) Revoke() error {
	q := fmt.Sprintf(`REVOKE %s ON SYSTEM FROM %s;`, b.privilege, b.role)
	return b.ddl.exec(q)
}

type SytemPrivilegeParams struct {
	Privileges sql.NullString `db:"privileges"`
}

var systemPrivilegeQuery = `SELECT privileges FROM mz_system_privileges`

func SystemPrivilegeId(conn *sqlx.DB, roleName, privilege string) (string, error) {
	r, err := RoleId(conn, roleName)
	if err != nil {
		return "", err
	}

	f := fmt.Sprintf(`GRANT SYSTEM|%[1]s|%[2]s`, r, privilege)
	return f, nil
}

func ScanSystemPrivileges(conn *sqlx.DB) ([]SytemPrivilegeParams, error) {
	var c []SytemPrivilegeParams
	if err := conn.Select(&c, systemPrivilegeQuery); err != nil {
		return c, err
	}

	return c, nil
}
