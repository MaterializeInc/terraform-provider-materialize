package materialize

import (
	"github.com/jmoiron/sqlx"
)

// SourceTableLoadGenBuilder for Load Generator sources
type SourceTableLoadGenBuilder struct {
	*SourceTableBuilder
}

func NewSourceTableLoadGenBuilder(conn *sqlx.DB, obj MaterializeObject) *SourceTableLoadGenBuilder {
	return &SourceTableLoadGenBuilder{
		SourceTableBuilder: NewSourceTableBuilder(conn, obj),
	}
}

func (b *SourceTableLoadGenBuilder) Create() error {
	return b.BaseCreate("load-generator", nil)
}
