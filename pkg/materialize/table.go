package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type TableColumn struct {
	ColName string
	ColType string
	NotNull bool
}

func GetTableColumnStruct(v []interface{}) []TableColumn {
	var columns []TableColumn
	for _, column := range v {
		c := column.(map[string]interface{})
		columns = append(columns, TableColumn{
			ColName: c["name"].(string),
			ColType: c["type"].(string),
			NotNull: !c["nullable"].(bool),
		})
	}
	return columns
}

type TableBuilder struct {
	ddl          Builder
	tableName    string
	schemaName   string
	databaseName string
	column       []TableColumn
}

func NewTableBuilder(conn *sqlx.DB, obj ObjectSchemaStruct) *TableBuilder {
	return &TableBuilder{
		ddl:          Builder{conn, Table},
		tableName:    obj.Name,
		schemaName:   obj.SchemaName,
		databaseName: obj.DatabaseName,
	}
}

func (b *TableBuilder) QualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.tableName)
}

func (b *TableBuilder) Column(c []TableColumn) *TableBuilder {
	b.column = c
	return b
}

func (b *TableBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE TABLE %s`, b.QualifiedName()))

	var column []string
	for _, c := range b.column {
		s := strings.Builder{}

		s.WriteString(fmt.Sprintf(`%s %s`, c.ColName, c.ColType))
		if c.NotNull {
			s.WriteString(` NOT NULL`)
		}
		o := s.String()
		column = append(column, o)

	}
	p := strings.Join(column[:], ", ")
	q.WriteString(fmt.Sprintf(` (%s);`, p))

	return b.ddl.exec(q.String())
}

func (b *TableBuilder) Rename(newName string) error {
	n := QualifiedName(newName)
	return b.ddl.rename(b.QualifiedName(), n)
}

func (b *TableBuilder) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

type TableParams struct {
	TableId      sql.NullString `db:"id"`
	TableName    sql.NullString `db:"name"`
	SchemaName   sql.NullString `db:"schema_name"`
	DatabaseName sql.NullString `db:"database_name"`
	OwnerName    sql.NullString `db:"owner_name"`
	Privileges   sql.NullString `db:"privileges"`
}

var tableQuery = NewBaseQuery(`
	SELECT
		mz_tables.id,
		mz_tables.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_roles.name AS owner_name,
		mz_tables.privileges
	FROM mz_tables
	JOIN mz_schemas
		ON mz_tables.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_tables.owner_id = mz_roles.id`)

func TableId(conn *sqlx.DB, obj ObjectSchemaStruct) (string, error) {
	p := map[string]string{
		"mz_tables.name":    obj.Name,
		"mz_schemas.name":   obj.SchemaName,
		"mz_databases.name": obj.DatabaseName,
	}
	q := tableQuery.QueryPredicate(p)

	var c TableParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return c.TableId.String, nil
}

func ScanTable(conn *sqlx.DB, id string) (TableParams, error) {
	p := map[string]string{
		"mz_tables.id": id,
	}
	q := tableQuery.QueryPredicate(p)

	var c TableParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}

func ListTables(conn *sqlx.DB, schemaName, databaseName string) ([]TableParams, error) {
	p := map[string]string{
		"mz_schemas.name":   schemaName,
		"mz_databases.name": databaseName,
	}
	q := tableQuery.QueryPredicate(p)

	var c []TableParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
