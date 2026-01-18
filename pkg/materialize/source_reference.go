package materialize

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type SourceReferenceParams struct {
	SourceId         sql.NullString `db:"source_id"`
	Namespace        sql.NullString `db:"namespace"`
	Name             sql.NullString `db:"name"`
	UpdatedAt        sql.NullString `db:"updated_at"`
	Columns          StringArray `db:"columns"`
	SourceName       sql.NullString `db:"source_name"`
	SourceSchemaName sql.NullString `db:"source_schema_name"`
	SourceDBName     sql.NullString `db:"source_database_name"`
	SourceType       sql.NullString `db:"source_type"`
}

var sourceReferenceQuery = NewBaseQuery(`
    SELECT
        sr.source_id,
        sr.namespace,
        sr.name,
        sr.updated_at,
        sr.columns,
        s.name AS source_name,
        ss.name AS source_schema_name,
        sd.name AS source_database_name,
        s.type AS source_type
    FROM mz_internal.mz_source_references sr
    JOIN mz_sources s ON sr.source_id = s.id
    JOIN mz_schemas ss ON s.schema_id = ss.id
    JOIN mz_databases sd ON ss.database_id = sd.id
`)

func SourceReferenceId(conn *sqlx.DB, sourceId string) (string, error) {
	p := map[string]string{
		"sr.source_id": sourceId,
	}
	q := sourceReferenceQuery.QueryPredicate(p)

	var s SourceReferenceParams
	if err := conn.Get(&s, q); err != nil {
		return "", err
	}

	return s.SourceId.String, nil
}

func ScanSourceReference(conn *sqlx.DB, id string) (SourceReferenceParams, error) {
	q := sourceReferenceQuery.QueryPredicate(map[string]string{"sr.source_id": id})

	var s SourceReferenceParams
	if err := conn.Get(&s, q); err != nil {
		return s, err
	}

	return s, nil
}

func refreshSourceReferences(conn *sqlx.DB, sourceName, schemaName, databaseName string) error {
	query := fmt.Sprintf(`ALTER SOURCE %s REFRESH REFERENCES`, QualifiedName(databaseName, schemaName, sourceName))
	_, err := conn.Exec(query)
	return err
}

func ListSourceReferences(conn *sqlx.DB, id string) ([]SourceReferenceParams, error) {
	source, err := ScanSource(conn, id)
	if err == nil {
		if err := refreshSourceReferences(conn, source.SourceName.String, source.SchemaName.String, source.DatabaseName.String); err != nil {
			return nil, fmt.Errorf("error refreshing source references: %v", err)
		}
	}

	p := map[string]string{
		"sr.source_id": id,
	}
	q := sourceReferenceQuery.QueryPredicate(p)

	var references []SourceReferenceParams
	if err := conn.Select(&references, q); err != nil {
		return references, err
	}

	return references, nil
}
