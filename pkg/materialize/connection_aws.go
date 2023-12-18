package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type ConnectionAwsBuilder struct {
	Connection
	endpoint              string
	region                string
	accessKeyId           ValueSecretStruct
	secretAccessKey       IdentifierSchemaStruct
	sessionToken          ValueSecretStruct
	assumeRoleArn         string
	assumeRoleSessionName string
	validate              bool
}

func NewConnectionAwsBuilder(conn *sqlx.DB, obj MaterializeObject) *ConnectionAwsBuilder {
	b := Builder{conn, BaseConnection}
	return &ConnectionAwsBuilder{
		Connection: Connection{b, obj.Name, obj.SchemaName, obj.DatabaseName},
	}
}

func (b *ConnectionAwsBuilder) Endpoint(s string) *ConnectionAwsBuilder {
	b.endpoint = s
	return b
}

func (b *ConnectionAwsBuilder) Region(s string) *ConnectionAwsBuilder {
	b.region = s
	return b
}

func (b *ConnectionAwsBuilder) AccessKeyId(s ValueSecretStruct) *ConnectionAwsBuilder {
	b.accessKeyId = s
	return b
}

func (b *ConnectionAwsBuilder) SecretAccessKey(s IdentifierSchemaStruct) *ConnectionAwsBuilder {
	b.secretAccessKey = s
	return b
}

func (b *ConnectionAwsBuilder) SessionToken(s ValueSecretStruct) *ConnectionAwsBuilder {
	b.sessionToken = s
	return b
}

func (b *ConnectionAwsBuilder) AssumeRoleArn(s string) *ConnectionAwsBuilder {
	b.assumeRoleArn = s
	return b
}

func (b *ConnectionAwsBuilder) AssumeRoleSessionName(s string) *ConnectionAwsBuilder {
	b.assumeRoleSessionName = s
	return b
}

func (b *ConnectionAwsBuilder) Validate(validate bool) *ConnectionAwsBuilder {
	b.validate = validate
	return b
}

func (b *ConnectionAwsBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO AWS`, b.QualifiedName()))

	w := []string{}
	if b.endpoint != "" {
		o := fmt.Sprintf(` ENDPOINT = %s`, QuoteString(b.endpoint))
		w = append(w, o)
	}
	if b.region != "" {
		o := fmt.Sprintf(` REGION = %s`, QuoteString(b.region))
		w = append(w, o)
	}

	if b.accessKeyId.Text != "" {
		o := fmt.Sprintf(` ACCESS KEY ID = %s`, QuoteString(b.accessKeyId.Text))
		w = append(w, o)
	} else if b.accessKeyId.Secret.Name != "" {
		o := fmt.Sprintf(` ACCESS KEY ID = SECRET %s`, b.accessKeyId.Secret.QualifiedName())
		w = append(w, o)
	}
	if b.secretAccessKey.Name != "" {
		o := fmt.Sprintf(` SECRET ACCESS KEY = SECRET %s`, b.secretAccessKey.QualifiedName())
		w = append(w, o)
	}
	if b.sessionToken.Text != "" {
		o := fmt.Sprintf(` SESSION TOKEN = %s`, QuoteString(b.sessionToken.Text))
		w = append(w, o)
	} else if b.sessionToken.Secret.Name != "" {
		o := fmt.Sprintf(` SESSION TOKEN = SECRET %s`, b.sessionToken.Secret.QualifiedName())
		w = append(w, o)
	}

	if b.assumeRoleArn != "" {
		o := fmt.Sprintf(` ASSUME ROLE ARN = %s`, QuoteString(b.assumeRoleArn))
		w = append(w, o)
	}
	if b.assumeRoleSessionName != "" {
		o := fmt.Sprintf(` ASSUME ROLE SESSION NAME = %s`, QuoteString(b.assumeRoleSessionName))
		w = append(w, o)
	}

	if !b.validate {
		w = append(w, " VALIDATE = false")
	}

	f := strings.Join(w, ", ")
	q.WriteString(fmt.Sprintf(` WITH (%s)`, f))
	return b.ddl.exec(q.String())
}

type ConnectionAwsParams struct {
	ConnectionId            sql.NullString `db:"id"`
	ConnectionName          sql.NullString `db:"connection_name"`
	SchemaName              sql.NullString `db:"schema_name"`
	DatabaseName            sql.NullString `db:"database_name"`
	Endpoint                sql.NullString `db:"endpoint"`
	Region                  sql.NullString `db:"region"`
	AccessKeyId             sql.NullString `db:"access_key_id"`
	AccessKeyIdSecretId     sql.NullString `db:"access_key_id_secret_id"`
	SecretAccessKeySecretId sql.NullString `db:"secret_access_key_secret_id"`
	SessionToken            sql.NullString `db:"session_token"`
	SessionTokenSecretId    sql.NullString `db:"session_token_secret_id"`
	AssumeRoleArn           sql.NullString `db:"assume_role_arn"`
	AssumeRoleSessionName   sql.NullString `db:"assume_role_session_name"`
	Comment                 sql.NullString `db:"comment"`
	Principal               sql.NullString `db:"principal"`
	OwnerName               sql.NullString `db:"owner_name"`
}

var connectionAwsQuery = NewBaseQuery(`
	SELECT
		mz_connections.id,
		mz_connections.name AS connection_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_aws_connections.endpoint,
		mz_aws_connections.region,
		mz_aws_connections.access_key_id,
		mz_aws_connections.access_key_id_secret_id,
		mz_aws_connections.secret_access_key_secret_id,
		mz_aws_connections.session_token,
		mz_aws_connections.session_token_secret_id,
		mz_aws_connections.assume_role_arn,
		mz_aws_connections.assume_role_session_name,
		comments.comment AS comment,
		mz_aws_connections.principal,
		mz_roles.name AS owner_name
	FROM mz_connections
	JOIN mz_schemas
		ON mz_connections.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_aws_connections
		ON mz_connections.id = mz_aws_connections.id
	JOIN mz_roles
		ON mz_connections.owner_id = mz_roles.id
	LEFT JOIN (
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'connection'
	) comments
		ON mz_connections.id = comments.id`)

func ScanConnectionAws(conn *sqlx.DB, id string) (ConnectionAwsParams, error) {
	q := connectionAwsQuery.QueryPredicate(map[string]string{"mz_connections.id": id})

	var c ConnectionAwsParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
