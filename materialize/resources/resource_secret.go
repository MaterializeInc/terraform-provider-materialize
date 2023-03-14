package resources

import (
	"context"
	"log"

	"terraform-materialize/materialize/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var secretSchema = map[string]*schema.Schema{
	"name": {
		Description: "The identifier for the secret.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"schema_name": {
		Description: "The schema of the secret.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "public",
		ForceNew:    true,
	},
	"database_name": {
		Description: "The database of the secret.",
		Type:        schema.TypeString,
		Optional:    true,
		DefaultFunc: schema.EnvDefaultFunc("MZ_DATABASE", "materialize"),
		ForceNew:    true,
	},
	"qualified_name": {
		Description: "The fully qualified name of the secret.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"value": {
		Description: "The value for the secret. The value expression may not reference any relations, and must be implicitly castable to bytea.",
		Type:        schema.TypeString,
		Optional:    true,
		Sensitive:   true,
	},
}

func Secret() *schema.Resource {
	return &schema.Resource{
		Description: "A secret securely stores sensitive credentials (like passwords and SSL keys) in Materialize’s secret management system.",

		CreateContext: SecretCreate,
		ReadContext:   SecretRead,
		UpdateContext: SecretUpdate,
		DeleteContext: SecretDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: secretSchema,
	}
}

func SecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := materialize.ReadSecretParams(i)

	var name, schema, database string
	if err := conn.QueryRowx(q).Scan(&name, &schema, &database); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", schema); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", database); err != nil {
		return diag.FromErr(err)
	}

	qn := QualifiedName(database, schema, name)
	if err := d.Set("qualified_name", qn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func SecretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	secretName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)
	value := d.Get("value").(string)

	builder := materialize.NewSecretBuilder(secretName, schemaName, databaseName)
	qc := builder.Create(value)
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "secret"); err != nil {
		return diag.FromErr(err)
	}
	return SecretRead(ctx, d, meta)
}

func SecretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	secretName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		_, newName := d.GetChange("name")

		q := materialize.NewSecretBuilder(secretName, schemaName, databaseName).Rename(newName.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not rename secret: %s", q)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("value") {
		_, newValue := d.GetChange("value")

		q := materialize.NewSecretBuilder(secretName, schemaName, databaseName).UpdateValue(newValue.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not update value of secret: %s", q)
			return diag.FromErr(err)
		}
	}

	return SecretRead(ctx, d, meta)
}

func SecretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	secretName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := materialize.NewSecretBuilder(secretName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "secret"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
