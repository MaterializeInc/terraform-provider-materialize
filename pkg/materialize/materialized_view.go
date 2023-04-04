package materialize

import (
	"fmt"
	"strings"
)

type MaterializedViewBuilder struct {
	materializedViewName string
	schemaName           string
	databaseName         string
	clusterName          string
	selectStmt           string
}

func (b *MaterializedViewBuilder) QualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.materializedViewName)
}

func NewMaterializedViewBuilder(materializedViewName, schemaName, databaseName string) *MaterializedViewBuilder {
	return &MaterializedViewBuilder{
		materializedViewName: materializedViewName,
		schemaName:           schemaName,
		databaseName:         databaseName,
	}
}

func (b *MaterializedViewBuilder) ClusterName(clusterName string) *MaterializedViewBuilder {
	b.clusterName = clusterName
	return b
}

func (b *MaterializedViewBuilder) SelectStmt(selectStmt string) *MaterializedViewBuilder {
	b.selectStmt = selectStmt
	return b
}

func (b *MaterializedViewBuilder) Create() string {
	q := strings.Builder{}

	q.WriteString(fmt.Sprintf(`CREATE MATERIALIZED VIEW %s`, b.QualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, QuoteIdentifier(b.clusterName)))
	}

	q.WriteString(fmt.Sprintf(` AS %s;`, b.selectStmt))
	return q.String()
}

func (b *MaterializedViewBuilder) Rename(newName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newName)
	return fmt.Sprintf(`ALTER MATERIALIZED VIEW %s RENAME TO %s;`, b.QualifiedName(), n)
}

func (b *MaterializedViewBuilder) Drop() string {
	return fmt.Sprintf(`DROP MATERIALIZED VIEW %s;`, b.QualifiedName())
}

func (b *MaterializedViewBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_materialized_views.id
		FROM mz_materialized_views
		JOIN mz_schemas
			ON mz_materialized_views.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_materialized_views.name = %s
		AND mz_schemas.name = %s
		AND mz_databases.name = %s;
	`, QuoteString(b.materializedViewName), QuoteString(b.schemaName), QuoteString(b.databaseName))
}

func ReadMaterializedViewParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_materialized_views.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_materialized_views
		JOIN mz_schemas
			ON mz_materialized_views.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_materialized_views.id = %s;`, QuoteString(id))
}

func ReadMaterializedViewDatasource(databaseName, schemaName string) string {
	q := strings.Builder{}
	q.WriteString(`
		SELECT
			mz_materialized_views.id,
			mz_materialized_views.name,
			mz_schemas.name,
			mz_databases.name,
		SELECT mz_materialized_views.id
		FROM mz_materialized_views
		JOIN mz_schemas
			ON mz_materialized_views.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id`)

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
