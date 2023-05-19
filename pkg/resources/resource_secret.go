package resources

import (
	"context"
	"database/sql"
	"log"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var secretSchema = map[string]*schema.Schema{
	"name":               NameSchema("secret", true, false),
	"schema_name":        SchemaNameSchema("secret", false),
	"database_name":      DatabaseNameSchema("secret", false),
	"qualified_sql_name": QualifiedNameSchema("secret"),
	"value": {
		Description: "The value for the secret. The value expression may not reference any relations, and must be a bytea string literal.",
		Type:        schema.TypeString,
		Optional:    true,
		Sensitive:   true,
	},
}

func Secret() *schema.Resource {
	return &schema.Resource{
		Description: "A secret securely stores sensitive credentials (like passwords and SSL keys) in Materialize’s secret management system.",

		CreateContext: secretCreate,
		ReadContext:   secretRead,
		UpdateContext: secretUpdate,
		DeleteContext: secretDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: secretSchema,
	}
}

type SecretParams struct {
	SecretName   sql.NullString `db:"name"`
	SchemaName   sql.NullString `db:"schema_name"`
	DatabaseName sql.NullString `db:"database_name"`
}

func secretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := materialize.ReadSecretParams(i)

	var s SecretParams
	if err := conn.Get(&s, q); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", s.SecretName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", s.SchemaName.String); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", s.DatabaseName.String); err != nil {
		return diag.FromErr(err)
	}

	qn := materialize.QualifiedName(s.DatabaseName.String, s.SchemaName.String, s.SecretName.String)
	if err := d.Set("qualified_sql_name", qn); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func secretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
	return secretRead(ctx, d, meta)
}

func secretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	secretName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("value") {
		_, newValue := d.GetChange("value")

		q := materialize.NewSecretBuilder(secretName, schemaName, databaseName).UpdateValue(newValue.(string))

		if err := execResource(conn, q); err != nil {
			log.Printf("[ERROR] could not update value of secret: %s", q)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("name") {
		_, newName := d.GetChange("name")

		q := materialize.NewSecretBuilder(secretName, schemaName, databaseName).Rename(newName.(string))

		if err := execResource(conn, q); err != nil {
			log.Printf("[ERROR] could not rename secret: %s", q)
			return diag.FromErr(err)
		}

	}

	return secretRead(ctx, d, meta)
}

func secretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
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
