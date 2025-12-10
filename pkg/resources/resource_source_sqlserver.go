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

var sourceSQLServerSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("source", true, false),
	"schema_name":        SchemaNameSchema("source", false),
	"database_name":      DatabaseNameSchema("source", false),
	"qualified_sql_name": QualifiedNameSchema("source"),
	"comment":            CommentSchema(false),
	"cluster_name":       ObjectClusterNameSchema("source"),
	"size":               ObjectSizeSchema("source"),
	"sqlserver_connection": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "sqlserver_connection",
		Description: "The SQL Server connection to use in the source.",
		Required:    true,
		ForceNew:    true,
	}),
	"exclude_columns": {
		Description: "Exclude specific columns when reading data from SQL Server. Can only be updated in place when also updating a corresponding `table` attribute.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
	},
	"text_columns": {
		Description: "Decode data as text for specific columns that contain SQL Server types that are unsupported in Materialize. Can only be updated in place when also updating a corresponding `table` attribute.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
	},
	"table": {
		Description: "(Deprecated) Specify the tables to be included in the source. If not specified, all tables are included. Use `materialize_source_table_sqlserver` resources instead.",
		Deprecated:  "The `table` attribute is deprecated and will be removed in a future release. Use `materialize_source_table_sqlserver` resources to create tables from SQL Server sources instead.",
		Type:        schema.TypeSet,
		Optional:    true,
		Computed:    true,
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
	"expose_progress": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "expose_progress",
		Description: "The name of the progress collection for the source. If this is not specified, the collection will be named `<src_name>_progress`.",
		Required:    false,
		ForceNew:    true,
	}),
	"aws_privatelink": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "aws_privatelink",
		Description: "The AWS PrivateLink configuration for the SQL Server database.",
		Required:    false,
		ForceNew:    false,
	}),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func SourceSQLServer() *schema.Resource {
	return &schema.Resource{
		Description: "A SQL Server source describes a SQL Server database instance you want Materialize to read data from using Change Data Capture (CDC).",

		CreateContext: sourceSQLServerCreate,
		ReadContext:   sourceSQLServerRead,
		UpdateContext: sourceSQLServerUpdate,
		DeleteContext: sourceSQLServerDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourceSQLServerSchema,
	}
}

func sourceSQLServerCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}

	o := materialize.MaterializeObject{ObjectType: "SOURCE", Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSourceSQLServerBuilder(metaDb, o)

	if v, ok := d.GetOk("cluster_name"); ok {
		b.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("sqlserver_connection"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.SQLServerConnection(conn)
	}

	if v, ok := d.GetOk("table"); ok {
		tables := v.(*schema.Set).List()
		t := materialize.GetTableStruct(tables)
		b.Table(t)
	}

	if v, ok := d.GetOk("exclude_columns"); ok && len(v.([]interface{})) > 0 {
		columns, err := materialize.GetSliceValueString("exclude_columns", v.([]interface{}))
		if err != nil {
			return diag.FromErr(err)
		}
		b.ExcludeColumns(columns)
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

	if v, ok := d.GetOk("aws_privatelink"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.AWSPrivateLink(conn)
	}

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

	i, err := materialize.SourceId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return sourceSQLServerRead(ctx, d, meta)
}

func sourceSQLServerRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
		sqlserverConnection := []interface{}{
			map[string]interface{}{
				"name":          s.ConnectionName.String,
				"schema_name":   s.ConnectionSchemaName.String,
				"database_name": s.ConnectionDatabaseName.String,
			},
		}
		if err := d.Set("sqlserver_connection", sqlserverConnection); err != nil {
			return diag.FromErr(err)
		}
	}

	deps, err := materialize.ListSQLServerSubsources(metaDb, utils.ExtractId(i), "subsource")
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

func sourceSQLServerUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "SOURCE", Name: sourceName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewSource(metaDb, o)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: "SOURCE", Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
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

	return sourceSQLServerRead(ctx, d, meta)
}

func sourceSQLServerDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
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
