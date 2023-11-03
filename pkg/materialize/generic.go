package materialize

import (
	"fmt"
	"log"

	"github.com/jmoiron/sqlx"
)

type EntityType string

const (
	ClusterReplica   EntityType = "CLUSTER REPLICA"
	Cluster          EntityType = "CLUSTER"
	BaseConnection   EntityType = "CONNECTION"
	Database         EntityType = "DATABASE"
	Index            EntityType = "INDEX"
	MaterializedView EntityType = "MATERIALIZED VIEW"
	Privilege        EntityType = "PRIVILEGE"
	Ownership        EntityType = "OWNERSHIP"
	Role             EntityType = "ROLE"
	Schema           EntityType = "SCHEMA"
	BaseSink         EntityType = "SINK"
	BaseSource       EntityType = "SOURCE"
	Secret           EntityType = "SECRET"
	Table            EntityType = "TABLE"
	BaseType         EntityType = "TYPE"
	View             EntityType = "VIEW"
)

type Builder struct {
	conn   *sqlx.DB
	entity EntityType
}

func (b *Builder) exec(statement string) error {
	if statement[len(statement)-1:] != ";" {
		statement += ";"
	}

	_, err := b.conn.Exec(statement)
	if err != nil {
		log.Printf("[DEBUG] error executing: %s", statement)
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
	q := fmt.Sprintf(`ALTER %s %s SET (SIZE = '%s');`, b.entity, name, size)
	return b.exec(q)
}
