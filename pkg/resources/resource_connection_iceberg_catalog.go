package resources

import (
	"context"
	"database/sql"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var connectionIcebergCatalogSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("connection", true, false),
	"schema_name":        SchemaNameSchema("connection", false),
	"database_name":      DatabaseNameSchema("connection", false),
	"qualified_sql_name": QualifiedNameSchema("connection"),
	"comment":            CommentSchema(false),
	"catalog_type": {
		Description: "The type of Iceberg catalog. Currently only `s3tablesrest` (AWS S3 Tables) is supported.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    false,
	},
	"url": {
		Description: "The URL of the Iceberg catalog endpoint. For AWS S3 Tables, use `https://s3tables.<region>.amazonaws.com/iceberg`.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    false,
	},
	"warehouse": {
		Description: "The ARN of the S3 Tables bucket: `arn:aws:s3tables:<region>:<account-id>:bucket/<bucket-name>`.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    false,
	},
	"aws_connection": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "aws_connection",
		Description: "The name of an AWS connection to use for authentication.",
		Required:    true,
		ForceNew:    false,
	}),
	"validate":       ValidateConnectionSchema(),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func ConnectionIcebergCatalog() *schema.Resource {
	return &schema.Resource{
		Description: "An Iceberg catalog connection provides Materialize with access to an Iceberg catalog, such as AWS S3 Tables.",

		CreateContext: connectionIcebergCatalogCreate,
		ReadContext:   connectionIcebergCatalogRead,
		UpdateContext: connectionIcebergCatalogUpdate,
		DeleteContext: connectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionIcebergCatalogSchema,
	}
}

func connectionIcebergCatalogRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	i := d.Id()

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	s, err := materialize.ScanConnectionIcebergCatalog(metaDb, utils.ExtractId(i))
	if err == sql.ErrNoRows {
		d.SetId("")
		return nil
	} else if err != nil {
		return diag.FromErr(err)
	}

	d.SetId(utils.TransformIdWithRegion(string(region), i))

	if err := d.Set("name", s.ConnectionName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("catalog_type", s.CatalogType.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("url", s.Url.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("warehouse", s.Warehouse.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("ownership_role", s.OwnerName.String); err != nil {
		return diag.FromErr(err)
	}

	b := materialize.Connection{ConnectionName: s.ConnectionName.String, SchemaName: s.SchemaName.String, DatabaseName: s.DatabaseName.String}
	if err := d.Set("qualified_sql_name", b.QualifiedName()); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("comment", s.Comment.String); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func connectionIcebergCatalogCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "CONNECTION", Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewConnectionIcebergCatalogBuilder(metaDb, o)

	if v, ok := d.GetOk("catalog_type"); ok {
		b.CatalogType(v.(string))
	}

	if v, ok := d.GetOk("url"); ok {
		b.Url(v.(string))
	}

	if v, ok := d.GetOk("warehouse"); ok {
		b.Warehouse(v.(string))
	}

	if v, ok := d.GetOk("aws_connection"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.AwsConnection(conn)
	}

	if v, ok := d.GetOk("validate"); ok {
		b.Validate(v.(bool))
	}

	// create resource
	if err := b.Create(); err != nil {
		return diag.FromErr(err)
	}

	// ownership
	if diags := applyOwnership(d, metaDb, o, b); diags != nil {
		return diags
	}

	// object comment
	if diags := applyComment(d, metaDb, o, b); diags != nil {
		return diags
	}

	// set id
	i, err := materialize.ConnectionId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return connectionIcebergCatalogRead(ctx, d, meta)
}

func connectionIcebergCatalogUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)
	validate := d.Get("validate").(bool)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "CONNECTION", Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: "CONNECTION", Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewConnectionIcebergCatalogBuilder(metaDb, o)
		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	if d.HasChange("catalog_type") {
		oldCatalogType, newCatalogType := d.GetChange("catalog_type")
		b := materialize.NewConnection(metaDb, o)
		options := map[string]interface{}{
			"CATALOG TYPE": newCatalogType.(string),
		}
		if err := b.Alter(options, nil, false, validate); err != nil {
			d.Set("catalog_type", oldCatalogType)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("url") {
		oldUrl, newUrl := d.GetChange("url")
		b := materialize.NewConnection(metaDb, o)
		options := map[string]interface{}{
			"URL": newUrl.(string),
		}
		if err := b.Alter(options, nil, false, validate); err != nil {
			d.Set("url", oldUrl)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("warehouse") {
		oldWarehouse, newWarehouse := d.GetChange("warehouse")
		b := materialize.NewConnection(metaDb, o)
		options := map[string]interface{}{
			"WAREHOUSE": newWarehouse.(string),
		}
		if err := b.Alter(options, nil, false, validate); err != nil {
			d.Set("warehouse", oldWarehouse)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("aws_connection") {
		oldAwsConnection, newAwsConnection := d.GetChange("aws_connection")
		b := materialize.NewConnection(metaDb, o)
		options := map[string]interface{}{
			"AWS CONNECTION": materialize.GetIdentifierSchemaStruct(newAwsConnection),
		}
		if err := b.Alter(options, nil, false, validate); err != nil {
			d.Set("aws_connection", oldAwsConnection)
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

	return connectionIcebergCatalogRead(ctx, d, meta)
}
