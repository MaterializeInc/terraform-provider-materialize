package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type Table struct {
	Name  string
	Alias string
}

func GetTableStruct(v []interface{}) []Table {
	var tables []Table
	for _, table := range v {
		t := table.(map[string]interface{})
		tables = append(tables, Table{
			Name:  t["name"].(string),
			Alias: t["alias"].(string),
		})
	}
	return tables
}

type Source struct {
	conn         *sqlx.DB
	SourceName   string
	SchemaName   string
	DatabaseName string
}

func NewSource(conn *sqlx.DB, name, schema, database string) *Source {
	return &Source{
		conn:         conn,
		SourceName:   name,
		SchemaName:   schema,
		DatabaseName: database,
	}
}

func (s *Source) QualifiedName() string {
	return QualifiedName(s.DatabaseName, s.SchemaName, s.SourceName)
}

func (b *Source) Rename(newName string) error {
	n := QualifiedName(b.DatabaseName, b.SchemaName, newName)
	q := fmt.Sprintf(`ALTER SOURCE %s RENAME TO %s;`, b.QualifiedName(), n)

	_, err := b.conn.Exec(q)
	if err != nil {
		return err
	}

	return nil
}

func (b *Source) UpdateSize(newSize string) error {
	q := fmt.Sprintf(`ALTER SOURCE %s SET (SIZE = %s);`, b.QualifiedName(), QuoteString(newSize))

	_, err := b.conn.Exec(q)
	if err != nil {
		return err
	}

	return nil
}

func (b *Source) Drop() error {
	q := fmt.Sprintf(`DROP SOURCE %s;`, b.QualifiedName())

	_, err := b.conn.Exec(q)
	if err != nil {
		return err
	}

	return nil
}

func (b *Source) ReadId() (string, error) {
	q := fmt.Sprintf(`
		SELECT mz_sources.id
		FROM mz_sources
		JOIN mz_schemas
			ON mz_sources.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sources.connection_id = mz_connections.id
		JOIN mz_clusters
			ON mz_sources.cluster_id = mz_clusters.id
		WHERE mz_sources.name = %s
		AND mz_schemas.name = %s
		AND mz_databases.name = %s;
	`, QuoteString(b.SourceName), QuoteString(b.SchemaName), QuoteString(b.DatabaseName))

	var i string
	if err := b.conn.QueryRowx(q).Scan(&i); err != nil {
		return "", err
	}

	return i, nil
}

type SourceParams struct {
	SourceName     string `db:"name"`
	SchemaName     string `db:"schema"`
	DatabaseName   string `db:"database"`
	Size           string `db:"size"`
	ConnectionName string `db:"connection_name"`
	ClusterName    string `db:"cluster_name"`
}

func (b *Source) Params(catalogId string) (SourceParams, error) {
	q := fmt.Sprintf(`
		SELECT
			mz_sources.name,
			mz_schemas.name,
			mz_databases.name,
			mz_sources.size,
			mz_connections.name as connection_name,
			mz_clusters.name as cluster_name
		FROM mz_sources
		JOIN mz_schemas
			ON mz_sources.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sources.connection_id = mz_connections.id
		JOIN mz_clusters
			ON mz_sources.cluster_id = mz_clusters.id
		WHERE mz_sources.id = %s;
	`, QuoteString(catalogId))

	var s SourceParams
	if err := b.conn.Get(&s, q); err != nil {
		return s, err
	}

	return s, nil
}

func ReadSourceDatasource(databaseName, schemaName string) string {
	q := strings.Builder{}
	q.WriteString(`
		SELECT
			mz_sources.id,
			mz_sources.name,
			mz_schemas.name,
			mz_databases.name,
			mz_sources.type,
			mz_sources.size,
			mz_sources.envelope_type,
			mz_connections.name as connection_name,
			mz_clusters.name as cluster_name
		FROM mz_sources
		JOIN mz_schemas
			ON mz_sources.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sources.connection_id = mz_connections.id
		LEFT JOIN mz_clusters
			ON mz_sources.cluster_id = mz_clusters.id`)

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
