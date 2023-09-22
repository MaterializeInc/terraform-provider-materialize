package materialize

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type CommentBuilder struct {
	ddl    Builder
	object MaterializeObject
}

func NewCommentBuilder(conn *sqlx.DB, obj MaterializeObject) *CommentBuilder {
	return &CommentBuilder{
		ddl:    Builder{conn, Cluster},
		object: obj,
	}
}

func (b *CommentBuilder) Object(comment string) error {
	c := QuoteString(comment)
	q := fmt.Sprintf(`COMMENT ON %s %s IS %s;`, b.object.ObjectType, b.object.QualifiedName(), c)
	return b.ddl.exec(q)
}

func (b *CommentBuilder) Column(column, comment string) error {
	c := QuoteString(comment)
	col := QuoteIdentifier(column)
	q := fmt.Sprintf(`COMMENT ON COLUMN %s.%s IS %s;`, b.object.QualifiedName(), col, c)
	return b.ddl.exec(q)
}
