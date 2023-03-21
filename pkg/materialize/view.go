package materialize

import (
	"fmt"
	"strings"
)

type ViewBuilder struct {
	viewName     string
	schemaName   string
	databaseName string
	selectStmt   string
}

func (b *ViewBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.viewName)
}

func NewViewBuilder(viewName, schemaName, databaseName string) *ViewBuilder {
	return &ViewBuilder{
		viewName:     viewName,
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (b *ViewBuilder) SelectStmt(selectStmt string) *ViewBuilder {
	b.selectStmt = selectStmt
	return b
}

func (b *ViewBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(`CREATE`)

	q.WriteString(fmt.Sprintf(` VIEW %s`, b.qualifiedName()))

	q.WriteString(` AS `)

	q.WriteString(b.selectStmt)

	q.WriteString(`;`)

	return q.String()
}

func (b *ViewBuilder) Rename(newName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newName)
	return fmt.Sprintf(`ALTER VIEW %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *ViewBuilder) Drop() string {
	return fmt.Sprintf(`DROP VIEW %s;`, b.qualifiedName())
}

func (b *ViewBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_views.id
		FROM mz_views
		JOIN mz_schemas
			ON mz_views.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_views.name = %s
		AND mz_schemas.name = %s
		AND mz_databases.name = %s;
	`, QuoteString(b.viewName), QuoteString(b.schemaName), QuoteString(b.databaseName))
}

func ReadViewParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_views.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_views
		JOIN mz_schemas
			ON mz_views.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_views.id = %s;`, QuoteString(id))
}

func ReadViewDatasource(databaseName, schemaName string) string {
	q := strings.Builder{}
	q.WriteString(`
		SELECT
			mz_views.id,
			mz_views.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_views
		JOIN mz_schemas
			ON mz_views.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id`)

	if databaseName != "" {
		q.WriteString(fmt.Sprintf(`
		WHERE mz_databases.name = '%s'`, databaseName))

		if schemaName != "" {
			q.WriteString(fmt.Sprintf(` AND mz_schemas.name = '%s'`, schemaName))
		}
	}

	q.WriteString(`;`)
	return q.String()
}
