package resources

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var connectionPostgresSchema = map[string]*schema.Schema{
	"name": {
		Description: "The name of the connection.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"schema_name": {
		Description: "The identifier for the connection schema.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "public",
	},
	"database_name": {
		Description: "The identifier for the connection database.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "materialize",
	},
	"qualified_name": {
		Description: "The fully qualified name of the connection.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"database": {
		Description: "The target Postgres database.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"host": {
		Description: "The Postgres database hostname.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"port": {
		Description: "The Postgres database port.",
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     5432,
	},
	"user": {
		Description: "The Postgres database username.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"password": {
		Description: "The Postgres database password.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"ssh_tunnel": {
		Description: "The SSH tunnel configuration for the Postgres database.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"ssl_certificate_authority": {
		Description: "The CA certificate for the Postgres database.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"ssl_certificate": {
		Description: "The client certificate for the Postgres database.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"ssl_key": {
		Description: "The client key for the Postgres database.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"ssl_mode": {
		Description: "The SSL mode for the Postgres database.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"aws_privatelink": {
		Description: "The AWS PrivateLink configuration for the Postgres database.",
		Type:        schema.TypeString,
		Optional:    true,
	},
}

func ConnectionPostgres() *schema.Resource {
	return &schema.Resource{
		Description: "The connection resource allows you to manage connections in Materialize.",

		CreateContext: connectionPostgresCreate,
		ReadContext:   connectionPostgresRead,
		UpdateContext: connectionPostgresUpdate,
		DeleteContext: connectionPostgresDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionPostgresSchema,
	}
}

type ConnectionPostgresBuilder struct {
	connectionName         string
	schemaName             string
	databaseName           string
	connectionType         string
	postgresDatabase       string
	postgresHost           string
	postgresPort           int
	postgresUser           string
	postgresPassword       string
	postgresSSHTunnel      string
	postgresSSLCa          string
	postgresSSLCert        string
	postgresSSLKey         string
	postgresSSLMode        string
	postgresAWSPrivateLink string
}

func newConnectionPostgresBuilder(connectionName, schemaName, databaseName string) *ConnectionPostgresBuilder {
	return &ConnectionPostgresBuilder{
		connectionName: connectionName,
		schemaName:     schemaName,
		databaseName:   databaseName,
	}
}

func (b *ConnectionPostgresBuilder) ConnectionType(connectionType string) *ConnectionPostgresBuilder {
	b.connectionType = connectionType
	return b
}

func (b *ConnectionPostgresBuilder) PostgresDatabase(postgresDatabase string) *ConnectionPostgresBuilder {
	b.postgresDatabase = postgresDatabase
	return b
}

func (b *ConnectionPostgresBuilder) PostgresHost(postgresHost string) *ConnectionPostgresBuilder {
	b.postgresHost = postgresHost
	return b
}

func (b *ConnectionPostgresBuilder) PostgresPort(postgresPort int) *ConnectionPostgresBuilder {
	b.postgresPort = postgresPort
	return b
}

func (b *ConnectionPostgresBuilder) PostgresUser(postgresUser string) *ConnectionPostgresBuilder {
	b.postgresUser = postgresUser
	return b
}

func (b *ConnectionPostgresBuilder) PostgresPassword(postgresPassword string) *ConnectionPostgresBuilder {
	b.postgresPassword = postgresPassword
	return b
}

func (b *ConnectionPostgresBuilder) PostgresSSHTunnel(postgresSSHTunnel string) *ConnectionPostgresBuilder {
	b.postgresSSHTunnel = postgresSSHTunnel
	return b
}

func (b *ConnectionPostgresBuilder) PostgresSSLCa(postgresSSLCa string) *ConnectionPostgresBuilder {
	b.postgresSSLCa = postgresSSLCa
	return b
}

func (b *ConnectionPostgresBuilder) PostgresSSLCert(postgresSSLCert string) *ConnectionPostgresBuilder {
	b.postgresSSLCert = postgresSSLCert
	return b
}

func (b *ConnectionPostgresBuilder) PostgresSSLKey(postgresSSLKey string) *ConnectionPostgresBuilder {
	b.postgresSSLKey = postgresSSLKey
	return b
}

func (b *ConnectionPostgresBuilder) PostgresSSLMode(postgresSSLMode string) *ConnectionPostgresBuilder {
	b.postgresSSLMode = postgresSSLMode
	return b
}

func (b *ConnectionPostgresBuilder) PostgresAWSPrivateLink(postgresAWSPrivateLink string) *ConnectionPostgresBuilder {
	b.postgresAWSPrivateLink = postgresAWSPrivateLink
	return b
}

