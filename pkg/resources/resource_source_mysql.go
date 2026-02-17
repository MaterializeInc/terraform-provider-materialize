package resources

import (
	"context"
	"database/sql"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var sourceMySQLSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("source", true, false),
	"schema_name":        SchemaNameSchema("source", false),
	"database_name":      DatabaseNameSchema("source", false),
	"qualified_sql_name": QualifiedNameSchema("source"),
	"comment":            CommentSchema(false),
	"cluster_name":       ObjectClusterNameSchema("source"),
	"size":               ObjectSizeSchema("source"),
	"mysql_connection": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "mysql_connection",
		Description: "The MySQL connection to use in the source.",
		Required:    true,
		ForceNew:    true,
	}),
	"ignore_columns": {
		Description: "(Deprecated) Ignore specific columns when reading data from MySQL. Use `materialize_source_table_mysql` resources instead.",
		Deprecated:  "The `ignore_columns` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_mysql` resources with `exclude_columns` instead.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
	},
	"text_columns": {
		Description: "(Deprecated) Decode data as text for specific columns that contain MySQL types that are unsupported in Materialize. Use `materialize_source_table_mysql` resources instead.",
		Deprecated:  "The `text_columns` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_mysql` resources instead.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
	},
	"table": {
		Description: "(Deprecated) Specify the tables to be included in the source. If not specified, all tables are included. Use `materialize_source_table_mysql` resources instead.",
		Deprecated:  "The `table` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_mysql` resources to create tables from MySQL sources instead.",
		Type:        schema.TypeSet,
		Optional:    true,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"upstream_name": {
					Description: "The name of the table in the upstream MySQL database.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"upstream_schema_name": {
					Description: "The schema of the table in the upstream MySQL database.",
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
		Description: "**IMPORTANT**: Include all tables in the source. If `table` is specified, this will be ignored. **We strongly recommend using `materialize_source_table_mysql` resources instead** for better control and management of individual tables. See the migration guide for details.",
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

func SourceMySQL() *schema.Resource {
	return &schema.Resource{
		Description: "A MySQL source describes a MySQL instance you want Materialize to read data from.",

		CreateContext: sourceMySQLCreate,
		ReadContext:   sourceMySQLRead,
		UpdateContext: sourceMySQLUpdate,
		DeleteContext: sourceMySQLDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourceMySQLSchema,
	}
}

func sourceMySQLCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{ObjectType: materialize.BaseSource, Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSourceMySQLBuilder(metaDb, o)

	if v, ok := d.GetOk("cluster_name"); ok {
		b.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("mysql_connection"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.MySQLConnection(conn)
	}

	if v, ok := d.GetOk("table"); ok {
		tables := v.(*schema.Set).List()
		t := materialize.GetTableStruct(tables)
		b.Tables(t)
	}

	if v, ok := d.GetOk("all_tables"); ok && v.(bool) {
		b.AllTables()
	}

	if v, ok := d.GetOk("ignore_columns"); ok && len(v.([]interface{})) > 0 {
		columns, err := materialize.GetSliceValueString("ignore_columns", v.([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}
		b.IgnoreColumns(columns)
	}

	if v, ok := d.GetOk("text_columns"); ok && len(v.([]interface{})) > 0 {
		columns, err := materialize.GetSliceValueString("text_columns", v.([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}
		b.TextColumns(columns)
	}

	if v, ok := d.GetOk("expose_progress"); ok {
		e := materialize.GetIdentifierSchemaStruct(v)
		b.ExposeProgress(e)
	}

	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// Handle ownership
	if diags := applyOwnership(d, metaDb, o, b); diags != nil {
		return diags
	}

	// Handle comments
	if diags := applyComment(d, metaDb, o, b); diags != nil {
		return diags
	}

	i, err := materialize.SourceId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return sourceMySQLRead(ctx, d, meta)
}

func sourceMySQLRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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

	if err := d.Set("size", s.Size.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("cluster_name", s.ClusterName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}

	b := materialize.Source{SourceName: s.SourceName.String, SchemaName: s.SchemaName.String, DatabaseName: s.DatabaseName.String}
	if err := d.Set("qualified_sql_name", b.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	if s.ConnectionName.Valid && s.ConnectionSchemaName.Valid && s.ConnectionDatabaseName.Valid {
		mysqlConnection := []interface{}{
			map[string]interface{}{
				"name":          s.ConnectionName.String,
				"schema_name":   s.ConnectionSchemaName.String,
				"database_name": s.ConnectionDatabaseName.String,
			},
		}
		if err := d.Set("mysql_connection", mysqlConnection); err != nil {
			return diag.FromErr(err)
		}
	}

	deps, err := materialize.ListMysqlSubsources(metaDb, utils.ExtractId(i), "subsource")
	if err != nil {
		return diag.FromErr(err)
	}

	// Tables
	tMaps := []interface{}{}
	for _, dep := range deps {
		tMap := map[string]interface{}{}
		tMap["upstream_name"] = dep.UpstreamTableName.String
		tMap["upstream_schema_name"] = dep.UpstreamTableSchemaName.String
		tMap["name"] = dep.ObjectName.String
		tMap["schema_name"] = dep.ObjectSchemaName.String
		tMap["database_name"] = dep.ObjectDatabaseName.String
		tMaps = append(tMaps, tMap)
	}
	if err := d.Set("table", tMaps); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func sourceMySQLUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: materialize.BaseSource, Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSource(metaDb, o)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: materialize.BaseSource, Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewSource(metaDb, o)
		if err := b.Rename(newName.(string)); err != nil {
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

	if d.HasChange("table") {
		ot, nt := d.GetChange("table")
		addTables := materialize.DiffTableStructs(nt.(*schema.Set).List(), ot.(*schema.Set).List())
		dropTables := materialize.DiffTableStructs(ot.(*schema.Set).List(), nt.(*schema.Set).List())
		if len(dropTables) > 0 {
			if err := b.DropSubsource(dropTables); err != nil {
				return diag.FromErr(err)
			}
		}
		if len(addTables) > 0 {
			var colDiff []string
			if d.HasChange("text_columns") {
				oc, nc := d.GetChange("text_columns")
				colDiff = diffTextColumns(nc.([]interface{}), oc.([]interface{}))
			}

			if err := b.AddSubsource(addTables, colDiff); err != nil {
				return diag.FromErr(err)
			}
		}
	}

	if d.HasChange("comment") {
		_, newComment := d.GetChange("comment")
		b := materialize.NewCommentBuilder(metaDb, o)

		if err := b.Object(newComment.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	return sourceMySQLRead(ctx, d, meta)
}

func sourceMySQLDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSource(metaDb, o)

	if err := b.DropCascade(); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
