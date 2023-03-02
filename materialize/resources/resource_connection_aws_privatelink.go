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

var connectionAwsPrivatelinkSchema = map[string]*schema.Schema{
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
	"service_name": {
		Description: "The name of the AWS PrivateLink service.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"availability_zones": {
		Description: "The availability zones of the AWS PrivateLink service.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Required:    true,
		ForceNew:    true,
	},
}

func ConnectionAwsPrivatelink() *schema.Resource {
	return &schema.Resource{
		Description: "The connection resource allows you to manage connections in Materialize.",

		CreateContext: connectionAwsPrivatelinkCreate,
		ReadContext:   connectionAwsPrivatelinkRead,
		UpdateContext: connectionAwsPrivatelinkUpdate,
		DeleteContext: connectionAwsPrivatelinkDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionAwsPrivatelinkSchema,
	}
}

type ConnectionAwsPrivatelinkBuilder struct {
	connectionName               string
	schemaName                   string
	databaseName                 string
	connectionType               string
	privateLinkServiceName       string
	privateLinkAvailabilityZones []string
}

func newConnectionAwsPrivatelinkBuilder(connectionName, schemaName, databaseName string) *ConnectionAwsPrivatelinkBuilder {
	return &ConnectionAwsPrivatelinkBuilder{
		connectionName: connectionName,
		schemaName:     schemaName,
		databaseName:   databaseName,
	}
}

func (b *ConnectionAwsPrivatelinkBuilder) PrivateLinkServiceName(privateLinkServiceName string) *ConnectionAwsPrivatelinkBuilder {
	b.privateLinkServiceName = privateLinkServiceName
	return b
}

func (b *ConnectionAwsPrivatelinkBuilder) PrivateLinkAvailabilityZones(privateLinkAvailabilityZones []string) *ConnectionAwsPrivatelinkBuilder {
	b.privateLinkAvailabilityZones = privateLinkAvailabilityZones
	return b
}

func (b *ConnectionAwsPrivatelinkBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CONNECTION %s.%s.%s TO AWS PRIVATELINK (`, b.databaseName, b.schemaName, b.connectionName))

	q.WriteString(fmt.Sprintf(`SERVICE NAME '%s',`, b.privateLinkServiceName))
	q.WriteString(fmt.Sprintf(`AVAILABILITY ZONES ('%s')`, strings.Join(b.privateLinkAvailabilityZones, "', '")))

	q.WriteString(`);`)
	return q.String()
}

func (b *ConnectionAwsPrivatelinkBuilder) ReadId() string {
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

func (b *ConnectionAwsPrivatelinkBuilder) Rename(newConnectionName string) string {
	return fmt.Sprintf(`ALTER CONNECTION %s.%s.%s RENAME TO %s.%s.%s;`, b.databaseName, b.schemaName, b.connectionName, b.databaseName, b.schemaName, newConnectionName)
}

func (b *ConnectionAwsPrivatelinkBuilder) Drop() string {
	return fmt.Sprintf(`DROP CONNECTION %s.%s.%s;`, b.databaseName, b.schemaName, b.connectionName)
}

func connectionAwsPrivatelinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newConnectionAwsPrivatelinkBuilder(connectionName, schemaName, databaseName)

	if v, ok := d.GetOk("service_name"); ok {
		builder.PrivateLinkServiceName(v.(string))
	}

	if v, ok := d.GetOk("availability_zones"); ok {
		azs := v.([]interface{})
		var azStrings []string
		for _, az := range azs {
			azStrings = append(azStrings, az.(string))
		}
		builder.PrivateLinkAvailabilityZones(azStrings)
	}

	qc := builder.Create()
	qr := builder.ReadId()

	createResource(conn, d, qc, qr, "connection")
	return connectionAwsPrivatelinkRead(ctx, d, meta)
}

func connectionAwsPrivatelinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readConnectionParams(i)

	readResource(conn, d, i, q, _connection{}, "connection")
	setQualifiedName(d)
	return nil
}

func connectionAwsPrivatelinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		newConnectionName := d.Get("name").(string)
		q := newConnectionAwsPrivatelinkBuilder(connectionName, schemaName, databaseName).Rename(newConnectionName)
		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return connectionAwsPrivatelinkRead(ctx, d, meta)
}

func connectionAwsPrivatelinkDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newConnectionAwsPrivatelinkBuilder(connectionName, schemaName, databaseName)
	q := builder.Drop()

	dropResource(conn, d, q, "connection")
	return nil
}
