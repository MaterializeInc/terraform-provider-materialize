package materialize

import (
	"fmt"
	"strings"
)

type IndexColumn struct {
	Field string
	Val   string
}

type IndexBuilder struct {
	indexName    string
	indexDefault bool
	objName      string
	clusterName  string
	method       string
	colExpr      []IndexColumn
}

func (b *IndexBuilder) qualifiedName(databaseName, schemaName string) string {
	return QualifiedName(databaseName, schemaName, b.indexName)
}

func NewIndexBuilder(indexName string) *IndexBuilder {
	return &IndexBuilder{
		indexName: indexName,
	}
}

func (b *IndexBuilder) IndexDefault() *IndexBuilder {
	b.indexDefault = true
	return b
}

func (b *IndexBuilder) ObjName(o string) *IndexBuilder {
	b.objName = o
	return b
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
		q.WriteString(` DEFAULT`)
	} else {
		q.WriteString(fmt.Sprintf(` INDEX %s`, b.indexName))
	}

	q.WriteString(fmt.Sprintf(` IN CLUSTER %s ON %s USING %s`, b.clusterName, b.objName, b.method))

	if len(b.colExpr) > 0 {
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

func (b *IndexBuilder) Drop(databaseName, schemaName string) string {
	return fmt.Sprintf(`DROP INDEX %s RESTRICT;`, b.qualifiedName(databaseName, schemaName))
}

func (b *IndexBuilder) ReadId() string {
	return fmt.Sprintf(`SELECT id FROM mz_indexes WHERE name = '%s';`, b.indexName)
}

func ReadIndexParams(id string) string {
	return fmt.Sprintf(`
		SELECT 
			mz_indexes.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_indexes
		JOIN mz_sources
			ON mz_indexes.on_id = mz_sources.id
		JOIN mz_schemas
			ON mz_sources.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_indexes.id = '%s';`, id)
}
