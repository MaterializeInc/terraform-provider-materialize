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
		SELECT
			mz_indexes.name,
			mz_objects.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_indexes
		JOIN mz_objects
			ON mz_indexes.on_id = mz_objects.id
		LEFT JOIN mz_schemas
			ON mz_objects.schema_id = mz_schemas.id
		LEFT JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_objects.type IN ('source', 'view', 'materialized-view')
		AND mz_indexes.id = %s;
	`, QuoteString(id))
}

func ReadIndexDatasource(databaseName, schemaName string) string {
	q := strings.Builder{}
	q.WriteString(`
		SELECT
			mz_indexes.id,
			mz_indexes.name,
			mz_objects.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_indexes
		JOIN mz_objects
			ON mz_indexes.on_id = mz_objects.id
		LEFT JOIN mz_schemas
			ON mz_objects.schema_id = mz_schemas.id
		LEFT JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_objects.type IN ('source', 'view', 'materialized-view')
	`)

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
