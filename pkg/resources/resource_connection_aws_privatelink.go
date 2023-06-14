package resources

import (
	"context"
	"database/sql"

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
		Description: "An AWS PrivateLink connection establishes a link to an AWS PrivateLink service.",

		CreateContext: connectionAwsPrivatelinkCreate,
		ReadContext:   connectionAwsPrivatelinkRead,
		UpdateContext: connectionAwsPrivatelinkUpdate,
		DeleteContext: connectionDelete,

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
	i := d.Id()

	s, err := materialize.ScanConnectionAwsPrivatelink(meta.(*sqlx.DB), i)
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
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
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	b := materialize.NewConnectionAwsPrivatelinkBuilder(meta.(*sqlx.DB), connectionName, schemaName, databaseName)

	if v, ok := d.GetOk("service_name"); ok {
		b.PrivateLinkServiceName(v.(string))
	}

	if v, ok := d.GetOk("availability_zones"); ok {
		azs := materialize.GetSliceValueString(v.([]interface{}))
		b.PrivateLinkAvailabilityZones(azs)
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// set id
	i, err := materialize.ConnectionId(meta.(*sqlx.DB), connectionName, schemaName, databaseName)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return connectionAwsPrivatelinkRead(ctx, d, meta)
}

func connectionAwsPrivatelinkUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")

		b := materialize.NewConnectionAwsPrivatelinkBuilder(meta.(*sqlx.DB), oldName.(string), schemaName, databaseName)

		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return connectionAwsPrivatelinkRead(ctx, d, meta)
}
