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
	"ssl_certificate_authority": {
		Description: "The CA certificate for the Confluent Schema Registry.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"ssl_certificate": {
		Description: "The client certificate for the Confluent Schema Registry.",
		Type:        schema.TypeString,
		Optional:    true,
	},
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
	"username": {
		Description:  "The username for the Confluent Schema Registry.",
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"password"},
	},
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
	confluentSchemaRegistrySSLCa          string
	confluentSchemaRegistrySSLCert        string
	confluentSchemaRegistrySSLKey         string
	confluentSchemaRegistryUsername       string
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

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistryUsername(confluentSchemaRegistryUsername string) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistryUsername = confluentSchemaRegistryUsername
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistryPassword(confluentSchemaRegistryPassword string) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistryPassword = confluentSchemaRegistryPassword
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistrySSLCa(confluentSchemaRegistrySSLCa string) *ConnectionConfluentSchemaRegistryBuilder {
	b.confluentSchemaRegistrySSLCa = confluentSchemaRegistrySSLCa
	return b
}

func (b *ConnectionConfluentSchemaRegistryBuilder) ConfluentSchemaRegistrySSLCert(confluentSchemaRegistrySSLCert string) *ConnectionConfluentSchemaRegistryBuilder {
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
	if b.confluentSchemaRegistryUsername != "" {
		q.WriteString(fmt.Sprintf(`, USERNAME = '%s'`, b.confluentSchemaRegistryUsername))
	}
	if b.confluentSchemaRegistryPassword != "" {
		q.WriteString(fmt.Sprintf(`, PASSWORD = SECRET %s`, b.confluentSchemaRegistryPassword))
	}
	if b.confluentSchemaRegistrySSLCa != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE AUTHORITY = %s`, b.confluentSchemaRegistrySSLCa))
	}
	if b.confluentSchemaRegistrySSLCert != "" {
		q.WriteString(fmt.Sprintf(`, SSL CERTIFICATE = %s`, b.confluentSchemaRegistrySSLCert))
	}
	if b.confluentSchemaRegistrySSLKey != "" {
		q.WriteString(fmt.Sprintf(`, SSL KEY = %s`, b.confluentSchemaRegistrySSLKey))
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
		builder.ConfluentSchemaRegistrySSLCa(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate"); ok {
		builder.ConfluentSchemaRegistrySSLCert(v.(string))
	}

	if v, ok := d.GetOk("ssl_key"); ok {
		builder.ConfluentSchemaRegistrySSLKey(v.(string))
	}

	if v, ok := d.GetOk("username"); ok {
		builder.ConfluentSchemaRegistryUsername(v.(string))
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
