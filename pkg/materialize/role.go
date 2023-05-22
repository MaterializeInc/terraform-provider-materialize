package materialize

import (
	"fmt"
	"strings"
)

type RoleBuilder struct {
	roleName      string
	inherit       bool
	createRole    bool
	createDb      bool
	createCluster bool
}

func (b *RoleBuilder) QualifiedName() string {
	return QualifiedName(b.roleName)
}

func NewRoleBuilder(roleName string) *RoleBuilder {
	return &RoleBuilder{
		roleName: roleName,
	}
}

func (b *RoleBuilder) Inherit() *RoleBuilder {
	b.inherit = true
	return b
}

func (b *RoleBuilder) CreateRole() *RoleBuilder {
	b.createRole = true
	return b
}

func (b *RoleBuilder) CreateDb() *RoleBuilder {
	b.createDb = true
	return b
}

func (b *RoleBuilder) CreateCluster() *RoleBuilder {
	b.createCluster = true
	return b
}

func (b *RoleBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE ROLE %s`, b.QualifiedName()))

	var p []string

	if b.inherit {
		p = append(p, ` INHERIT`)
	}

	if b.createRole {
		p = append(p, ` CREATEROLE`)
	}

	if b.createDb {
		p = append(p, ` CREATEDB`)
	}

	if b.createCluster {
		p = append(p, ` CREATECLUSTER`)
	}

	if len(p) > 0 {
		f := strings.Join(p, "")
		q.WriteString(f)
	}

	q.WriteString(`;`)
	return q.String()
}

func (b *RoleBuilder) Alter(permission string) string {
	return fmt.Sprintf(`ALTER ROLE %s %s;`, b.QualifiedName(), permission)
}

func (b *RoleBuilder) Drop() string {
	return fmt.Sprintf(`DROP ROLE %s;`, b.QualifiedName())
}

func (b *RoleBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT id
		FROM mz_roles
		WHERE name = %s`, QuoteString(b.roleName))
}

func ReadRoleParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			name AS role_name,
			inherit,
			create_role,
			create_db,
			create_cluster
		FROM mz_roles
		WHERE id = %s;`, QuoteString(id))
}

func ReadRoleDatasource() string {
	return "SELECT id, name FROM mz_roles;"
}
