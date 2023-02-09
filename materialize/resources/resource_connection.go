package resources

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

func Connection() *schema.Resource {
	return &schema.Resource{
		Description: "The connection resource allows you to manage connections in Materialize.",

		CreateContext: resourceConnectionCreate,
		ReadContext:   resourceConnectionRead,
		UpdateContext: resourceConnectionUpdate,
		DeleteContext: resourceConnectionDelete,

		Schema: map[string]*schema.Schema{
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
			"connection_type": {
				Description:  "The type of the connection.",
				Type:         schema.TypeString,
				Required:     true,
				ValidateFunc: validation.StringInSlice(connectionTypes, true),
			},
			"ssh_host": {
				Description:  "The host of the SSH tunnel.",
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"ssh_user", "ssh_port"},
			},
			"ssh_user": {
				Description:  "The user of the SSH tunnel.",
				Type:         schema.TypeString,
				Optional:     true,
				RequiredWith: []string{"ssh_host", "ssh_port"},
			},
			"ssh_port": {
				Description:  "The port of the SSH tunnel.",
				Type:         schema.TypeInt,
				Optional:     true,
				RequiredWith: []string{"ssh_host", "ssh_user"},
			},
			"aws_privatelink_service_name": {
				Description:   "The name of the AWS PrivateLink service.",
				Type:          schema.TypeString,
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"ssh_host", "ssh_user", "ssh_port"},
				RequiredWith:  []string{"aws_privatelink_availability_zones"},
			},
			"aws_privatelink_availability_zones": {
				Description:   "The availability zones of the AWS PrivateLink service.",
				Type:          schema.TypeList,
				Elem:          &schema.Schema{Type: schema.TypeString},
				Optional:      true,
				ForceNew:      true,
				ConflictsWith: []string{"ssh_host", "ssh_user", "ssh_port"},
				RequiredWith:  []string{"aws_privatelink_service_name"},
			},
			"postgres_database": {
				Description: "The target Postgres database.",
				Type:        schema.TypeString,
				Optional:    true,
				RequiredWith: []string{
					"postgres_host",
					"postgres_port",
					"postgres_user",
					"postgres_password",
				},
			},
			"postgres_host": {
				Description: "The Postgres database hostname.",
				Type:        schema.TypeString,
				Optional:    true,
				RequiredWith: []string{
					"postgres_database",
					"postgres_port",
					"postgres_user",
					"postgres_password",
				},
			},
			"postgres_port": {
				Description: "The Postgres database port.",
				Type:        schema.TypeInt,
				Optional:    true,
				RequiredWith: []string{
					"postgres_database",
					"postgres_host",
					"postgres_user",
					"postgres_password",
				},
				Default: 5432,
			},
			"postgres_user": {
				Description: "The Postgres database username.",
				Type:        schema.TypeString,
				Optional:    true,
				RequiredWith: []string{
					"postgres_database",
					"postgres_host",
					"postgres_port",
					"postgres_password",
				},
			},
			"postgres_password": {
				Description: "The Postgres database password.",
				Type:        schema.TypeString,
				Optional:    true,
				RequiredWith: []string{
					"postgres_database",
					"postgres_host",
					"postgres_port",
					"postgres_user",
				},
			},
			"postgres_ssh_tunnel": {
				Description: "The SSH tunnel configuration for the Postgres database.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"postgres_ssl_ca": {
				Description: "The CA certificate for the Postgres database.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"postgres_ssl_cert": {
				Description: "The client certificate for the Postgres database.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"postgres_ssl_key": {
				Description: "The client key for the Postgres database.",
				Type:        schema.TypeString,
				Optional:    true,
			},
			"postgres_ssl_mode": {
				Description: "The SSL mode for the Postgres database.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "disable",
			},
			"postgres_aws_privatelink": {
				Description: "The AWS PrivateLink configuration for the Postgres database.",
				Type:        schema.TypeString,
				Optional:    true,
			},
		},
	}
}

type ConnectionBuilder struct {
	connectionName               string
	schemaName                   string
	connectionType               string
	sshHost                      string
	sshUser                      string
	sshPort                      int
	privateLinkServiceName       string
	privateLinkAvailabilityZones []string
	postgresDatabase             string
	postgresHost                 string
	postgresPort                 int
	postgresUser                 string
	postgresPassword             string
	postgresSSHTunnel            string
	postgresSSLCa                string
	postgresSSLCert              string
	postgresSSLKey               string
	postgresSSLMode              string
	postgresAWSPrivateLink       string
}

func newConnectionBuilder(connectionName, schemaName string) *ConnectionBuilder {
	return &ConnectionBuilder{
		connectionName: connectionName,
		schemaName:     schemaName,
	}
}

func (b *ConnectionBuilder) ConnectionName(connectionName string) *ConnectionBuilder {
	b.connectionName = connectionName
	return b
}

