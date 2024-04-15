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

var connectionMySQLSchema = map[string]*schema.Schema{
	"name":               ObjectNameSchema("connection", true, false),
	"schema_name":        SchemaNameSchema("connection", false),
	"database_name":      DatabaseNameSchema("connection", false),
	"qualified_sql_name": QualifiedNameSchema("connection"),
	"comment":            CommentSchema(false),
	"host": {
		Description: "The MySQL database hostname.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    false,
	},
	"port": {
		Description: "The MySQL database port.",
		Type:        schema.TypeInt,
		Optional:    true,
		Default:     3306,
		ForceNew:    false,
	},
	"user": ValueSecretSchema("user", "The MySQL database username.", true, false),
	"password": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "password",
		Description: "The MySQL database password.",
		Required:    false,
		ForceNew:    false,
	}),
	"ssh_tunnel": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "ssh_tunnel",
		Description: "The SSH tunnel configuration for the MySQL database.",
		Required:    false,
		ForceNew:    false,
	}),
	"ssl_certificate_authority": ValueSecretSchema("ssl_certificate_authority", "The CA certificate for the MySQL database.", false, false),
	"ssl_certificate":           ValueSecretSchema("ssl_certificate", "The client certificate for the MySQL database.", false, false),
	"ssl_key": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "ssl_key",
		Description: "The client key for the MySQL database.",
		Required:    false,
		ForceNew:    false,
	}),
	"ssl_mode": {
		Description:  "The SSL mode for the MySQL database. Allowed values are " + strings.Join(mysqlSSLMode, ", ") + ".",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     false,
		ValidateFunc: validation.StringInSlice(mysqlSSLMode, true),
	},
	"aws_privatelink": IdentifierSchema(IdentifierSchemaParams{
		Elem:        "aws_privatelink",
		Description: "The AWS PrivateLink configuration for the MySQL database.",
		Required:    false,
		ForceNew:    false,
	}),
	"validate":       ValidateConnectionSchema(),
	"ownership_role": OwnershipRoleSchema(),
	"region":         RegionSchema(),
}

func ConnectionMySQL() *schema.Resource {
	return &schema.Resource{
		Description: "A MySQL connection establishes a link to a single database of a MySQL server.",

		CreateContext: connectionMySQLCreate,
		ReadContext:   connectionRead,
		UpdateContext: connectionUpdate,
		DeleteContext: connectionDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: connectionMySQLSchema,
	}
}

func connectionMySQLCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	connectionName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	metaDb, region, err := utils.GetDBClientFromMeta(meta, d)
	if err != nil {
		return diag.FromErr(err)
	}
	o := materialize.MaterializeObject{ObjectType: "CONNECTION", Name: connectionName, SchemaName: schemaName, DatabaseName: databaseName}
	b := materialize.NewConnectionMySQLBuilder(metaDb, o)

	if v, ok := d.GetOk("host"); ok {
		b.MySQLHost(v.(string))
	}

	if v, ok := d.GetOk("port"); ok {
		b.MySQLPort(v.(int))
	}

	if v, ok := d.GetOk("user"); ok {
		user := materialize.GetValueSecretStruct(v)
		b.MySQLUser(user)
	}

	if v, ok := d.GetOk("password"); ok {
		pass := materialize.GetIdentifierSchemaStruct(v)
		b.MySQLPassword(pass)
	}

	if v, ok := d.GetOk("ssl_mode"); ok {
		b.MySQLSSLMode(v.(string))
	}

	if v, ok := d.GetOk("ssl_certificate_authority"); ok {
		ssl_ca := materialize.GetValueSecretStruct(v)
		b.MySQLSSLCa(ssl_ca)
	}

	if v, ok := d.GetOk("ssl_certificate"); ok {
		ssl_cert := materialize.GetValueSecretStruct(v)
		b.MySQLSSLCert(ssl_cert)
	}

	if v, ok := d.GetOk("ssl_key"); ok {
		k := materialize.GetIdentifierSchemaStruct(v)
		b.MySQLSSLKey(k)
	}

	if v, ok := d.GetOk("ssh_tunnel"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.MySQLSSHTunnel(conn)
	}

	if v, ok := d.GetOk("aws_privatelink"); ok {
		conn := materialize.GetIdentifierSchemaStruct(v)
		b.MySQLAWSPrivateLink(conn)
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
