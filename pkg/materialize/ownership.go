package materialize

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type OwnershipBuilder struct {
	ddl    Builder
	object MaterializeObject
}

func NewOwnershipBuilder(conn *sqlx.DB, object MaterializeObject) *OwnershipBuilder {
	return &OwnershipBuilder{
		ddl:    Builder{conn, Ownership},
		object: object,
	}
}

func (b *OwnershipBuilder) Object(o MaterializeObject) *OwnershipBuilder {
	b.object = o
	return b
}

func (b *OwnershipBuilder) Alter(roleName string) error {
	q := fmt.Sprintf(`ALTER %s %s OWNER TO "%s";`, b.object.ObjectType, b.object.QualifiedName(), roleName)
	return b.ddl.exec(q)
}
