package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type RoleBuilder struct {
	ddl          Builder
	roleName     string
	inherit      bool
	password     string
	superuser    bool
	superuserSet bool
}

func NewRoleBuilder(conn *sqlx.DB, obj MaterializeObject) *RoleBuilder {
	return &RoleBuilder{
		ddl:      Builder{conn, Role},
		roleName: obj.Name,
	}
}

func (b *RoleBuilder) QualifiedName() string {
	return QualifiedName(b.roleName)
}

func (b *RoleBuilder) Inherit() *RoleBuilder {
	b.inherit = true
	return b
}

func (b *RoleBuilder) Password(password string) *RoleBuilder {
	b.password = password
	return b
}

func (b *RoleBuilder) Superuser(superuser bool) *RoleBuilder {
	b.superuser = superuser
	b.superuserSet = true
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

	if b.password != "" {
		p = append(p, fmt.Sprintf(` WITH PASSWORD %s`, QuoteString(b.password)))
	}

	if b.superuserSet {
		if b.superuser {
			p = append(p, ` SUPERUSER`)
		} else {
			p = append(p, ` NOSUPERUSER`)
		}
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

func (b *RoleBuilder) AlterPassword(password string) error {
	q := fmt.Sprintf(`ALTER ROLE %s WITH PASSWORD %s;`, b.QualifiedName(), QuoteString(password))
	return b.ddl.exec(q)
}

func (b *RoleBuilder) AlterSuperuser(superuser bool) error {
	permission := "NOSUPERUSER"
	if superuser {
		permission = "SUPERUSER"
	}
	return b.Alter(permission)
}

func (b *RoleBuilder) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

type RoleParams struct {
	RoleId    sql.NullString `db:"id"`
	RoleName  sql.NullString `db:"role_name"`
	Inherit   sql.NullBool   `db:"inherit"`
	Superuser sql.NullBool   `db:"superuser"`
	Comment   sql.NullString `db:"comment"`
}

var roleQuery = NewBaseQuery(`
	SELECT
		mz_roles.id,
		mz_roles.name AS role_name,
		mz_roles.inherit,
		pg_roles.rolsuper AS superuser,
		comments.comment AS comment
	FROM mz_roles
	LEFT JOIN pg_roles ON mz_roles.name = pg_roles.rolname
	LEFT JOIN (
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'role'
	) comments
		ON mz_roles.id = comments.id`)

func RoleId(conn *sqlx.DB, roleName string) (string, error) {
	if roleName == "PUBLIC" {
		return "p", nil
	} else {
		p := map[string]string{"mz_roles.name": roleName}
		q := roleQuery.QueryPredicate(p)

		var c RoleParams
		if err := conn.Get(&c, q); err != nil {
			return "", err
		}

		return c.RoleId.String, nil
	}
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
