package materialize

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type Sink struct {
	ddl          Builder
	SinkName     string
	SchemaName   string
	DatabaseName string
}

func NewSink(conn *sqlx.DB, sinkName, schemaName, databaseName string) *Sink {
	return &Sink{
		ddl:          Builder{conn, BaseSink},
		SinkName:     sinkName,
		SchemaName:   schemaName,
		DatabaseName: databaseName,
	}
}
func (s *Sink) QualifiedName() string {
	return QualifiedName(s.DatabaseName, s.SchemaName, s.SinkName)
}

func (b *Sink) Rename(newConnectionName string) error {
	old := b.QualifiedName()
	new := QualifiedName(b.DatabaseName, b.SchemaName, newConnectionName)
	return b.ddl.rename(old, new)
}

func (b *Sink) Resize(newSize string) error {
	return b.ddl.resize(b.QualifiedName(), newSize)
}

func (b *Sink) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

type SinkParams struct {
	SinkId         sql.NullString `db:"id"`
	SinkName       sql.NullString `db:"name"`
	SchemaName     sql.NullString `db:"schema_name"`
	DatabaseName   sql.NullString `db:"database_name"`
	SinkType       sql.NullString `db:"sink_type"`
	Size           sql.NullString `db:"size"`
	EnvelopeType   sql.NullString `db:"envelope_type"`
	ConnectionName sql.NullString `db:"connection_name"`
	ClusterName    sql.NullString `db:"cluster_name"`
}

var sinkQuery = NewBaseQuery(`
	SELECT
		mz_sinks.id,
		mz_sinks.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_sinks.type AS sink_type,
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

func SinkId(conn *sqlx.DB, sinkName, schemaName, databaseName string) (string, error) {
	p := map[string]string{
		"mz_sinks.name":     sinkName,
		"mz_schemas.name":   schemaName,
		"mz_databases.name": databaseName,
	}
	q := sinkQuery.QueryPredicate(p)

	var c SinkParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return c.SinkId.String, nil
}

func ScanSink(conn *sqlx.DB, id string) (SinkParams, error) {
	q := sinkQuery.QueryPredicate(map[string]string{"mz_sinks.id": id})

	var c SinkParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}

func ListSinks(conn *sqlx.DB, schemaName, databaseName string) ([]SinkParams, error) {
	p := map[string]string{
		"mz_schemas.name":   schemaName,
		"mz_databases.name": databaseName,
	}
	q := sinkQuery.QueryPredicate(p)

	var c []SinkParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
