package materialize

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// DDL
type SchemaBuilder struct {
	ddl          Builder
	schemaName   string
	databaseName string
}

func NewSchemaBuilder(conn *sqlx.DB, obj MaterializeObject) *SchemaBuilder {
	return &SchemaBuilder{
		ddl:          Builder{conn, Schema},
		schemaName:   obj.Name,
		databaseName: obj.DatabaseName,
	}
}

func (b *SchemaBuilder) QualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName)
}

func (b *SchemaBuilder) Create() error {
	q := fmt.Sprintf(`CREATE SCHEMA %s;`, b.QualifiedName())
	return b.ddl.exec(q)
}

func (b *SchemaBuilder) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

// DML
type SchemaParams struct {
	SchemaId     sql.NullString `db:"id"`
	SchemaName   sql.NullString `db:"schema_name"`
	DatabaseName sql.NullString `db:"database_name"`
	Comment      sql.NullString `db:"comment"`
	OwnerName    sql.NullString `db:"owner_name"`
	Privileges   sql.NullString `db:"privileges"`
}

var schemaQuery = NewBaseQuery(`
	SELECT
		mz_schemas.id,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		comments.comment AS comment,
		mz_roles.name AS owner_name,
		mz_schemas.privileges
	FROM mz_schemas
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_schemas.owner_id = mz_roles.id
	LEFT JOIN (
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'schema'
	) comments
		ON mz_schemas.id = comments.id`)

func SchemaId(conn *sqlx.DB, obj MaterializeObject) (string, error) {
	p := map[string]string{
		"mz_schemas.name":   obj.Name,
		"mz_databases.name": obj.DatabaseName,
	}
	q := schemaQuery.QueryPredicate(p)

	var c SchemaParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return c.SchemaId.String, nil
}

func ScanSchema(conn *sqlx.DB, id string) (SchemaParams, error) {
	p := map[string]string{
		"mz_schemas.id": id,
	}
	q := schemaQuery.QueryPredicate(p)

	var c SchemaParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}

func ListSchemas(conn *sqlx.DB, databaseName string) ([]SchemaParams, error) {
	p := map[string]string{"mz_databases.name": databaseName}
	q := schemaQuery.QueryPredicate(p)

	var c []SchemaParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
