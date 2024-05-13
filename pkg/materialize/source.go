package materialize

import (
	"database/sql"
	"reflect"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

type TableStruct struct {
	Name               string
	SchemaName         string
	UpstreamName       string
	UpstreamSchemaName string
}

func GetTableStruct(v []interface{}) []TableStruct {
	var tables []TableStruct
	for _, table := range v {
		t := table.(map[string]interface{})
		tables = append(tables, TableStruct{
			Name:               t["name"].(string),
			SchemaName:         t["schema_name"].(string),
			UpstreamName:       t["upstream_name"].(string),
			UpstreamSchemaName: t["upstream_schema_name"].(string),
		})
	}
	return tables
}

func DiffTableStructs(arr1, arr2 []interface{}) []TableStruct {
	var difference []TableStruct

	for _, item1 := range arr1 {
		found := false
		for _, item2 := range arr2 {
			if areEqual(item1, item2) {
				found = true
				break
			}
		}
		if !found {
			if diffItem, ok := item1.(map[string]interface{}); ok {
				difference = append(difference, TableStruct{
					Name:               diffItem["name"].(string),
					SchemaName:         diffItem["schema_name"].(string),
					UpstreamName:       diffItem["upstream_name"].(string),
					UpstreamSchemaName: diffItem["upstream_schema_name"].(string),
				})
			}
		}
	}

	return difference
}

func areEqual(a, b interface{}) bool {
	if reflect.DeepEqual(a, b) {
		return true
	}

	if aItem, ok := a.(map[string]interface{}); ok {
		if bItem, ok := b.(map[string]interface{}); ok {
			return aItem["upstream_name"].(string) == bItem["upstream_name"].(string) && aItem["name"].(string) == bItem["name"].(string) && aItem["schema_name"].(string) == bItem["schema_name"].(string) && aItem["upstream_schema_name"].(string) == bItem["upstream_schema_name"].(string)
		}
	}

	return false
}

type Source struct {
	ddl          Builder
	SourceName   string
	SchemaName   string
	DatabaseName string
}

func NewSource(conn *sqlx.DB, obj MaterializeObject) *Source {
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

func (b *Source) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

func (b *Source) DropCascade() error {
	qn := b.QualifiedName()
	return b.ddl.dropCascade(qn)
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
	Comment        sql.NullString `db:"comment"`
	OwnerName      sql.NullString `db:"owner_name"`
	Privileges     pq.StringArray `db:"privileges"`
}

var sourceQuery = NewBaseQuery(`
		SELECT
			mz_sources.id,
			mz_sources.name,
			mz_schemas.name AS schema_name,
			mz_databases.name AS database_name,
			mz_sources.type AS source_type,
			COALESCE(mz_sources.size, mz_clusters.size) AS size,
			mz_sources.envelope_type,
			mz_connections.name as connection_name,
			mz_clusters.name as cluster_name,
			comments.comment AS comment,
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
			ON mz_sources.owner_id = mz_roles.id
		LEFT JOIN (
			SELECT id, comment
			FROM mz_internal.mz_comments
			WHERE object_type = 'source'
		) comments
			ON mz_sources.id = comments.id`)

func SourceId(conn *sqlx.DB, obj MaterializeObject) (string, error) {
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