func (b *ConnectionBuilder) SchemaName(schemaName string) *ConnectionBuilder {
	b.schemaName = schemaName
	return b
}

func (b *ConnectionBuilder) ConnectionType(connectionType string) *ConnectionBuilder {
	b.connectionType = connectionType
	return b
}

func (b *ConnectionBuilder) SSHHost(sshHost string) *ConnectionBuilder {
	b.sshHost = sshHost
	return b
}

func (b *ConnectionBuilder) SSHUser(sshUser string) *ConnectionBuilder {
	b.sshUser = sshUser
	return b
}

func (b *ConnectionBuilder) SSHPort(sshPort int) *ConnectionBuilder {
	b.sshPort = sshPort
	return b
}

func (b *ConnectionBuilder) PrivateLinkServiceName(privateLinkServiceName string) *ConnectionBuilder {
	b.privateLinkServiceName = privateLinkServiceName
	return b
}

func (b *ConnectionBuilder) PrivateLinkAvailabilityZones(privateLinkAvailabilityZones []string) *ConnectionBuilder {
	b.privateLinkAvailabilityZones = privateLinkAvailabilityZones
	return b
}

func (b *ConnectionBuilder) PostgresDatabase(postgresDatabase string) *ConnectionBuilder {
	b.postgresDatabase = postgresDatabase
	return b
}

func (b *ConnectionBuilder) PostgresHost(postgresHost string) *ConnectionBuilder {
	b.postgresHost = postgresHost
	return b
}

func (b *ConnectionBuilder) PostgresPort(postgresPort int) *ConnectionBuilder {
	b.postgresPort = postgresPort
	return b
}

func (b *ConnectionBuilder) PostgresUser(postgresUser string) *ConnectionBuilder {
	b.postgresUser = postgresUser
	return b
}

func (b *ConnectionBuilder) PostgresPassword(postgresPassword string) *ConnectionBuilder {
	b.postgresPassword = postgresPassword
	return b
}

func (b *ConnectionBuilder) PostgresSSHTunnel(postgresSSHTunnel string) *ConnectionBuilder {
	b.postgresSSHTunnel = postgresSSHTunnel
	return b
}

func (b *ConnectionBuilder) PostgresSSLCa(postgresSSLCa string) *ConnectionBuilder {
	b.postgresSSLCa = postgresSSLCa
	return b
}

func (b *ConnectionBuilder) PostgresSSLCert(postgresSSLCert string) *ConnectionBuilder {
	b.postgresSSLCert = postgresSSLCert
	return b
}

func (b *ConnectionBuilder) PostgresSSLKey(postgresSSLKey string) *ConnectionBuilder {
	b.postgresSSLKey = postgresSSLKey
	return b
}

func (b *ConnectionBuilder) PostgresSSLMode(postgresSSLMode string) *ConnectionBuilder {
	b.postgresSSLMode = postgresSSLMode
	return b
}

func (b *ConnectionBuilder) PostgresAWSPrivateLink(postgresAWSPrivateLink string) *ConnectionBuilder {
	b.postgresAWSPrivateLink = postgresAWSPrivateLink
	return b
}

