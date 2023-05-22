package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var connectionAwsPrivatelinkSchema = map[string]*schema.Schema{
	"name":               NameSchema("connection", true, false),
	"schema_name":        SchemaNameSchema("connection", false),
	"database_name":      DatabaseNameSchema("connection", false),
	"qualified_sql_name": QualifiedNameSchema("connection"),
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
	"principal": {
		Description: "The principal of the AWS PrivateLink service.",
		Type:        schema.TypeString,
		Computed:    true,
		Sensitive:   true,
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

type ConnectionAwsPrivatelinkParams struct {
	ConnectionName sql.NullString `db:"connection_name"`
	SchemaName     sql.NullString `db:"schema_name"`
	DatabaseName   sql.NullString `db:"database_name"`
	Principal      sql.NullString `db:"principal"`
}

func connectionAwsPrivatelinkRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := materialize.ReadConnectionAwsPrivatelinkParams(i)

	var s ConnectionAwsPrivatelinkParams
	if err := conn.Get(&s, q); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", s.ConnectionName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("principal", s.Principal.String); err != nil {
		return diag.FromErr(err)
	}

	b := materialize.Connection{ConnectionName: s.ConnectionName.String, SchemaName: s.SchemaName.String, DatabaseName: s.DatabaseName.String}
	if err := d.Set("qualified_sql_name", b.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func connectionAwsPrivatelinkCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewConnectionAwsPrivatelinkBuilder(connectionName, schemaName, databaseName)

	if v, ok := d.GetOk("service_name"); ok {
		builder.PrivateLinkServiceName(v.(string))
	}

	if v, ok := d.GetOk("availability_zones"); ok {
		azs := materialize.GetSliceValueString(v.([]interface{}))
		builder.PrivateLinkAvailabilityZones(azs)
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "connection"); err != nil {
		return diag.FromErr(err)
	}

	return connectionAwsPrivatelinkRead(ctx, d, meta)
}

func connectionAwsPrivatelinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		_, newConnectionName := d.GetChange("name")
		q := materialize.NewConnectionAwsPrivatelinkBuilder(connectionName, schemaName, databaseName).Rename(newConnectionName.(string))
		if err := execResource(conn, q); err != nil {
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

	builder := materialize.NewConnectionAwsPrivatelinkBuilder(connectionName, schemaName, databaseName)
	q := builder.Drop()

	if err := dropResource(conn, d, q, "connection"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
