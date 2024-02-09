package materialize

import (
	"errors"
	"fmt"
	"log"

	"github.com/jackc/pgx"
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
	System           EntityType = "SYSTEM"
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
		var pgErr pgx.PgError
		pgErr, ok := err.(pgx.PgError)
		if ok {
			msg := fmt.Sprintf("%s: %s", pgErr.Severity, pgErr.Message)
			if pgErr.Detail != "" {
				msg += fmt.Sprintf(" DETAIL: %s", pgErr.Detail)
			}
			if pgErr.Hint != "" {
				msg += fmt.Sprintf(" HINT: %s", pgErr.Hint)
			}
			msg += fmt.Sprintf(" (SQLSTATE %s)", pgErr.SQLState())
			return errors.New(msg)
		}
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
