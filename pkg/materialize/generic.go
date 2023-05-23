package materialize

import (
	"fmt"

	"github.com/jmoiron/sqlx"
)

type EntityType string

const (
	ClusterReplica EntityType = "CLUSTER REPLICA"
	Cluster        EntityType = "CLUSTER"
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

	_, err := b.conn.Exec(q)
	if err != nil {
		return err
	}

	return nil
}

func (b *Builder) rename(oldName, newName string) error {
	q := fmt.Sprintf(`ALTER %s %s RENAME TO %s;`, b.entity, oldName, newName)

	_, err := b.conn.Exec(q)
	if err != nil {
		return err
	}

	return nil
}

func (b *Builder) resize(name, size string) error {
	q := fmt.Sprintf(`ALTER %s %s SET (SIZE = %s);`, b.entity, name, size)

	_, err := b.conn.Exec(q)
	if err != nil {
		return err
	}

	return nil
}
