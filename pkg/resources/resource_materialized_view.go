package resources

import (
	"context"
	"database/sql"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var materializedViewSchema = map[string]*schema.Schema{
	"name":               NameSchema("materialized view", true, false),
	"schema_name":        SchemaNameSchema("materialized view", false),
	"database_name":      DatabaseNameSchema("materialized view", false),
	"qualified_sql_name": QualifiedNameSchema("materialized view"),
	"cluster_name": {
		Description: "The cluster to maintain the materialized view. If not specified, defaults to the default cluster.",
		Type:        schema.TypeString,
		Optional:    true,
		Computed:    true,
	},
	"statement": {
		Description: "The SQL statement to create the materialized view.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"ownership_role": OwnershipRole(),
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
	i := d.Id()

	s, err := materialize.ScanMaterializedView(meta.(*sqlx.DB), i)
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", s.MaterializedViewName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_name", s.Cluster.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}

	qn := materialize.QualifiedName(s.DatabaseName.String, s.SchemaName.String, s.MaterializedViewName.String)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func materializedViewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	materializedViewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	o := materialize.ObjectSchemaStruct{Name: materializedViewName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewMaterializedViewBuilder(meta.(*sqlx.DB), o)

	if v, ok := d.GetOk("cluster_name"); ok && v.(string) != "" {
		b.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("statement"); ok && v.(string) != "" {
		b.SelectStmt(v.(string))
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), "MATERIALIZED VIEW", o)

		if err := ownership.Alter(v.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	// set id
	i, err := materialize.MaterializedViewId(meta.(*sqlx.DB), o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(i)

	return materializedViewRead(ctx, d, meta)
}

func materializedViewUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	materializedViewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		oldMaterializedViewName, newMaterializedViewName := d.GetChange("name")

		o := materialize.ObjectSchemaStruct{Name: oldMaterializedViewName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewMaterializedViewBuilder(meta.(*sqlx.DB), o)

		if err := b.Rename(newMaterializedViewName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")

		o := materialize.ObjectSchemaStruct{Name: materializedViewName, SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewOwnershipBuilder(meta.(*sqlx.DB), "MATERIALIZED VIEW", o)

		if err := b.Alter(newRole.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return materializedViewRead(ctx, d, meta)
}

func materializedViewDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	materializedViewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	o := materialize.ObjectSchemaStruct{Name: materializedViewName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewMaterializedViewBuilder(meta.(*sqlx.DB), o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
