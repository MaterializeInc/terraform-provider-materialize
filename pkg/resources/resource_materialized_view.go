package resources

import (
	"context"
	"log"
	"terraform-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var materializedViewSchema = map[string]*schema.Schema{
	"name":           SchemaResourceName("materialized view", true, false),
	"schema_name":    SchemaResourceSchemaName("materialized view", false),
	"database_name":  SchemaResourceDatabaseName("materialized view", false),
	"qualified_name": SchemaResourceQualifiedName("materialized view"),
	"in_cluster": {
		Description: "The cluster to maintain the materialized view. If not specified, defaults to the default cluster.",
		Type:        schema.TypeString,
		Optional:    true,
	},
	"select_stmt": {
		Description: "The SQL statement to create the materialized view.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
}

func MaterializedView() *schema.Resource {
	return &schema.Resource{
		Description: "A materialized view persists in durable storage and can be written to, updated and seamlessly joined with other views, views or sources.",

		CreateContext: materializedViewCreate,
		ReadContext:   materializedViewRead,
		UpdateContext: materializedViewUpdate,
		DeleteContext: materializedViewDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: materializedViewSchema,
	}
}

func materializedViewRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := materialize.ReadMaterializedViewParams(i)

	var name, schema, database *string
	if err := conn.QueryRowx(q).Scan(&name, &schema, &database); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", schema); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", database); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func materializedViewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	materializedViewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := materialize.NewMaterializedViewBuilder(materializedViewName, schemaName, databaseName)

	if v, ok := d.GetOk("in_cluster"); ok && v.(string) != "" {
		builder.InCluster(v.(string))
	}

	if v, ok := d.GetOk("select_stmt"); ok && v.(string) != "" {
		builder.SelectStmt(v.(string))
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "materialized view"); err != nil {
		return diag.FromErr(err)
	}
	return materializedViewRead(ctx, d, meta)
}

func materializedViewUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	materializedViewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		_, newName := d.GetChange("name")

		q := materialize.NewMaterializedViewBuilder(materializedViewName, schemaName, databaseName).Rename(newName.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not rename materialized view: %s", q)
			return diag.FromErr(err)
		}
	}

	return materializedViewRead(ctx, d, meta)
}

func materializedViewDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	materializedViewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := materialize.NewMaterializedViewBuilder(materializedViewName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "materialized view"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
