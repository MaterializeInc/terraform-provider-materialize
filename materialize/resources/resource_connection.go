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
				Description: "The type of the connection.",
				Type:        schema.TypeString,
				Required:    true,
				ValidateFunc: validation.StringInSlice([]string{
					"SSH TUNNEL",
				}, false),
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
		},
	}
}

type ConnectionBuilder struct {
	connectionName string
	schemaName     string
	connectionType string
	sshHost        string
	sshUser        string
	sshPort        int
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

func (b *ConnectionBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s.%s`, b.connectionName, b.schemaName))

	if b.connectionType != "" {
		q.WriteString(fmt.Sprintf(` TO %s (`, b.connectionType))
	}

	if b.connectionType == "SSH TUNNEL" {
		q.WriteString(fmt.Sprintf(`HOST '%s'`, b.sshHost))
		q.WriteString(fmt.Sprintf(`USER '%s'`, b.sshUser))
		q.WriteString(fmt.Sprintf(`PORT %d`, b.sshPort))
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
	connectionType := d.Get("connection_type").(string)

	builder := newConnectionBuilder(connectionName, schemaName).ConnectionType(connectionType)

	if connectionType == "SSH TUNNEL" {
		sshHost := d.Get("ssh_host").(string)
		sshUser := d.Get("ssh_user").(string)
		sshPort := d.Get("ssh_port").(int)

		builder.SSHHost(sshHost).SSHUser(sshUser).SSHPort(sshPort)
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
