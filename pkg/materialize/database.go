package materialize

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

type DatabaseBuilder struct {
	ddl          Builder
	databaseName string
}

func NewDatabaseBuilder(conn *sqlx.DB, databaseName string) *DatabaseBuilder {
	return &DatabaseBuilder{
		ddl:          Builder{conn, Database},
		databaseName: databaseName,
	}
}

func (b *DatabaseBuilder) QualifiedName() string {
	return QualifiedName(b.databaseName)
}

func (b *DatabaseBuilder) Create() error {
	q := fmt.Sprintf(`CREATE DATABASE %s;`, b.QualifiedName())
	return b.ddl.exec(q)
}

func (b *DatabaseBuilder) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

type DatabaseParams struct {
	DatabaseId   sql.NullString `db:"id"`
	DatabaseName sql.NullString `db:"database_name"`
}

var databaseQuery = "SELECT id, name AS database_name FROM mz_databases"

func DatabaseId(conn *sqlx.DB, databaseName string) (string, error) {
	p := map[string]string{"name": databaseName}
	q := queryPredicate(databaseQuery, p)

	var c DatabaseParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return c.DatabaseId.String, nil
}

func ScanDatabase(conn *sqlx.DB, id string) (DatabaseParams, error) {
	p := map[string]string{"id": id}
	q := queryPredicate(databaseQuery, p)

	var c DatabaseParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}

func ListDatabases(conn *sqlx.DB) ([]DatabaseParams, error) {
	p := map[string]string{}
	q := queryPredicate(databaseQuery, p)

	var c []DatabaseParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
