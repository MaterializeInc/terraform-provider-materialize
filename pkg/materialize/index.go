package materialize

import (
	"fmt"
	"strings"
)

type IndexColumn struct {
	Field string
	Val   string
}

func GetIndexColumnStruct(v []interface{}) []IndexColumn {
	var i []IndexColumn
	for _, column := range v {
		c := column.(map[string]interface{})
		i = append(i, IndexColumn{
			Field: c["field"].(string),
			Val:   c["val"].(string),
		})
	}
	return i
}

type IndexBuilder struct {
	indexName    string
	indexDefault bool
	objName      IdentifierSchemaStruct
	clusterName  string
	method       string
	colExpr      []IndexColumn
}

func (b *IndexBuilder) QualifiedName() string {
	return QualifiedName(b.objName.DatabaseName, b.objName.SchemaName, b.indexName)
}

func NewIndexBuilder(indexName string, indexDefault bool, objName IdentifierSchemaStruct) *IndexBuilder {
	return &IndexBuilder{
		indexName:    indexName,
		indexDefault: indexDefault,
		objName:      objName,
	}
}

func (b *IndexBuilder) ClusterName(c string) *IndexBuilder {
	b.clusterName = c
	return b
}

func (b *IndexBuilder) Method(m string) *IndexBuilder {
	b.method = m
	return b
}

func (b *IndexBuilder) ColExpr(c []IndexColumn) *IndexBuilder {
	b.colExpr = c
	return b
}

func (b *IndexBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(`CREATE`)

	if b.indexDefault {
		q.WriteString(` DEFAULT INDEX`)
	} else {
		q.WriteString(fmt.Sprintf(` INDEX %s`, b.indexName))
	}

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, b.clusterName))
	}

	q.WriteString(fmt.Sprintf(` ON %s`, b.objName.QualifiedName()))

	if b.method != "" {
		q.WriteString(fmt.Sprintf(` USING %s`, b.method))
	}

	if len(b.colExpr) > 0 && !b.indexDefault {
		var columns []string

		for _, c := range b.colExpr {
			s := strings.Builder{}

			s.WriteString(fmt.Sprintf(`%s %s`, c.Field, c.Val))
			o := s.String()
			columns = append(columns, o)

		}
		p := strings.Join(columns[:], ", ")
		q.WriteString(fmt.Sprintf(` (%s)`, p))
	} else {
		q.WriteString(` ()`)
	}

	q.WriteString(`;`)
	return q.String()
}

func (b *IndexBuilder) Drop() string {
	return fmt.Sprintf(`DROP INDEX %s RESTRICT;`, b.QualifiedName())
}

func (b *IndexBuilder) ReadId() string {
	return fmt.Sprintf(`SELECT id FROM mz_indexes WHERE name = %s;`, QuoteString(b.indexName))
}

func ReadIndexParams(id string) string {
	return fmt.Sprintf(`
		SELECT DISTINCT
			mz_indexes.name,
			COALESCE(sources.o_name, views.o_name, mviews.o_name),
			COALESCE(sources.s_name, views.s_name, mviews.s_name),
			COALESCE(sources.d_name, views.d_name, mviews.d_name)
		FROM mz_indexes
		LEFT JOIN (
			SELECT
				mz_sources.id,
				mz_sources.name AS o_name,
				mz_schemas.name AS s_name,
				mz_databases.name AS d_name
			FROM mz_sources
			JOIN mz_schemas
				ON mz_sources.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
		) sources
			ON mz_indexes.on_id = sources.id
		LEFT JOIN (
			SELECT
				mz_views.id,
				mz_views.name AS o_name,
				mz_schemas.name AS s_name,
				mz_databases.name AS d_name
			FROM mz_views
			JOIN mz_schemas
				ON mz_views.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
		) views
			ON mz_indexes.on_id = sources.id
		LEFT JOIN (
			SELECT
				mz_materialized_views.id,
				mz_materialized_views.name AS o_name,
				mz_schemas.name AS s_name,
				mz_databases.name AS d_name
			FROM mz_materialized_views
			JOIN mz_schemas
				ON mz_materialized_views.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
		) mviews
			ON mz_indexes.on_id = sources.id
		WHERE mz_indexes.id = %s;`, QuoteString(id))
}

func ReadIndexDatasource(databaseName, schemaName string) string {
	q := strings.Builder{}
	q.WriteString(`
		SELECT DISTINCT
			mz_indexes.id,
			mz_indexes.name,
			COALESCE(sources.o_name, views.o_name, mviews.o_name),
			COALESCE(sources.s_name, views.s_name, mviews.s_name),
			COALESCE(sources.d_name, views.d_name, mviews.d_name)
		FROM mz_indexes
		LEFT JOIN (
			SELECT
				mz_sources.id,
				mz_sources.name AS o_name,
				mz_schemas.name AS s_name,
				mz_databases.name AS d_name
			FROM mz_sources
			JOIN mz_schemas
				ON mz_sources.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
		) sources
			ON mz_indexes.on_id = sources.id
		LEFT JOIN (
			SELECT
				mz_views.id,
				mz_views.name AS o_name,
				mz_schemas.name AS s_name,
				mz_databases.name AS d_name
			FROM mz_views
			JOIN mz_schemas
				ON mz_views.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
		) views
			ON mz_indexes.on_id = sources.id
		LEFT JOIN (
			SELECT
				mz_materialized_views.id,
				mz_materialized_views.name AS o_name,
				mz_schemas.name AS s_name,
				mz_databases.name AS d_name
			FROM mz_materialized_views
			JOIN mz_schemas
				ON mz_materialized_views.schema_id = mz_schemas.id
			JOIN mz_databases
				ON mz_schemas.database_id = mz_databases.id
		) mviews
			ON mz_indexes.on_id = sources.id
		WHERE COALESCE(sources.o_name, views.o_name, mviews.o_name) IS NOT NULL`)

	if databaseName != "" {
		q.WriteString(fmt.Sprintf(`
		AND mz_databases.name = %s`, QuoteString(databaseName)))

		if schemaName != "" {
			q.WriteString(fmt.Sprintf(` AND mz_schemas.name = %s`, QuoteString(schemaName)))
		}
	}

	q.WriteString(`;`)
	return q.String()
}
