package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var GrantDefinition = "Manages the privileges on a Materailize %[1]s for roles."
var DefaultPrivilegeDefinition = "Defines default privileges that will be applied to objects created in the future. It does not affect any existing objects."

var materializedViewSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("materialized view", true, false),
	"schema_name":        SchemaNameSchema("materialized view", false),
	"database_name":      DatabaseNameSchema("materialized view", false),
	"qualified_sql_name": QualifiedNameSchema("materialized view"),
	"comment":            CommentSchema(false),
	"cluster_name": {
		Description: "The cluster to maintain the materialized view.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"not_null_assertion": {
		Description: "**Private Preview** A list of columns for which to create non-null assertions.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
		ForceNew:    true,
	},
	"statement": {
		Description: "The SQL statement for the materialized view.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"create_sql": {
		Description: "The SQL statement used to create the materialized view.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func MaterializedView() *schema.Resource {
	return &schema.Resource{
		Description: "Materialized views represent query results stored durably.",

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

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	s, err := materialize.ScanMaterializedView(metaDb, utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

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

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("create_sql", s.CreateSQL.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func materializedViewCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	materializedViewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "MATERIALIZED VIEW", Name: materializedViewName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewMaterializedViewBuilder(metaDb, o)

	if v, ok := d.GetOk("cluster_name"); ok && v.(string) != "" {
		b.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("not_null_assertion"); ok && len(v.([]interface{})) > 0 {
		nas := materialize.GetSliceValueString(v.([]interface{}))
		if len(nas) > 0 {
			b.NotNullAssertions(nas)
		}
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
		ownership := materialize.NewOwnershipBuilder(metaDb, o)

		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// object comment
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(metaDb, o)

		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// set id
	i, err := materialize.MaterializedViewId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return materializedViewRead(ctx, d, meta)
}

func materializedViewUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	materializedViewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "MATERIALIZED VIEW", Name: materializedViewName, SchemaName: schemaName, DatabaseName: databaseName}

	if d.HasChange("name") {
		oldName, newMaterializedViewName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: "MATERIALIZED VIEW", Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewMaterializedViewBuilder(metaDb, o)
		if err := b.Rename(newMaterializedViewName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("ownership_role") {
		_, newRole := d.GetChange("ownership_role")
		b := materialize.NewOwnershipBuilder(metaDb, o)

		if err := b.Alter(newRole.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(metaDb, o)

		if err := b.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return materializedViewRead(ctx, d, meta)
}

func materializedViewDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	materializedViewName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{Name: materializedViewName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewMaterializedViewBuilder(metaDb, o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
