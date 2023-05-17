package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Sink struct {
	conn         *sqlx.DB
	SinkName     string
	SchemaName   string
	DatabaseName string
}

func NewSink(conn *sqlx.DB, name, schema, database string) *Sink {
	return &Sink{
		conn:         conn,
		SinkName:     name,
		SchemaName:   schema,
		DatabaseName: database,
	}
}

func (s *Sink) QualifiedName() string {
	return QualifiedName(s.DatabaseName, s.SchemaName, s.SinkName)
}

func (b *Sink) Rename(newName string) error {
	n := QualifiedName(b.DatabaseName, b.SchemaName, newName)
	q := fmt.Sprintf(`ALTER SINK %s RENAME TO %s;`, b.QualifiedName(), n)

	_, err := b.conn.Exec(q)
	if err != nil {
		return err
	}

	return nil
}

func (b *Sink) UpdateSize(newSize string) error {
	q := fmt.Sprintf(`ALTER SINK %s SET (SIZE = %s);`, b.QualifiedName(), QuoteString(newSize))

	_, err := b.conn.Exec(q)
	if err != nil {
		return err
	}

	return nil
}

func (b *Sink) Drop() error {
	q := fmt.Sprintf(`DROP SINK %s;`, b.QualifiedName())

	_, err := b.conn.Exec(q)
	if err != nil {
		return err
	}

	return nil
}

func (b *Sink) ReadId() (string, error) {
	q := fmt.Sprintf(`
		SELECT mz_sinks.id
		FROM mz_sinks
		JOIN mz_schemas
			ON mz_sinks.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sinks.connection_id = mz_connections.id
		JOIN mz_clusters
			ON mz_sinks.cluster_id = mz_clusters.id
		WHERE mz_sinks.name = %s
		AND mz_schemas.name = %s
		AND mz_databases.name = %s;
	`, QuoteString(b.SinkName), QuoteString(b.SchemaName), QuoteString(b.DatabaseName))

	var i string
	if err := b.conn.QueryRowx(q).Scan(&i); err != nil {
		return "", err
	}

	return i, nil
}

type SinkParams struct {
	SinkName       string `db:"name"`
	SchemaName     string `db:"schema"`
	DatabaseName   string `db:"database"`
	Size           string `db:"size"`
	ConnectionName string `db:"connection_name"`
	ClusterName    string `db:"cluster_name"`
}

func (b *Sink) Params(catalogId string) (SinkParams, error) {
	q := fmt.Sprintf(`
		SELECT
			mz_sinks.name,
			mz_schemas.name,
			mz_databases.name,
			mz_sinks.size,
			mz_connections.name AS connection_name,
			mz_clusters.name AS cluster_name
		FROM mz_sinks
		JOIN mz_schemas
			ON mz_sinks.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sinks.connection_id = mz_connections.id
		JOIN mz_clusters
			ON mz_sinks.cluster_id = mz_clusters.id
		WHERE mz_sinks.id = %s;
	`, QuoteString(catalogId))

	var s SinkParams
	if err := b.conn.Get(&s, q); err != nil {
		return s, err
	}

	return s, nil
}

func ReadSinkDatasource(databaseName, schemaName string) string {
	q := strings.Builder{}
	q.WriteString(`
		SELECT
			mz_sinks.id,
			mz_sinks.name,
			mz_schemas.name,
			mz_databases.name,
			mz_sinks.type,
			mz_sinks.size,
			mz_sinks.envelope_type,
			mz_connections.name as connection_name,
			mz_clusters.name as cluster_name
		FROM mz_sinks
		JOIN mz_schemas
			ON mz_sinks.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sinks.connection_id = mz_connections.id
		LEFT JOIN mz_clusters
			ON mz_sinks.cluster_id = mz_clusters.id`)

	if databaseName != "" {
		q.WriteString(fmt.Sprintf(`
		WHERE mz_databases.name = %s`, QuoteString(databaseName)))

		if schemaName != "" {
			q.WriteString(fmt.Sprintf(` AND mz_schemas.name = %s`, QuoteString(schemaName)))
		}
	}

	q.WriteString(`;`)
	return q.String()
}
