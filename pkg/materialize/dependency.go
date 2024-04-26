package materialize

import (
	"database/sql"

	"github.com/jmoiron/sqlx"
)

type DependencyParams struct {
	ObjectId          sql.NullString `db:"object_id"`
	ReferenceObjectId sql.NullString `db:"referenced_object_id"`
	ObjectName        sql.NullString `db:"object_name"`
	SchemaName        sql.NullString `db:"schema_name"`
	DatabaseName      sql.NullString `db:"database_name"`
	Type              sql.NullString `db:"type"`
}

var dependencyQuery = NewBaseQuery(`
	SELECT
		mz_object_dependencies.object_id,
		mz_object_dependencies.referenced_object_id,
		mz_objects.name AS object_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_objects.type
	FROM mz_internal.mz_object_dependencies
	JOIN mz_objects
		ON mz_object_dependencies.object_id = mz_objects.id
	JOIN mz_schemas
		ON mz_objects.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id`)

func ListDependencies(conn *sqlx.DB, objectId, objectType string) ([]DependencyParams, error) {
	p := map[string]string{
		"mz_object_dependencies.referenced_object_id": objectId,
	}

	if objectType != "" {
		p["mz_objects.type"] = objectType
	}

	q := dependencyQuery.QueryPredicate(p)

	var d []DependencyParams
	if err := conn.Select(&d, q); err != nil {
		return d, err
	}

	return d, nil
}
