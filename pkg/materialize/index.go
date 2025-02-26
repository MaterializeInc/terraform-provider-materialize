package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type IndexColumn struct {
	Field string
}

func GetIndexColumnStruct(v []interface{}) []IndexColumn {
	var i []IndexColumn
	for _, column := range v {
		c := column.(map[string]interface{})
		i = append(i, IndexColumn{
			Field: c["field"].(string),
		})
	}
	return i
}

type IndexBuilder struct {
	ddl          Builder
	indexName    string
	indexDefault bool
	objName      IdentifierSchemaStruct
	clusterName  string
	method       string
	colExpr      []IndexColumn
}

func NewIndexBuilder(conn *sqlx.DB, obj MaterializeObject, indexDefault bool, objName IdentifierSchemaStruct) *IndexBuilder {
	return &IndexBuilder{
		ddl:          Builder{conn, Index},
		indexName:    obj.Name,
		indexDefault: indexDefault,
		objName:      objName,
	}
}

func (b *IndexBuilder) QualifiedName() string {
	return QualifiedName(b.objName.DatabaseName, b.objName.SchemaName, b.indexName)
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

func (b *IndexBuilder) Create() error {
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

	if !b.indexDefault {
		if len(b.colExpr) > 0 {
			var columns []string
			for _, c := range b.colExpr {
				columns = append(columns, c.Field)
			}
			p := strings.Join(columns[:], ", ")
			q.WriteString(fmt.Sprintf(` (%s)`, p))
		} else {
			q.WriteString(` ()`)
		}
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}

func (b *IndexBuilder) Drop() error {
	q := fmt.Sprintf(`DROP INDEX %s RESTRICT;`, b.QualifiedName())
	return b.ddl.exec(q)
}

// Requires a specific comment for the way indexes handle qualified name
func (b *IndexBuilder) Comment(comment string) error {
	c := QuoteString(comment)
	q := fmt.Sprintf(`COMMENT ON INDEX %s IS %s;`, b.QualifiedName(), c)
	return b.ddl.exec(q)
}

type IndexParams struct {
	IndexId            sql.NullString `db:"id"`
	IndexName          sql.NullString `db:"index_name"`
	ObjectName         sql.NullString `db:"obj_name"`
	ObjectSchemaName   sql.NullString `db:"obj_schema_name"`
	ObjectDatabaseName sql.NullString `db:"obj_database_name"`
	Comment            sql.NullString `db:"comment"`
}

var indexQuery = NewBaseQuery(`
	SELECT
		mz_indexes.id,
		mz_indexes.name AS index_name,
		mz_objects.name AS obj_name,
		mz_schemas.name AS obj_schema_name,
		mz_databases.name AS obj_database_name,
		comments.comment AS comment
	FROM mz_indexes
	JOIN mz_objects
		ON mz_indexes.on_id = mz_objects.id
	LEFT JOIN mz_schemas
		ON mz_objects.schema_id = mz_schemas.id
	LEFT JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN (
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'index'
	) comments
		ON mz_indexes.id = comments.id`).
	CustomPredicate([]string{"mz_objects.type IN ('source', 'view', 'materialized-view')"})

func IndexId(conn *sqlx.DB, indexName string) (string, error) {
	q := indexQuery.QueryPredicate(map[string]string{"mz_indexes.name": indexName})

	var c IndexParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return c.IndexId.String, nil
}

func ScanIndex(conn *sqlx.DB, id string) (IndexParams, error) {
	q := indexQuery.QueryPredicate(map[string]string{"mz_indexes.id": id})

	var c IndexParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}

func ListIndexes(conn *sqlx.DB, schemaName, databaseName string) ([]IndexParams, error) {
	p := map[string]string{
		"mz_schemas.name":   schemaName,
		"mz_databases.name": databaseName,
	}

	q := indexQuery.QueryPredicate(p)

	var c []IndexParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