func (b *ConnectionPostgresBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s.%s.%s TO POSTGRES (`, b.databaseName, b.schemaName, b.connectionName))

	q.WriteString(fmt.Sprintf(`HOST '%s'`, b.postgresHost))
	q.WriteString(fmt.Sprintf(`, PORT %d`, b.postgresPort))
	q.WriteString(fmt.Sprintf(`, USER '%s'`, b.postgresUser))
	if b.postgresPassword != "" {
		q.WriteString(fmt.Sprintf(`, PASSWORD SECRET %s`, b.postgresPassword))
	}
	if b.postgresSSLMode != "" {
		q.WriteString(fmt.Sprintf(`, SSL MODE '%s'`, b.postgresSSLMode))
	}
	if b.postgresSSHTunnel != "" {
		q.WriteString(fmt.Sprintf(`, SSH TUNNEL '%s'`, b.postgresSSHTunnel))
	}
	if b.postgresSSLCa != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY SECRET %s`, b.postgresSSLCa))
	}
	if b.postgresSSLCert != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE SECRET %s`, b.postgresSSLCert))
	}
	if b.postgresSSLKey != "" {
		q.WriteString(fmt.Sprintf(`, SSL KEY SECRET %s`, b.postgresSSLKey))
	}
	if b.postgresAWSPrivateLink != "" {
		q.WriteString(fmt.Sprintf(`, AWS PRIVATELINK %s`, b.postgresAWSPrivateLink))
	}

	q.WriteString(fmt.Sprintf(`, DATABASE '%s'`, b.postgresDatabase))

	q.WriteString(`);`)
	return q.String()
}

func (b *ConnectionPostgresBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_connections.id
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_connections.name = '%s'
		AND mz_schemas.name = '%s'
		AND mz_databases.name = '%s';
	`, b.connectionName, b.schemaName, b.databaseName)
}

func (b *ConnectionPostgresBuilder) Rename(newConnectionName string) string {
	return fmt.Sprintf(`ALTER CONNECTION %s.%s.%s RENAME TO %s.%s.%s;`, b.databaseName, b.schemaName, b.connectionName, b.databaseName, b.schemaName, newConnectionName)
}

func (b *ConnectionPostgresBuilder) Drop() string {
	return fmt.Sprintf(`DROP CONNECTION %s.%s.%s;`, b.databaseName, b.schemaName, b.connectionName)
}

func connectionPostgresCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newConnectionPostgresBuilder(connectionName, schemaName, databaseName)

	if v, ok := d.GetOk("connection_type"); ok {
		builder.ConnectionType(v.(string))
	}

	if v, ok := d.GetOk("host"); ok {
		builder.PostgresHost(v.(string))
	}

	if v, ok := d.GetOk("port"); ok {
		builder.PostgresPort(v.(int))
	}

	if v, ok := d.GetOk("user"); ok {
		builder.PostgresUser(v.(string))
	}

	if v, ok := d.GetOk("password"); ok {
		builder.PostgresPassword(v.(string))
	}

	if v, ok := d.GetOk("database"); ok {
		builder.PostgresDatabase(v.(string))
	}

	if v, ok := d.GetOk("ssl_mode"); ok {
		builder.PostgresSSLMode(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate_authority"); ok {
		builder.PostgresSSLCa(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate"); ok {
		builder.PostgresSSLCert(v.(string))
	}

	if v, ok := d.GetOk("ssl_key"); ok {
		builder.PostgresSSLKey(v.(string))
	}

	if v, ok := d.GetOk("aws_privatelink"); ok {
		builder.PostgresAWSPrivateLink(v.(string))
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		builder.PostgresSSHTunnel(v.(string))
	}

	qc := builder.Create()
	qr := builder.ReadId()

	createResource(conn, d, qc, qr, "connection")
	return connectionPostgresRead(ctx, d, meta)
}

func connectionPostgresRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readConnectionParams(i)

	readResource(conn, d, i, q, _connection{}, "connection")
	setQualifiedName(d)
	return nil
}

func connectionPostgresUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		newConnectionName := d.Get("name").(string)
		q := newConnectionPostgresBuilder(connectionName, schemaName, databaseName).Rename(newConnectionName)
		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return connectionPostgresRead(ctx, d, meta)
}

func connectionPostgresDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newConnectionPostgresBuilder(connectionName, schemaName, databaseName)
	q := builder.Drop()

	dropResource(conn, d, q, "connection")
	return nil
}
