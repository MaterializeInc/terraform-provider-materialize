package resources

import (
	"context"
	"fmt"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var connectionIcebergCatalogSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("connection", true, false),
	"schema_name":        SchemaNameSchema("connection", false),
	"database_name":      DatabaseNameSchema("connection", false),
	"qualified_sql_name": QualifiedNameSchema("connection"),
	"comment":            CommentSchema(false),
	"catalog_type": {
		Description:  "The type of Iceberg catalog: `s3tablesrest` for AWS S3 Tables, or `rest` for any Iceberg REST catalog such as Databricks Unity Catalog.",
		Type:         schema.TypeString,
		Required:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice([]string{"s3tablesrest", "rest"}, false),
	},
	"url": {
		Description: "The URL of the Iceberg catalog endpoint. For AWS S3 Tables, use `https://s3tables.<region>.amazonaws.com/iceberg`. For a REST catalog, the path the catalog serves `/v1/` under.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"warehouse": {
		Description: "The warehouse to operate in. For AWS S3 Tables, the ARN of the S3 Tables bucket: `arn:aws:s3tables:<region>:<account-id>:bucket/<bucket-name>`. For a REST catalog this is catalog specific, for example the Unity Catalog name on Databricks.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"aws_connection": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "aws_connection",
		Description: "The name of an AWS connection to use for authentication. Required for `s3tablesrest` catalogs and not allowed for `rest` catalogs.",
		Required:    false,
		ForceNew:    true,
	}),
	"credential": ValueSecretSchema("credential", "OAuth2 client credentials for a `rest` catalog, as `<client_id>:<client_secret>`. A value without a colon is sent as the client secret alone. Required for `rest` catalogs", false, true),
	"oauth2_server_url": {
		Description: "The token endpoint the `credential` is exchanged at. Defaults to the catalog's own `/v1/oauth/tokens` endpoint.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"scope": {
		Description: "The OAuth2 scope to request, for example `all-apis` on Databricks.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"access_delegation": {
		Description:  "Ask the catalog to vend temporary, table-scoped storage credentials. The only accepted value is `vended-credentials`. Only valid with `rest` catalogs, and required by Databricks Unity Catalog.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ValidateFunc: validation.StringInSlice([]string{"vended-credentials"}, false),
	},
	"validate":       ValidateConnectionSchema(),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func ConnectionIcebergCatalog() *schema.Resource {
	return &schema.Resource{
		Description: "An Iceberg catalog connection establishes a link to an Apache Iceberg catalog. You can use Iceberg catalog connections to create Iceberg sinks.",

		CreateContext: connectionIcebergCatalogCreate,
		ReadContext:   connectionRead,
		UpdateContext: connectionIcebergCatalogUpdate,
		DeleteContext: connectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		CustomizeDiff: connectionIcebergCatalogValidateOptions,

		Schema: connectionIcebergCatalogSchema,
	}
}

// Materialize rejects these combinations as well, but only at apply time.
// Values still unknown at plan time are left for Materialize to check.
func connectionIcebergCatalogValidateOptions(ctx context.Context, d *schema.ResourceDiff, meta interface{}) error {
	for _, k := range []string{"catalog_type", "aws_connection", "credential", "access_delegation"} {
		if !d.NewValueKnown(k) {
			return nil
		}
	}
	catalogType := d.Get("catalog_type").(string)
	hasAws := len(d.Get("aws_connection").([]interface{})) > 0
	hasCredential := len(d.Get("credential").([]interface{})) > 0
	switch catalogType {
	case "s3tablesrest":
		if !hasAws {
			return fmt.Errorf("aws_connection is required when catalog_type is %q", catalogType)
		}
		if d.Get("access_delegation").(string) != "" {
			return fmt.Errorf("access_delegation is not supported when catalog_type is %q", catalogType)
		}
	case "rest":
		if hasAws {
			return fmt.Errorf("aws_connection is not supported when catalog_type is %q, use credential", catalogType)
		}
		if !hasCredential {
			return fmt.Errorf("credential is required when catalog_type is %q", catalogType)
		}
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
	o := materialize.MaterializeObject{ObjectType: materialize.BaseConnection, Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}
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

	if v, ok := d.GetOk("credential"); ok {
		b.Credential(materialize.GetValueSecretStruct(v))
	}

	if v, ok := d.GetOk("oauth2_server_url"); ok {
		b.Oauth2ServerUrl(v.(string))
	}

	if v, ok := d.GetOk("scope"); ok {
		b.Scope(v.(string))
	}

	if v, ok := d.GetOk("access_delegation"); ok {
		b.AccessDelegation(v.(string))
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

	return connectionRead(ctx, d, meta)
}

func connectionIcebergCatalogUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, _, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: materialize.BaseConnection, Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")
		o := materialize.MaterializeObject{ObjectType: materialize.BaseConnection, Name: oldName.(string), SchemaName: schemaName, DatabaseName: databaseName}
		b := materialize.NewConnectionIcebergCatalogBuilder(metaDb, o)
		if err := b.Rename(newName.(string)); err != nil {
			return diag.FromErr(err)
		}
	}

	// TODO: catalog_type, url, warehouse, aws_connection, credential, oauth2_server_url, scope
	// and access_delegation cannot be altered and are marked with ForceNew: true, so changes to
	// them will recreate the resource.
	// Error: "storage error: cannot be altered in the requested way (SQLSTATE XX000)"
	// Once Materialize supports ALTER for these properties, remove ForceNew and add ALTER logic here.

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

	return connectionRead(ctx, d, meta)
}
