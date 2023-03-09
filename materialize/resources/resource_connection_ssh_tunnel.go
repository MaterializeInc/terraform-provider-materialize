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
	"connection_type": {
		Description: "The type of connection.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"host": {
		Description: "The host of the SSH tunnel.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"user": {
		Description: "The user of the SSH tunnel.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"port": {
		Description: "The port of the SSH tunnel.",
		Type:        schema.TypeInt,
		Required:    true,
	},
}

func ConnectionSshTunnel() *schema.Resource {
	return &schema.Resource{
		Description: "The connection resource allows you to manage connections in Materialize.",

		CreateContext: connectionSshTunnelCreate,
		ReadContext:   ConnectionRead,
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
	sshHost        string
	sshUser        string
	sshPort        int
}

func (b *ConnectionSshTunnelBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.connectionName)
}

func newConnectionSshTunnelBuilder(connectionName, schemaName, databaseName string) *ConnectionSshTunnelBuilder {
	return &ConnectionSshTunnelBuilder{
		connectionName: connectionName,
		schemaName:     schemaName,
		databaseName:   databaseName,
	}
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
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s TO SSH TUNNEL (`, b.qualifiedName()))

	q.WriteString(fmt.Sprintf(`HOST '%s', `, b.sshHost))
	q.WriteString(fmt.Sprintf(`USER '%s', `, b.sshUser))
	q.WriteString(fmt.Sprintf(`PORT %d`, b.sshPort))

	q.WriteString(`);`)
	return q.String()
}

func (b *ConnectionSshTunnelBuilder) Rename(newConnectionName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newConnectionName)
	return fmt.Sprintf(`ALTER CONNECTION %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *ConnectionSshTunnelBuilder) Drop() string {
	return fmt.Sprintf(`DROP CONNECTION %s;`, b.qualifiedName())
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

func connectionSshTunnelCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newConnectionSshTunnelBuilder(connectionName, schemaName, databaseName)

	builder.SSHHost(d.Get("host").(string))
	builder.SSHUser(d.Get("user").(string))
	builder.SSHPort(d.Get("port").(int))

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "connection"); err != nil {
		return diag.FromErr(err)
	}
	return ConnectionRead(ctx, d, meta)
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

	return ConnectionRead(ctx, d, meta)
}

func connectionSshTunnelDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := newConnectionSshTunnelBuilder(connectionName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "connection"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
