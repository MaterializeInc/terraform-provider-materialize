package materialize

import (
	"fmt"
)

type DatabaseBuilder struct {
	databaseName string
}

func NewDatabaseBuilder(databaseName string) *DatabaseBuilder {
	return &DatabaseBuilder{
		databaseName: databaseName,
	}
}

func (b *DatabaseBuilder) QualifiedName() string {
	return QualifiedName(b.databaseName)
}

func (b *DatabaseBuilder) Create() string {
	return fmt.Sprintf(`CREATE DATABASE %s;`, b.QualifiedName())
}

func (b *DatabaseBuilder) Drop() string {
	return fmt.Sprintf(`DROP DATABASE %s;`, b.QualifiedName())
}

func (b *DatabaseBuilder) ReadId() string {
	return fmt.Sprintf(`SELECT id FROM mz_databases WHERE name = %s;`, QuoteString(b.databaseName))
}

func ReadDatabaseParams(id string) string {
	return fmt.Sprintf("SELECT name FROM mz_databases WHERE id = %s;", QuoteString(id))
}

func ReadDatabaseDatasource() string {
	return "SELECT id, name FROM mz_databases;"
}
