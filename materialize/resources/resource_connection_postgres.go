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
	"connection_type": {
		Description: "The type of connection.",
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
	"user":                      ValueSecretSchema("user", "The Postgres database username.", true, false),
	"password":                  IdentifierSchema("password", "The Postgres database password.", false, true),
	"ssh_tunnel":                IdentifierSchema("ssh_tunnel", "The SSH tunnel configuration for the Postgres database.", false, true),
	"ssl_certificate_authority": ValueSecretSchema("ssl_certificate_authority", "The CA certificate for the Postgres database.", false, true),
	"ssl_certificate":           ValueSecretSchema("ssl_certificate", "The client certificate for the Postgres database.", false, true),
	"ssl_key":                   IdentifierSchema("ssl_key", "The client key for the Postgres database.", false, true),
	"ssl_mode": {
		Description: "The SSL mode for the Postgres database.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"aws_privatelink": IdentifierSchema("aws_privatelink", "The AWS PrivateLink configuration for the Postgres database.", false, true),
}

func ConnectionPostgres() *schema.Resource {
	return &schema.Resource{
		Description: "The connection resource allows you to manage connections in Materialize.",

		CreateContext: connectionPostgresCreate,
		ReadContext:   ConnectionRead,
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
	postgresUser           ValueSecretStruct
	postgresPassword       IdentifierSchemaStruct
	postgresSSHTunnel      IdentifierSchemaStruct
	postgresSSLCa          ValueSecretStruct
	postgresSSLCert        ValueSecretStruct
	postgresSSLKey         IdentifierSchemaStruct
	postgresSSLMode        string
	postgresAWSPrivateLink IdentifierSchemaStruct
}

func (b *ConnectionPostgresBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.connectionName)
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

func (b *ConnectionPostgresBuilder) PostgresUser(postgresUser ValueSecretStruct) *ConnectionPostgresBuilder {
	b.postgresUser = postgresUser
	return b
}

func (b *ConnectionPostgresBuilder) PostgresPassword(postgresPassword IdentifierSchemaStruct) *ConnectionPostgresBuilder {
	b.postgresPassword = postgresPassword
	return b
}

func (b *ConnectionPostgresBuilder) PostgresSSHTunnel(postgresSSHTunnel IdentifierSchemaStruct) *ConnectionPostgresBuilder {
	b.postgresSSHTunnel = postgresSSHTunnel
	return b
}

func (b *ConnectionPostgresBuilder) PostgresSSLCa(postgresSSLCa ValueSecretStruct) *ConnectionPostgresBuilder {
	b.postgresSSLCa = postgresSSLCa
	return b
}

func (b *ConnectionPostgresBuilder) PostgresSSLCert(postgresSSLCert ValueSecretStruct) *ConnectionPostgresBuilder {
	b.postgresSSLCert = postgresSSLCert
	return b
}

func (b *ConnectionPostgresBuilder) PostgresSSLKey(postgresSSLKey IdentifierSchemaStruct) *ConnectionPostgresBuilder {
	b.postgresSSLKey = postgresSSLKey
	return b
}

func (b *ConnectionPostgresBuilder) PostgresSSLMode(postgresSSLMode string) *ConnectionPostgresBuilder {
	b.postgresSSLMode = postgresSSLMode
	return b
}

func (b *ConnectionPostgresBuilder) PostgresAWSPrivateLink(postgresAWSPrivateLink IdentifierSchemaStruct) *ConnectionPostgresBuilder {
	b.postgresAWSPrivateLink = postgresAWSPrivateLink
	return b
}

func (b *ConnectionPostgresBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO POSTGRES (`, b.qualifiedName()))

	q.WriteString(fmt.Sprintf(`HOST %s`, QuoteString(b.postgresHost)))
	q.WriteString(fmt.Sprintf(`, PORT %d`, b.postgresPort))
	if b.postgresUser.Text != "" {
		q.WriteString(fmt.Sprintf(`, USER %s`, QuoteString(b.postgresUser.Text)))
	}
	if b.postgresUser.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, USER SECRET %s`, QualifiedName(b.postgresUser.Secret.DatabaseName, b.postgresUser.Secret.SchemaName, b.postgresUser.Secret.Name)))
	}
	if b.postgresPassword.Name != "" {
		q.WriteString(fmt.Sprintf(`, PASSWORD SECRET %s`, QualifiedName(b.postgresPassword.DatabaseName, b.postgresPassword.SchemaName, b.postgresPassword.Name)))
	}
	if b.postgresSSLMode != "" {
		q.WriteString(fmt.Sprintf(`, SSL MODE %s`, QuoteString(b.postgresSSLMode)))
	}
	if b.postgresSSHTunnel.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSH TUNNEL %s`, QualifiedName(b.postgresSSHTunnel.DatabaseName, b.postgresSSHTunnel.SchemaName, b.postgresSSHTunnel.Name)))
	}
	if b.postgresSSLCa.Text != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY %s`, QuoteString(b.postgresSSLCa.Text)))
	}
	if b.postgresSSLCa.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY SECRET %s`, QualifiedName(b.postgresSSLCa.Secret.DatabaseName, b.postgresSSLCa.Secret.SchemaName, b.postgresSSLCa.Secret.Name)))
	}
	if b.postgresSSLCert.Text != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE  %s`, QuoteString(b.postgresSSLCert.Text)))
	}
	if b.postgresSSLCert.Secret.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE SECRET %s`, QualifiedName(b.postgresSSLCert.Secret.DatabaseName, b.postgresSSLCert.Secret.SchemaName, b.postgresSSLCert.Secret.Name)))
	}
	if b.postgresSSLKey.Name != "" {
		q.WriteString(fmt.Sprintf(`, SSL KEY SECRET %s`, QualifiedName(b.postgresSSLKey.DatabaseName, b.postgresSSLKey.SchemaName, b.postgresSSLKey.Name)))
	}
	if b.postgresAWSPrivateLink.Name != "" {
		q.WriteString(fmt.Sprintf(`, AWS PRIVATELINK %s`, QualifiedName(b.postgresAWSPrivateLink.DatabaseName, b.postgresAWSPrivateLink.SchemaName, b.postgresAWSPrivateLink.Name)))
	}

	q.WriteString(fmt.Sprintf(`, DATABASE %s`, QuoteString(b.postgresDatabase)))

	q.WriteString(`);`)
	return q.String()
}

func (b *ConnectionPostgresBuilder) Rename(newConnectionName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newConnectionName)
	return fmt.Sprintf(`ALTER CONNECTION %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *ConnectionPostgresBuilder) Drop() string {
	return fmt.Sprintf(`DROP CONNECTION %s;`, b.qualifiedName())
}

