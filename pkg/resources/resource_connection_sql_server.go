package resources

import (
	"context"
	"log"
	"strings"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
)

var connectionSqlServerSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("connection", true, false),
	"schema_name":        SchemaNameSchema("connection", false),
	"database_name":      DatabaseNameSchema("connection", false),
	"qualified_sql_name": QualifiedNameSchema("connection"),
	"comment":            CommentSchema(false),
	"host": {
		Description: "The SQL Server database hostname.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    false,
	},
	"port": {
		Description: "The SQL Server database port.",
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     1433,
		ForceNew:    false,
	},
	"user": ValueSecretSchema("user", "The SQL Server database username.", true, false),
	"password": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "password",
		Description: "The SQL Server database password.",
		Required:    false,
		ForceNew:    false,
	}),
	"database": {
		Description: "The SQL Server database to connect to.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    false,
	},
	"ssl_mode": {
		Description:  "The SSL mode for the SQL Server database. Allowed values are " + strings.Join(sqlServerSSLMode, ", ") + ".",
		Type:         schema.TypeString,
		Optional:     true,
		Default:      "require",
		ForceNew:     false,
		ValidateFunc: validation.StringInSlice(sqlServerSSLMode, true),
	},
	"ssh_tunnel": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "ssh_tunnel",
		Description: "The SSH tunnel configuration for the SQL Server database.",
		Required:    false,
		ForceNew:    false,
	}),
	"ssl_certificate_authority": ValueSecretSchema("ssl_certificate_authority", "The CA certificate for the SQL Server database.", false, false),
	"aws_privatelink": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "aws_privatelink",
		Description: "The AWS PrivateLink configuration for the SQL Server database.",
		Required:    false,
		ForceNew:    false,
	}),
	"validate":       ValidateConnectionSchema(),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func ConnectionSqlServer() *schema.Resource {
	return &schema.Resource{
		Description: "A SQL Server connection establishes a link to a SQL Server database.",

		CreateContext: connectionSqlServerCreate,
		ReadContext:   connectionRead,
		UpdateContext: connectionUpdate,
		DeleteContext: connectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionSqlServerSchema,
	}
}

func connectionSqlServerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "CONNECTION", Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewConnectionSqlServerBuilder(metaDb, o)

	if v, ok := d.GetOk("host"); ok {
		b.SqlServerHost(v.(string))
	}

	if v, ok := d.GetOk("port"); ok {
		b.SqlServerPort(v.(int))
	}

	if v, ok := d.GetOk("user"); ok {
		user := materialize.GetValueSecretStruct(v)
		b.SqlServerUser(user)
	}

	if v, ok := d.GetOk("password"); ok {
		pass := materialize.GetIdentifierSchemaStruct(v)
		b.SqlServerPassword(pass)
	}

	if v, ok := d.GetOk("database"); ok {
		b.SqlServerDatabase(v.(string))
	}

	if v, ok := d.GetOk("ssl_mode"); ok {
		b.SqlServerSSLMode(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate_authority"); ok {
		ssl_ca := materialize.GetValueSecretStruct(v)
		b.SqlServerSSLCa(ssl_ca)
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.SqlServerSSHTunnel(conn)
	}

	if v, ok := d.GetOk("aws_privatelink"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.SqlServerAWSPrivateLink(conn)
	}

	if v, ok := d.GetOk("validate"); ok {
		b.Validate(v.(bool))
	}

	// Create connection
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

	// Query the connection
	i, err := materialize.ConnectionId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return connectionRead(ctx, d, meta)
}
