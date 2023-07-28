package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type RoleBuilder struct {
	ddl      Builder
	roleName string
	inherit  bool
}

func NewRoleBuilder(conn *sqlx.DB, roleName string) *RoleBuilder {
	return &RoleBuilder{
		ddl:      Builder{conn, Role},
		roleName: roleName,
	}
}

func (b *RoleBuilder) QualifiedName() string {
	return QualifiedName(b.roleName)
}

func (b *RoleBuilder) Inherit() *RoleBuilder {
	b.inherit = true
	return b
}

func (b *RoleBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE ROLE %s`, b.QualifiedName()))

	var p []string

	// NOINHERIT currently not supported
	// https://materialize.com/docs/sql/create-role/#details
	if b.inherit {
		p = append(p, ` INHERIT`)
	}

	if len(p) > 0 {
		f := strings.Join(p, "")
		q.WriteString(f)
	}

	q.WriteString(`;`)

	return b.ddl.exec(q.String())
}

func (b *RoleBuilder) Alter(permission string) error {
	q := fmt.Sprintf(`ALTER ROLE %s %s;`, b.QualifiedName(), permission)
	return b.ddl.exec(q)
}

func (b *RoleBuilder) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

type RoleParams struct {
	RoleId   sql.NullString `db:"id"`
	RoleName sql.NullString `db:"role_name"`
	Inherit  sql.NullBool   `db:"inherit"`
}

var roleQuery = NewBaseQuery(`
	SELECT
		id,
		name AS role_name,
		inherit
	FROM mz_roles`)

func RoleId(conn *sqlx.DB, roleName string) (string, error) {
	p := map[string]string{"mz_roles.name": roleName}
	q := roleQuery.QueryPredicate(p)

	var c RoleParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return c.RoleId.String, nil
}

func ScanRole(conn *sqlx.DB, id string) (RoleParams, error) {
	p := map[string]string{"mz_roles.id": id}
	q := roleQuery.QueryPredicate(p)

	var c RoleParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}

func ListRoles(conn *sqlx.DB) ([]RoleParams, error) {
	q := roleQuery.QueryPredicate(map[string]string{})

	var c []RoleParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
