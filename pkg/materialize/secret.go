package materialize

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// DDL
type SecretBuilder struct {
	ddl          Builder
	secretName   string
	schemaName   string
	databaseName string
	value        string
}

func NewSecretBuilder(conn *sqlx.DB, obj MaterializeObject) *SecretBuilder {
	return &SecretBuilder{
		ddl:          Builder{conn, Secret},
		secretName:   obj.Name,
		schemaName:   obj.SchemaName,
		databaseName: obj.DatabaseName,
	}
}

func (b *SecretBuilder) QualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.secretName)
}

func (b *SecretBuilder) Value(v string) *SecretBuilder {
	b.value = v
	return b
}

func (b *SecretBuilder) Create() error {
	q := fmt.Sprintf(`CREATE SECRET %s AS %s;`, b.QualifiedName(), QuoteString(b.value))
	return b.ddl.exec(q)
}

func (b *SecretBuilder) Rename(newName string) error {
	n := QualifiedName(newName)
	return b.ddl.rename(b.QualifiedName(), n)
}

func (b *SecretBuilder) UpdateValue(newValue string) error {
	q := fmt.Sprintf(`ALTER SECRET %s AS %s;`, b.QualifiedName(), QuoteString(newValue))
	return b.ddl.exec(q)
}

func (b *SecretBuilder) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

// DML
type SecretParams struct {
	SecretId     sql.NullString `db:"id"`
	SecretName   sql.NullString `db:"name"`
	SchemaName   sql.NullString `db:"schema_name"`
	DatabaseName sql.NullString `db:"database_name"`
	Comment      sql.NullString `db:"comment"`
	OwnerName    sql.NullString `db:"owner_name"`
	Privileges   sql.NullString `db:"privileges"`
}

var secretQuery = NewBaseQuery(`
	SELECT
		mz_secrets.id,
		mz_secrets.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		comments.comment AS comment,
		mz_roles.name AS owner_name,
		mz_secrets.privileges
	FROM mz_secrets
	JOIN mz_schemas
		ON mz_secrets.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_secrets.owner_id = mz_roles.id
	LEFT JOIN (
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'secret'
	) comments
		ON mz_secrets.id = comments.id`)

func SecretId(conn *sqlx.DB, obj MaterializeObject) (string, error) {
	p := map[string]string{
		"mz_secrets.name":   obj.Name,
		"mz_schemas.name":   obj.SchemaName,
		"mz_databases.name": obj.DatabaseName,
	}
	q := secretQuery.QueryPredicate(p)

	var c SecretParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return c.SecretId.String, nil
}

func ScanSecret(conn *sqlx.DB, id string) (SecretParams, error) {
	p := map[string]string{
		"mz_secrets.id": id,
	}
	q := secretQuery.QueryPredicate(p)

	var c SecretParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}

func ListSecrets(conn *sqlx.DB, schemaName, databaseName string) ([]SecretParams, error) {
	p := map[string]string{
		"mz_schemas.name":   schemaName,
		"mz_databases.name": databaseName,
	}
	q := secretQuery.QueryPredicate(p)

	var c []SecretParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