func (b *ConnectionPostgresBuilder) ReadId() string {
	return readConnectionId(b.connectionName, b.schemaName, b.databaseName)

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
		var user ValueSecretStruct
		u := v.([]interface{})[0].(map[string]interface{})
		if v, ok := u["text"]; ok {
			user.Text = v.(string)
		}
		if v, ok := u["secret"]; ok && len(v.([]interface{})) > 0 {
			user.Secret = GetIdentifierSchemaStruct(databaseName, schemaName, v)
		}
		builder.PostgresUser(user)
	}

	if v, ok := d.GetOk("password"); ok {
		pass := GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.PostgresPassword(pass)
	}

	if v, ok := d.GetOk("database"); ok {
		builder.PostgresDatabase(v.(string))
	}

	if v, ok := d.GetOk("ssl_mode"); ok {
		builder.PostgresSSLMode(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate_authority"); ok {
		var ssl_ca ValueSecretStruct
		u := v.([]interface{})[0].(map[string]interface{})
		if v, ok := u["text"]; ok {
			ssl_ca.Text = v.(string)
		}
		if v, ok := u["secret"]; ok && len(v.([]interface{})) > 0 {
			ssl_ca.Secret = GetIdentifierSchemaStruct(databaseName, schemaName, v)
		}
		builder.PostgresSSLCa(ssl_ca)
	}

	if v, ok := d.GetOk("ssl_certificate"); ok {
		var ssl_cert ValueSecretStruct
		u := v.([]interface{})[0].(map[string]interface{})
		if v, ok := u["text"]; ok {
			ssl_cert.Text = v.(string)
		}
		if v, ok := u["secret"]; ok && len(v.([]interface{})) > 0 {
			ssl_cert.Secret = GetIdentifierSchemaStruct(databaseName, schemaName, v)
		}
		builder.PostgresSSLCert(ssl_cert)
	}

	if v, ok := d.GetOk("ssl_key"); ok {
		k := GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.PostgresSSLKey(k)
	}

	if v, ok := d.GetOk("aws_privatelink"); ok {
		conn := GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.PostgresAWSPrivateLink(conn)
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		conn := GetIdentifierSchemaStruct(databaseName, schemaName, v)
		builder.PostgresSSHTunnel(conn)
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "connection"); err != nil {
		return diag.FromErr(err)
	}
	return ConnectionRead(ctx, d, meta)
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

	return ConnectionRead(ctx, d, meta)
}

func connectionPostgresDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := newConnectionPostgresBuilder(connectionName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "connection"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
