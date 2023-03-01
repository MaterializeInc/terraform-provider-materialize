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

var connectionSshTunnelSchema = map[string]*schema.Schema{
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
	"host": {
		Description:  "The host of the SSH tunnel.",
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"user", "port"},
	},
	"user": {
		Description:  "The user of the SSH tunnel.",
		Type:         schema.TypeString,
		Optional:     true,
		RequiredWith: []string{"host", "port"},
	},
	"port": {
		Description:  "The port of the SSH tunnel.",
		Type:         schema.TypeInt,
		Optional:     true,
		RequiredWith: []string{"host", "user"},
	},
}

func ConnectionSshTunnel() *schema.Resource {
	return &schema.Resource{
		Description: "The connection resource allows you to manage connections in Materialize.",

		CreateContext: connectionSshTunnelCreate,
		ReadContext:   connectionSshTunnelRead,
		UpdateContext: connectionSshTunnelUpdate,
		DeleteContext: connectionSshTunnelDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionSshTunnelSchema,
	}
}

type ConnectionSshTunnelBuilder struct {
	connectionName string
	schemaName     string
	databaseName   string
	connectionType string
	sshHost        string
	sshUser        string
	sshPort        int
}

func newConnectionSshTunnelBuilder(connectionName, schemaName, databaseName string) *ConnectionSshTunnelBuilder {
	return &ConnectionSshTunnelBuilder{
		connectionName: connectionName,
		schemaName:     schemaName,
		databaseName:   databaseName,
	}
}

func (b *ConnectionSshTunnelBuilder) ConnectionName(connectionName string) *ConnectionSshTunnelBuilder {
	b.connectionName = connectionName
	return b
}

func (b *ConnectionSshTunnelBuilder) SchemaName(schemaName string) *ConnectionSshTunnelBuilder {
	b.schemaName = schemaName
	return b
}

func (b *ConnectionSshTunnelBuilder) SSHHost(sshHost string) *ConnectionSshTunnelBuilder {
	b.sshHost = sshHost
	return b
}

func (b *ConnectionSshTunnelBuilder) SSHUser(sshUser string) *ConnectionSshTunnelBuilder {
	b.sshUser = sshUser
	return b
}

func (b *ConnectionSshTunnelBuilder) SSHPort(sshPort int) *ConnectionSshTunnelBuilder {
	b.sshPort = sshPort
	return b
}

func (b *ConnectionSshTunnelBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s.%s.%s TO SSH TUNNEL (`, b.databaseName, b.schemaName, b.connectionName))

	q.WriteString(fmt.Sprintf(`HOST '%s', `, b.sshHost))
	q.WriteString(fmt.Sprintf(`USER '%s', `, b.sshUser))
	q.WriteString(fmt.Sprintf(`PORT %d`, b.sshPort))

	q.WriteString(`);`)
	return q.String()
}

func (b *ConnectionSshTunnelBuilder) ReadId() string {
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

func (b *ConnectionSshTunnelBuilder) Rename(newConnectionName string) string {
	return fmt.Sprintf(`ALTER CONNECTION %s.%s.%s RENAME TO %s.%s.%s;`, b.databaseName, b.schemaName, b.connectionName, b.databaseName, b.schemaName, newConnectionName)
}

func (b *ConnectionSshTunnelBuilder) Drop() string {
	return fmt.Sprintf(`DROP CONNECTION %s.%s.%s;`, b.databaseName, b.schemaName, b.connectionName)
}

func connectionSshTunnelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newConnectionSshTunnelBuilder(connectionName, schemaName, databaseName)

	if v, ok := d.GetOk("host"); ok {
		builder.SSHHost(v.(string))
	}

	if v, ok := d.GetOk("user"); ok {
		builder.SSHUser(v.(string))
	}

	if v, ok := d.GetOk("port"); ok {
		builder.SSHPort(v.(int))
	}

	qc := builder.Create()
	qr := builder.ReadId()

	createResource(conn, d, qc, qr, "connection")
	return connectionSshTunnelRead(ctx, d, meta)
}

func connectionSshTunnelRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readConnectionParams(i)

	readResource(conn, d, i, q, _connection{}, "connection")
	setQualifiedName(d)
	return nil
}

func connectionSshTunnelUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		newConnectionName := d.Get("name").(string)
		q := newConnectionSshTunnelBuilder(connectionName, schemaName, databaseName).Rename(newConnectionName)
		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return connectionSshTunnelRead(ctx, d, meta)
}

func connectionSshTunnelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newConnectionSshTunnelBuilder(connectionName, schemaName, databaseName)
	q := builder.Drop()

	dropResource(conn, d, q, "connection")
	return nil
}
