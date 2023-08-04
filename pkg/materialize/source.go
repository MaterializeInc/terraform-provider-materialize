package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type TableStruct struct {
	Name  string
	Alias string
}

func GetTableStruct(v []interface{}) []TableStruct {
	var tables []TableStruct
	for _, table := range v {
		t := table.(map[string]interface{})
		tables = append(tables, TableStruct{
			Name:  t["name"].(string),
			Alias: t["alias"].(string),
		})
	}
	return tables
}

func DiffTables(old []interface{}, new []interface{}) ([]TableStruct, []TableStruct) {
	var added, dropped []TableStruct
	// TODO: Implement
	return added, dropped
}

type Source struct {
	ddl          Builder
	SourceName   string
	SchemaName   string
	DatabaseName string
}

func NewSource(conn *sqlx.DB, obj ObjectSchemaStruct) *Source {
	return &Source{
		ddl:          Builder{conn, BaseSource},
		SourceName:   obj.Name,
		SchemaName:   obj.SchemaName,
		DatabaseName: obj.DatabaseName,
	}
}

func (s *Source) QualifiedName() string {
	return QualifiedName(s.DatabaseName, s.SchemaName, s.SourceName)
}

func (b *Source) Rename(newConnectionName string) error {
	old := b.QualifiedName()
	new := QualifiedName(newConnectionName)
	return b.ddl.rename(old, new)
}

func (b *Source) Resize(newSize string) error {
	return b.ddl.resize(b.QualifiedName(), newSize)
}

func (b *Source) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

func (b *Source) AddSubsource(subsources []TableStruct) error {
	var subsrc []string
	for _, t := range subsources {
		if t.Alias != "" {
			f := fmt.Sprintf("%s AS %s", t.Name, t.Alias)
			subsrc = append(subsrc, f)
		} else {
			subsrc = append(subsrc, t.Name)
		}
	}
	s := strings.Join(subsrc, ", ")
	q := fmt.Sprintf(`ALTER SOURCE %s ADD SUBSOURCE %s;`, b.QualifiedName(), s)
	return b.ddl.exec(q)
}

func (b *Source) DropSubsource(subsources []TableStruct) error {
	var subsrc []string
	for _, t := range subsources {
		if t.Alias != "" {
			subsrc = append(subsrc, t.Alias)
		} else {
			subsrc = append(subsrc, t.Name)
		}
	}
	s := strings.Join(subsrc, ", ")
	q := fmt.Sprintf(`ALTER SOURCE %s DROP SUBSOURCE %s;`, b.QualifiedName(), s)
	return b.ddl.exec(q)
}

type SourceParams struct {
	SourceId       sql.NullString `db:"id"`
	SourceName     sql.NullString `db:"name"`
	SchemaName     sql.NullString `db:"schema_name"`
	DatabaseName   sql.NullString `db:"database_name"`
	SourceType     sql.NullString `db:"source_type"`
	Size           sql.NullString `db:"size"`
	EnvelopeType   sql.NullString `db:"envelope_type"`
	ConnectionName sql.NullString `db:"connection_name"`
	ClusterName    sql.NullString `db:"cluster_name"`
	OwnerName      sql.NullString `db:"owner_name"`
	Privileges     sql.NullString `db:"privileges"`
}

var sourceQuery = NewBaseQuery(`
		SELECT
			mz_sources.id,
			mz_sources.name,
			mz_schemas.name AS schema_name,
			mz_databases.name AS database_name,
			mz_sources.type AS source_type,
			mz_sources.size,
			mz_sources.envelope_type,
			mz_connections.name as connection_name,
			mz_clusters.name as cluster_name,
			mz_roles.name AS owner_name,
			mz_sources.privileges
		FROM mz_sources
		JOIN mz_schemas
			ON mz_sources.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		LEFT JOIN mz_connections
			ON mz_sources.connection_id = mz_connections.id
		LEFT JOIN mz_clusters
			ON mz_sources.cluster_id = mz_clusters.id
		JOIN mz_roles
			ON mz_sources.owner_id = mz_roles.id`)

func SourceId(conn *sqlx.DB, obj ObjectSchemaStruct) (string, error) {
	p := map[string]string{
		"mz_sources.name":   obj.Name,
		"mz_schemas.name":   obj.SchemaName,
		"mz_databases.name": obj.DatabaseName,
	}
	q := sourceQuery.QueryPredicate(p)

	var c SourceParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return c.SourceId.String, nil
}

func ScanSource(conn *sqlx.DB, id string) (SourceParams, error) {
	q := sourceQuery.QueryPredicate(map[string]string{"mz_sources.id": id})

	var c SourceParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}

func ListSources(conn *sqlx.DB, schemaName, databaseName string) ([]SourceParams, error) {
	p := map[string]string{
		"mz_schemas.name":   schemaName,
		"mz_databases.name": databaseName,
	}
	q := sourceQuery.QueryPredicate(p)

	var c []SourceParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
