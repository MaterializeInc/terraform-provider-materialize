package materialize

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type EntityType string

const (
	BaseConnection   EntityType = "CONNECTION"
	BaseSink         EntityType = "SINK"
	BaseSource       EntityType = "SOURCE"
	ClusterReplica   EntityType = "CLUSTER REPLICA"
	Cluster          EntityType = "CLUSTER"
	Database         EntityType = "DATABASE"
	Index            EntityType = "INDEX"
	MaterializedView EntityType = "MATERIALIZED VIEW"
	Ownership        EntityType = "OWNERSHIP"
	Role             EntityType = "ROLE"
	Schema           EntityType = "SCHEMA"
	Secret           EntityType = "SECRET"
	Table            EntityType = "TABLE"
	View             EntityType = "VIEW"
)

type Builder struct {
	conn   *sqlx.DB
	entity EntityType
}

func (b *Builder) exec(statement string) error {
	_, err := b.conn.Exec(statement)
	if err != nil {
		return err
	}

	return nil
}

func (b *Builder) drop(name string) error {
	q := fmt.Sprintf(`DROP %s %s;`, b.entity, name)
	return b.exec(q)
}

func (b *Builder) rename(oldName, newName string) error {
	q := fmt.Sprintf(`ALTER %s %s RENAME TO %s;`, b.entity, oldName, newName)
	return b.exec(q)
}

func (b *Builder) resize(name, size string) error {
	q := fmt.Sprintf(`ALTER %s %s SET (SIZE = %s);`, b.entity, name, size)
	return b.exec(q)
}