func (b *ConnectionBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s.%s`, b.schemaName, b.connectionName))

	q.WriteString(fmt.Sprintf(` TO %s (`, b.connectionType))

	if b.connectionType == "SSH TUNNEL" {
		q.WriteString(fmt.Sprintf(`HOST '%s',`, b.sshHost))
		q.WriteString(fmt.Sprintf(`USER '%s',`, b.sshUser))
		q.WriteString(fmt.Sprintf(`PORT %d`, b.sshPort))
	}

	if b.connectionType == "AWS PRIVATELINK" {
		q.WriteString(fmt.Sprintf(`SERVICE NAME '%s',`, b.privateLinkServiceName))
		q.WriteString(fmt.Sprintf(`AVAILABILITY ZONES (%s)`, strings.Join(b.privateLinkAvailabilityZones, ",")))
	}

	if b.connectionType == "POSTGRES" {
		q.WriteString(fmt.Sprintf(`HOST '%s',`, b.postgresHost))
		q.WriteString(fmt.Sprintf(`PORT %d,`, b.postgresPort))
		if b.postgresUser != "" {
			q.WriteString(fmt.Sprintf(`USER '%s',`, b.postgresUser))
		}
		if b.postgresPassword != "" {
			q.WriteString(fmt.Sprintf(`PASSWORD SECRET %s,`, b.postgresPassword))
		}
		if b.postgresSSLMode != "" {
			q.WriteString(fmt.Sprintf(`SSL MODE '%s',`, b.postgresSSLMode))
		}
		if b.postgresSSHTunnel != "" {
			q.WriteString(fmt.Sprintf(`SSH TUNNEL '%s',`, b.postgresSSHTunnel))
		}
		if b.postgresSSLCa != "" {
			q.WriteString(fmt.Sprintf(`SSL CERTIFICATE AUTHORITY SECRET %s,`, b.postgresSSLCa))
		}
		if b.postgresSSLCert != "" {
			q.WriteString(fmt.Sprintf(`SSL CERTIFICATE SECRET %s,`, b.postgresSSLCert))
		}
		if b.postgresSSLKey != "" {
			q.WriteString(fmt.Sprintf(`SSL KEY SECRET %s,`, b.postgresSSLKey))
		}
		if b.postgresAWSPrivateLink != "" {
			q.WriteString(fmt.Sprintf(`AWS PRIVATELINK %s,`, b.postgresAWSPrivateLink))
		}

		q.WriteString(fmt.Sprintf(`DATABASE '%s'`, b.postgresDatabase))
	}

	q.WriteString(`);`)
	return q.String()
}

func (b *ConnectionBuilder) Read() string {
	return fmt.Sprintf(`
		SELECT
			mz_connections.id,
			mz_connections.name,
			mz_connections.type,
			mz_schemas.name AS schema_name
		FROM mz_connections
		JOIN mz_schemas
			ON mz_connections.schema_id = mz_schemas.id
		WHERE mz_connections.name = '%s'
		AND mz_schemas.name = '%s';
	`, b.connectionName, b.schemaName)
}

func (b *ConnectionBuilder) Rename(newConnectionName string) string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`ALTER CONNECTION %s.%s RENAME TO %s.%s;`, b.schemaName, b.connectionName, b.schemaName, newConnectionName))
	return q.String()
}

func (b *ConnectionBuilder) Drop() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`DROP CONNECTION %s.%s;`, b.schemaName, b.connectionName))
	return q.String()
}

func resourceConnectionRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)

	builder := newConnectionBuilder(connectionName, schemaName)
	q := builder.Read()

	var id, name, connection_type, schema_name string
	conn.QueryRow(q).Scan(&id, &name, &connection_type, &schema_name)

	d.SetId(id)

	return diags
}

func resourceConnectionCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sql.DB)

	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)

	builder := newConnectionBuilder(connectionName, schemaName)

	if v, ok := d.GetOk("connection_type"); ok {
		builder.ConnectionType(v.(string))
	}

	if v, ok := d.GetOk("ssh_host"); ok {
		builder.SSHHost(v.(string))
	}

	if v, ok := d.GetOk("ssh_user"); ok {
		builder.SSHUser(v.(string))
	}

	if v, ok := d.GetOk("ssh_port"); ok {
		builder.SSHPort(v.(int))
	}

	if v, ok := d.GetOk("private_link_service_name"); ok {
		builder.PrivateLinkServiceName(v.(string))
	}

	if v, ok := d.GetOk("private_link_availability_zones"); ok {
		builder.PrivateLinkAvailabilityZones(v.([]string))
	}

	if v, ok := d.GetOk("postgres_host"); ok {
		builder.PostgresHost(v.(string))
	}

	if v, ok := d.GetOk("postgres_port"); ok {
		builder.PostgresPort(v.(int))
	}

	if v, ok := d.GetOk("postgres_user"); ok {
		builder.PostgresUser(v.(string))
	}

	if v, ok := d.GetOk("postgres_password"); ok {
		builder.PostgresPassword(v.(string))
	}

	if v, ok := d.GetOk("postgres_database"); ok {
		builder.PostgresDatabase(v.(string))
	}

	if v, ok := d.GetOk("postgres_ssl_mode"); ok {
		builder.PostgresSSLMode(v.(string))
	}

	if v, ok := d.GetOk("postgres_ssl_ca"); ok {
		builder.PostgresSSLCa(v.(string))
	}

	if v, ok := d.GetOk("postgres_ssl_cert"); ok {
		builder.PostgresSSLCert(v.(string))
	}

	if v, ok := d.GetOk("postgres_ssl_key"); ok {
		builder.PostgresSSLKey(v.(string))
	}

	if v, ok := d.GetOk("postgres_aws_privatelink"); ok {
		builder.PostgresAWSPrivateLink(v.(string))
	}

	if v, ok := d.GetOk("postgres_ssh_tunnel"); ok {
		builder.PostgresSSHTunnel(v.(string))
	}

	q := builder.Create()

	ExecResource(conn, q)

	return resourceConnectionRead(ctx, d, meta)
}

func resourceConnectionUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sql.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)

	if d.HasChange("name") {
		newConnectionName := d.Get("name").(string)
		q := newConnectionBuilder(connectionName, schemaName).Rename(newConnectionName)
		ExecResource(conn, q)
	}

	return resourceConnectionRead(ctx, d, meta)
}

func resourceConnectionDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)

	builder := newConnectionBuilder(connectionName, schemaName)
	q := builder.Drop()

	ExecResource(conn, q)
	return diags
}
