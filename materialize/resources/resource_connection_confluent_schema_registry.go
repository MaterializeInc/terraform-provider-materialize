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

var connectionConfluentSchemaRegistrySchema = map[string]*schema.Schema{
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
	"url": {
		Description: "The URL of the Confluent Schema Registry.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"ssl_certificate_authority": ValueSecretSchema("ssl_certificate_authority", "The CA certificate for the Confluent Schema Registry.", false, true),
	"ssl_certificate":           ValueSecretSchema("ssl_certificate", "The client certificate for the Confluent Schema Registry.", false, true),
	"ssl_key": {
		Description: "The client key for the Confluent Schema Registry.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"password": {
		Description:  "The password for the Confluent Schema Registry.",
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"username"},
	},
	"username": ValueSecretSchema("username", "The username for the Confluent Schema Registry.", false, true),
	"ssh_tunnel": {
		Description: "The SSH tunnel configuration for the Confluent Schema Registry.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"aws_privatelink": {
		Description: "The AWS PrivateLink configuration for the Confluent Schema Registry.",
		Type:        schema.TypeString,
		Optional:    true,
	},
}

func ConnectionConfluentSchemaRegistry() *schema.Resource {
	return &schema.Resource{
		Description: "The connection resource allows you to manage connections in Materialize.",

		CreateContext: connectionConfluentSchemaRegistryCreate,
		ReadContext:   ConnectionRead,
		UpdateContext: connectionConfluentSchemaRegistryUpdate,
		DeleteContext: connectionConfluentSchemaRegistryDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionConfluentSchemaRegistrySchema,
	}
}

type ConnectionConfluentSchemaRegistryBuilder struct {
	connectionName                        string
	schemaName                            string
	databaseName                          string
	confluentSchemaRegistryUrl            string
	confluentSchemaRegistrySSLCa          ValueSecretStruct
	confluentSchemaRegistrySSLCert        ValueSecretStruct
	confluentSchemaRegistrySSLKey         string
	confluentSchemaRegistryUsername       ValueSecretStruct
	confluentSchemaRegistryPassword       string
	confluentSchemaRegistrySSHTunnel      string
	confluentSchemaRegistryAWSPrivateLink string
}

func (b *ConnectionConfluentSchemaRegistryBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.connectionName)
}

