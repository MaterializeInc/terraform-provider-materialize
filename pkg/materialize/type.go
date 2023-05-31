package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type ListProperties struct {
	ElementType string
}

func GetListProperties(v interface{}) []ListProperties {
	var lp []ListProperties
	for _, properties := range v.([]interface{}) {
		b := properties.(map[string]interface{})
		lp = append(lp, ListProperties{
			ElementType: b["element_type"].(string),
		})
	}
	return lp
}

type MapProperties struct {
	KeyType   string
	ValueType string
}

func GetMapProperties(v interface{}) []MapProperties {
	var lp []MapProperties
	for _, properties := range v.([]interface{}) {
		b := properties.(map[string]interface{})
		lp = append(lp, MapProperties{
			KeyType:   b["key_type"].(string),
			ValueType: b["value_type"].(string),
		})
	}
	return lp
}

type Type struct {
	ddl            Builder
	typeName       string
	schemaName     string
	databaseName   string
	listProperties []ListProperties
	mapProperties  []MapProperties
}

func NewTypeBuilder(conn *sqlx.DB, typeName, schemaName, databaseName string) *Type {
	return &Type{
		ddl:          Builder{conn, BaseType},
		typeName:     typeName,
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (c *Type) QualifiedName() string {
	return QualifiedName(c.databaseName, c.schemaName, c.typeName)
}

func (b *Type) ListProperties(l []ListProperties) *Type {
	b.listProperties = l
	return b
}

func (b *Type) MapProperties(m []MapProperties) *Type {
	b.mapProperties = m
	return b
}

func (b *Type) Create() error {
	q := strings.Builder{}

	q.WriteString(fmt.Sprintf(`CREATE TYPE %s AS `, b.QualifiedName()))

	var properties []string
	if len(b.listProperties) > 0 {
		q.WriteString(`LIST `)

		for _, p := range b.listProperties {
			f := fmt.Sprintf(`ELEMENT TYPE = %s`, p.ElementType)
			properties = append(properties, f)
		}
	}

	if len(b.mapProperties) > 0 {
		q.WriteString(`MAP `)

		for _, p := range b.mapProperties {
			f := fmt.Sprintf(`KEY TYPE %s, VALUE TYPE = %s`, p.KeyType, p.ValueType)
			properties = append(properties, f)
		}
	}

	p := strings.Join(properties[:], ", ")
	q.WriteString(fmt.Sprintf(`(%s);`, p))
	return b.ddl.exec(q.String())
}

func (b *Type) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

type TypeParams struct {
	TypeId       sql.NullString `db:"id"`
	TypeName     sql.NullString `db:"name"`
	SchemaName   sql.NullString `db:"schema_name"`
	DatabaseName sql.NullString `db:"database_name"`
	Category     sql.NullString `db:"category"`
}

var typeQuery = NewBaseQuery(`
	SELECT
		mz_types.id,
		mz_types.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_types.category
	FROM mz_types
	JOIN mz_schemas
		ON mz_types.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id`)

func TypeId(conn *sqlx.DB, typeName, schemaName, databaseName string) (string, error) {
	p := map[string]string{
		"mz_types.name":     typeName,
		"mz_schemas.name":   schemaName,
		"mz_databases.name": databaseName,
	}
	q := typeQuery.QueryPredicate(p)

	var c TypeParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return c.TypeId.String, nil
}

func ScanType(conn *sqlx.DB, id string) (TypeParams, error) {
	p := map[string]string{
		"mz_types.id": id,
	}
	q := typeQuery.QueryPredicate(p)

	var c TypeParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}

func ListTypes(conn *sqlx.DB, schemaName, databaseName string) ([]TypeParams, error) {
	p := map[string]string{
		"mz_schemas.name":   schemaName,
		"mz_databases.name": databaseName,
	}
	q := typeQuery.QueryPredicate(p)

	var c []TypeParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
