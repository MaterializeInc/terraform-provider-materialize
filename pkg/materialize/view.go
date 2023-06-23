package materialize

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// DDL
type ViewBuilder struct {
	ddl          Builder
	viewName     string
	schemaName   string
	databaseName string
	selectStmt   string
}

func NewViewBuilder(conn *sqlx.DB, obj ObjectSchemaStruct) *ViewBuilder {
	return &ViewBuilder{
		ddl:          Builder{conn, View},
		viewName:     obj.Name,
		schemaName:   obj.SchemaName,
		databaseName: obj.DatabaseName,
	}
}

func (b *ViewBuilder) QualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.viewName)
}

func (b *ViewBuilder) SelectStmt(selectStmt string) *ViewBuilder {
	b.selectStmt = selectStmt
	return b
}

func (b *ViewBuilder) Create() error {
	q := fmt.Sprintf(`CREATE VIEW %s AS %s;`, b.QualifiedName(), b.selectStmt)
	return b.ddl.exec(q)
}

func (b *ViewBuilder) Rename(newName string) error {
	n := QualifiedName(newName)
	return b.ddl.rename(b.QualifiedName(), n)
}

func (b *ViewBuilder) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

// DML
type ViewParams struct {
	ViewId       sql.NullString `db:"id"`
	ViewName     sql.NullString `db:"name"`
	SchemaName   sql.NullString `db:"schema_name"`
	DatabaseName sql.NullString `db:"database_name"`
	OwnerName    sql.NullString `db:"owner_name"`
	Privileges   sql.NullString `db:"privileges"`
}

var viewQuery = NewBaseQuery(`
	SELECT
		mz_views.id,
		mz_views.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_roles.name AS owner_name,
		mz_views.privileges
	FROM mz_views
	JOIN mz_schemas
		ON mz_views.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_views.owner_id = mz_roles.id`)

func ViewId(conn *sqlx.DB, obj ObjectSchemaStruct) (string, error) {
	p := map[string]string{
		"mz_views.name":     obj.Name,
		"mz_schemas.name":   obj.SchemaName,
		"mz_databases.name": obj.DatabaseName,
	}
	q := viewQuery.QueryPredicate(p)

	var c ViewParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return c.ViewId.String, nil
}

func ScanView(conn *sqlx.DB, id string) (ViewParams, error) {
	p := map[string]string{
		"mz_views.id": id,
	}
	q := viewQuery.QueryPredicate(p)

	var c ViewParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}

func ListViews(conn *sqlx.DB, schemaName, databaseName string) ([]ViewParams, error) {
	p := map[string]string{
		"mz_schemas.name":   schemaName,
		"mz_databases.name": databaseName,
	}
	q := viewQuery.QueryPredicate(p)

	var c []ViewParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
