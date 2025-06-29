package resources

import (
	"context"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

var connectionSQLServerSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("connection", true, false),
	"schema_name":        SchemaNameSchema("connection", false),
	"database_name":      DatabaseNameSchema("connection", false),
	"qualified_sql_name": QualifiedNameSchema("connection"),
	"comment":            CommentSchema(false),
	"database": {
		Description: "The target SQL Server database.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"host": {
		Description: "The SQL Server database hostname.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"port": {
		Description: "The SQL Server database port.",
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     1433,
		ForceNew:    true,
	},
	"user": ValueSecretSchema("user", "The SQL Server database username.", true, true),
	"password": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "password",
		Description: "The SQL Server database password.",
		Required:    false,
		ForceNew:    true,
	}),
	"ssh_tunnel": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "ssh_tunnel",
		Description: "The SSH tunnel configuration for the SQL Server database.",
		Required:    false,
		ForceNew:    true,
	}),
	"aws_privatelink": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "aws_privatelink",
		Description: "The AWS PrivateLink configuration for the SQL Server database.",
		Required:    false,
		ForceNew:    true,
	}),
	"validate":       ValidateConnectionSchema(),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func ConnectionSQLServer() *schema.Resource {
	return &schema.Resource{
		Description: "A SQL Server connection establishes a link to a single database of a SQL Server instance.",

		CreateContext: connectionSQLServerCreate,
		ReadContext:   connectionRead,
		UpdateContext: connectionUpdate,
		DeleteContext: connectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionSQLServerSchema,
	}
}

func connectionSQLServerCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "CONNECTION", Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewConnectionSQLServerBuilder(metaDb, o)

	if v, ok := d.GetOk("connection_type"); ok {
		b.ConnectionType(v.(string))
	}

	if v, ok := d.GetOk("host"); ok {
		b.SQLServerHost(v.(string))
	}

	if v, ok := d.GetOk("port"); ok {
		b.SQLServerPort(v.(int))
	}

	if v, ok := d.GetOk("user"); ok {
		user := materialize.GetValueSecretStruct(v)
		b.SQLServerUser(user)
	}

	if v, ok := d.GetOk("password"); ok {
		pass := materialize.GetIdentifierSchemaStruct(v)
		b.SQLServerPassword(pass)
	}

	if v, ok := d.GetOk("database"); ok {
		b.SQLServerDatabase(v.(string))
	}

	if v, ok := d.GetOk("aws_privatelink"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.SQLServerAWSPrivateLink(conn)
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.SQLServerSSHTunnel(conn)
	}

	if v, ok := d.GetOk("validate"); ok {
		b.Validate(v.(bool))
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
	i, err := materialize.ConnectionId(metaDb, o)
	if err != nil {
		return diag.FromErr(err)
	}
	d.SetId(utils.TransformIdWithRegion(string(region), i))

	return connectionRead(ctx, d, meta)
}
