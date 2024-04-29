package materialize

import (
	"database/sql"
	"log"

	"github.com/jmoiron/sqlx"
)

type SubsourceDetail struct {
	ObjectId          sql.NullString `db:"object_id"`
	ReferenceObjectId sql.NullString `db:"referenced_object_id"`
	ObjectName        sql.NullString `db:"object_name"`
	SchemaName        sql.NullString `db:"schema_name"`
	DatabaseName      sql.NullString `db:"database_name"`
	Type              sql.NullString `db:"type"`
	TableName         sql.NullString `db:"table_name"`
	TableSchemaName   sql.NullString `db:"table_schema_name"`
}

var subsourceQuery = NewBaseQuery(`
	SELECT
		mz_sources.id AS object_id,
		subsources.id AS referenced_object_id,
		mz_sources.name AS object_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_sources.type,
		mz_postgres_source_tables.table_name,
		mz_postgres_source_tables.schema_name AS table_schema_name
	FROM mz_sources AS subsources
	JOIN mz_internal.mz_object_dependencies
		ON subsources.id = mz_object_dependencies.referenced_object_id
	JOIN mz_sources
		ON mz_sources.id = mz_object_dependencies.object_id
	JOIN mz_schemas
		ON mz_sources.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_internal.mz_postgres_source_tables
		ON mz_sources.id = mz_postgres_source_tables.id
`)

func ListPostgresSubsources(conn *sqlx.DB, sourceId string, objectType string) ([]SubsourceDetail, error) {
	p := map[string]string{
		"mz_object_dependencies.referenced_object_id": sourceId,
	}

	if objectType != "" {
		p["mz_sources.type"] = objectType
	}

	q := subsourceQuery.QueryPredicate(p)
	// debug log the query
	log.Printf("[DEBUG] Query: %s", q)
	var subsources []SubsourceDetail
	if err := conn.Select(&subsources, q); err != nil {
		return nil, err
	}
	return subsources, nil
}