func newConnectionConfluentSchemaRegistryBuilder(connectionName, schemaName, databaseName string) *ConnectionConfluentSchemaRegistryBuilder {
	return &ConnectionConfluentSchemaRegistryBuilder{
		connectionName: connectionName,
		schemaName:     schemaName,
		databaseName:   databaseName,
	}
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistryUrl(confluentSchemaRegistryUrl string) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistryUrl = confluentSchemaRegistryUrl
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistryUsername(confluentSchemaRegistryUsername ValueSecretStruct) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistryUsername = confluentSchemaRegistryUsername
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistryPassword(confluentSchemaRegistryPassword string) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistryPassword = confluentSchemaRegistryPassword
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistrySSLCa(confluentSchemaRegistrySSLCa ValueSecretStruct) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistrySSLCa = confluentSchemaRegistrySSLCa
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistrySSLCert(confluentSchemaRegistrySSLCert ValueSecretStruct) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistrySSLCert = confluentSchemaRegistrySSLCert
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistrySSLKey(confluentSchemaRegistrySSLKey string) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistrySSLKey = confluentSchemaRegistrySSLKey
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistrySSHTunnel(confluentSchemaRegistrySSHTunnel string) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistrySSHTunnel = confluentSchemaRegistrySSHTunnel
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistryAWSPrivateLink(confluentSchemaRegistryAWSPrivateLink string) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistryAWSPrivateLink = confluentSchemaRegistryAWSPrivateLink
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO CONFLUENT SCHEMA REGISTRY (`, b.qualifiedName()))

	q.WriteString(fmt.Sprintf(`URL '%s'`, b.confluentSchemaRegistryUrl))
	if b.confluentSchemaRegistryUsername.Text != "" {
		q.WriteString(fmt.Sprintf(`, USERNAME = %s`, QuoteString(b.confluentSchemaRegistryUsername.Text)))
	}
	if b.confluentSchemaRegistryUsername.Secret != "" {
		q.WriteString(fmt.Sprintf(`, USERNAME = SECRET %s`, b.confluentSchemaRegistryUsername.Secret))
	}
	if b.confluentSchemaRegistryPassword != "" {
		q.WriteString(fmt.Sprintf(`, PASSWORD = SECRET %s`, b.confluentSchemaRegistryPassword))
	}
	if b.confluentSchemaRegistrySSLCa.Text != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY = %s`, QuoteString(b.confluentSchemaRegistrySSLCa.Text)))
	}
	if b.confluentSchemaRegistrySSLCa.Secret != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY = SECRET %s`, b.confluentSchemaRegistrySSLCa.Secret))
	}
	if b.confluentSchemaRegistrySSLCert.Text != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE = %s`, QuoteString(b.confluentSchemaRegistrySSLCert.Text)))
	}
	if b.confluentSchemaRegistrySSLCert.Secret != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE = SECRET %s`, b.confluentSchemaRegistrySSLCert.Secret))
	}
	if b.confluentSchemaRegistrySSLKey != "" {
		q.WriteString(fmt.Sprintf(`, SSL KEY = SECRET %s`, b.confluentSchemaRegistrySSLKey))
	}
	if b.confluentSchemaRegistryAWSPrivateLink != "" {
		q.WriteString(fmt.Sprintf(`, AWS PRIVATELINK %s`, b.confluentSchemaRegistryAWSPrivateLink))
	}
	if b.confluentSchemaRegistrySSHTunnel != "" {
		q.WriteString(fmt.Sprintf(`, SSH TUNNEL %s`, b.confluentSchemaRegistrySSHTunnel))
	}

	q.WriteString(`);`)
	return q.String()
}

func (b *ConnectionConfluentSchemaRegistryBuilder) Rename(newConnectionName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newConnectionName)
	return fmt.Sprintf(`ALTER CONNECTION %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *ConnectionConfluentSchemaRegistryBuilder) Drop() string {
	return fmt.Sprintf(`DROP CONNECTION %s;`, b.qualifiedName())
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ReadId() string {
	return readConnectionId(b.connectionName, b.schemaName, b.databaseName)
}

func connectionConfluentSchemaRegistryCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newConnectionConfluentSchemaRegistryBuilder(connectionName, schemaName, databaseName)

	if v, ok := d.GetOk("url"); ok {
		builder.ConfluentSchemaRegistryUrl(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate_authority"); ok {
		var ssl_ca ValueSecretStruct
		u := v.([]interface{})[0].(map[string]interface{})
		if v, ok := u["text"]; ok {
			ssl_ca.Text = v.(string)
		}
		if v, ok := u["secret"]; ok {
			ssl_ca.Secret = v.(string)
		}
		builder.ConfluentSchemaRegistrySSLCa(ssl_ca)
	}

	if v, ok := d.GetOk("ssl_certificate"); ok {
		var ssl_cert ValueSecretStruct
		u := v.([]interface{})[0].(map[string]interface{})
		if v, ok := u["text"]; ok {
			ssl_cert.Text = v.(string)
		}
		if v, ok := u["secret"]; ok {
			ssl_cert.Secret = v.(string)
		}
		builder.ConfluentSchemaRegistrySSLCert(ssl_cert)
	}

	if v, ok := d.GetOk("ssl_key"); ok {
		builder.ConfluentSchemaRegistrySSLKey(v.(string))
	}

	if v, ok := d.GetOk("username"); ok {
		var user ValueSecretStruct
		u := v.([]interface{})[0].(map[string]interface{})
		if v, ok := u["text"]; ok {
			user.Text = v.(string)
		}
		if v, ok := u["secret"]; ok {
			user.Secret = v.(string)
		}
		builder.ConfluentSchemaRegistryUsername(user)
	}

	if v, ok := d.GetOk("password"); ok {
		builder.ConfluentSchemaRegistryPassword(v.(string))
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		builder.ConfluentSchemaRegistrySSHTunnel(v.(string))
	}

	if v, ok := d.GetOk("aws_privatelink"); ok {
		builder.ConfluentSchemaRegistryAWSPrivateLink(v.(string))
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "connection"); err != nil {
		return diag.FromErr(err)
	}
	return ConnectionRead(ctx, d, meta)
}

func connectionConfluentSchemaRegistryUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		newConnectionName := d.Get("name").(string)
		q := newConnectionConfluentSchemaRegistryBuilder(connectionName, schemaName, databaseName).Rename(newConnectionName)
		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return ConnectionRead(ctx, d, meta)
}

func connectionConfluentSchemaRegistryDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := newConnectionConfluentSchemaRegistryBuilder(connectionName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "connection"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
