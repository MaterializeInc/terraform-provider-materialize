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

var sourceSqlServerSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("source", true, false),
	"schema_name":        SchemaNameSchema("source", false),
	"database_name":      DatabaseNameSchema("source", false),
	"qualified_sql_name": QualifiedNameSchema("source"),
	"comment":            CommentSchema(false),
	"cluster_name":       ObjectClusterNameSchema("source"),
	"size":               ObjectSizeSchema("source"),
	"sql_server_connection": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "sql_server_connection",
		Description: "The SQL Server connection to use in the source.",
		Required:    true,
		ForceNew:    true,
	}),
	"exclude_columns": {
		Description: "Exclude specific columns when reading data from SQL Server. Can only be updated in place when also updating a corresponding `table` attribute. Deprecated: Use the new `materialize_source_table_sql_server` resource instead.",
		Deprecated:  "Use the new `materialize_source_table_sql_server` resource instead.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
	},
	"text_columns": {
		Description: "Decode data as text for specific columns that contain SQL Server types that are unsupported in Materialize. Can only be updated in place when also updating a corresponding `table` attribute. Deprecated: Use the new `materialize_source_table_sql_server` resource instead.",
		Deprecated:  "Use the new `materialize_source_table_sql_server` resource instead.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
	},
	"table": {
		Description: "Specify the tables to be included in the source. Deprecated: Use the new `materialize_source_table_sql_server` resource instead.",
		Deprecated:  "Use the new `materialize_source_table_sql_server` resource instead.",
		Type:        schema.TypeSet,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"upstream_name": {
					Description: "The name of the table in the upstream SQL Server database.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"upstream_schema_name": {
					Description: "The schema of the table in the upstream SQL Server database.",
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
				},
				"name": {
					Description: "The name for the table, used in Materialize.",
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
				},
				"schema_name": {
					Description: "The schema of the table in Materialize.",
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
				},
				"database_name": {
					Description: "The database of the table in Materialize.",
					Type:        schema.TypeString,
					Optional:    true,
					Computed:    true,
				},
			},
		},
	},
	"all_tables": {
		Description: "Include all tables in the source. If `table` is specified, this will be ignored.",
		Deprecated:  "Use the new `materialize_source_table_sql_server` resource instead.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
	},
	"expose_progress": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "expose_progress",
		Description: "The name of the progress collection for the source. If this is not specified, the collection will be named `<src_name>_progress`.",
		Required:    false,
		ForceNew:    true,
	}),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func SourceSqlServer() *schema.Resource {
	return &schema.Resource{
		CreateContext: sourceSqlServerCreate,
		ReadContext:   sourceSqlServerRead,
		UpdateContext: sourceSqlServerUpdate,
		DeleteContext: sourceSqlServerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourceSqlServerSchema,
	}
}

func sourceSqlServerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "SOURCE", Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSourceSqlServerBuilder(metaDb, o)

	if v, ok := d.GetOk("cluster_name"); ok {
		b.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("sql_server_connection"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.SqlServerConnection(conn)
	}

	// Handle tables
	tables := d.Get("table").(*schema.Set).List()
	allTables := d.Get("all_tables").(bool)
	
	if len(tables) > 0 {
		t := materialize.GetTableStruct(tables)
		b.Tables(t)
	} else if allTables {
		b.AllTables()
	}

	// Handle columns
	if v, ok := d.GetOk("text_columns"); ok && len(v.([]interface{})) > 0 {
		columns, err := materialize.GetSliceValueString("text_columns", v.([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}
		b.TextColumns(columns)
	}

	if v, ok := d.GetOk("exclude_columns"); ok && len(v.([]interface{})) > 0 {
		columns, err := materialize.GetSliceValueString("exclude_columns", v.([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}
		b.IgnoreColumns(columns)
	}

	// Expose Progress
	if v, ok := d.GetOk("expose_progress"); ok {
		e := materialize.GetIdentifierSchemaStruct(v)
		b.ExposeProgress(e)
	}

	// Create source
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// Handle ownership
	if v, ok := d.GetOk("ownership_role"); ok {
		ownership := materialize.NewOwnershipBuilder(metaDb, o)
		if err := ownership.Alter(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed ownership, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// Handle comments
	if v, ok := d.GetOk("comment"); ok {
		comment := materialize.NewCommentBuilder(metaDb, o)
		if err := comment.Object(v.(string)); err != nil {
			log.Printf("[DEBUG] resource failed comment, dropping object: %s", o.Name)
			b.Drop()
			return diag.FromErr(err)
		}
	}

	// Query source ID
	i, err := materialize.SourceId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return sourceSqlServerRead(ctx, d, meta)
}

func sourceSqlServerUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "SOURCE", Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: "SOURCE", Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewSourceSqlServerBuilder(metaDb, o)
		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("ownership_role") {
		newRole := d.Get("ownership_role").(string)
		ownership := materialize.NewOwnershipBuilder(metaDb, o)
		if err := ownership.Alter(newRole); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		comment := materialize.NewCommentBuilder(metaDb, o)
		if err := comment.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return sourceSqlServerRead(ctx, d, meta)
}

func sourceSqlServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()
	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	s, err := materialize.ScanSource(metaDb, utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

	if err := d.Set("name", s.SourceName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	qn := materialize.QualifiedName(s.DatabaseName.String, s.SchemaName.String, s.SourceName.String)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}

	if s.ClusterName.Valid {
		if err := d.Set("cluster_name", s.ClusterName.String); err != nil {
			return diag.FromErr(err)
		}
	}

	if s.Size.Valid {
		if err := d.Set("size", s.Size.String); err != nil {
			return diag.FromErr(err)
		}
	}

	if s.Comment.Valid {
		if err := d.Set("comment", s.Comment.String); err != nil {
			return diag.FromErr(err)
		}
	}

	if s.OwnerName.Valid {
		if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
			return diag.FromErr(err)
		}
	}

	return nil
}

func sourceSqlServerDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{ObjectType: "SOURCE", Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSourceSqlServerBuilder(metaDb, o)

	if err := b.Drop(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}