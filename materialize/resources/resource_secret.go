package resources

import (
	"context"
	"fmt"
	"log"

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

type SecretBuilder struct {
	secretName   string
	schemaName   string
	databaseName string
}

func (b *SecretBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.secretName)
}

func newSecretBuilder(secretName, schemaName, databaseName string) *SecretBuilder {
	return &SecretBuilder{
		secretName:   secretName,
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (b *SecretBuilder) Create(value string) string {
	return fmt.Sprintf(`CREATE SECRET %s AS %s;`, b.qualifiedName(), QuoteString(value))
}

func (b *SecretBuilder) Rename(newName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newName)
	return fmt.Sprintf(`ALTER SECRET %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *SecretBuilder) UpdateValue(newValue string) string {
	return fmt.Sprintf(`ALTER SECRET %s AS %s;`, b.qualifiedName(), QuoteString(newValue))
}

func (b *SecretBuilder) Drop() string {
	return fmt.Sprintf(`DROP SECRET %s;`, b.qualifiedName())
}

func (b *SecretBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_secrets.id
		FROM mz_secrets
		JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_secrets.name = %s
		AND mz_schemas.name = %s
		AND mz_databases.name = %s;`, QuoteString(b.secretName), QuoteString(b.schemaName), QuoteString(b.databaseName))
}

func readSecretParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_secrets.name AS name,
			mz_schemas.name AS schema_name,
			mz_databases.name AS database_name
		FROM mz_secrets
		JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_secrets.id = %s;`, QuoteString(id))
}

func secretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readSecretParams(i)

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

	qn := fmt.Sprintf("%s.%s.%s", database, schema, name)
	if err := d.Set("qualified_name", qn); err != nil {
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

	builder := newSecretBuilder(secretName, schemaName, databaseName)
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

	if d.HasChange("name") {
		_, newName := d.GetChange("name")

		q := newSecretBuilder(secretName, schemaName, databaseName).Rename(newName.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not rename secret: %s", q)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("value") {
		_, newValue := d.GetChange("value")

		q := newSecretBuilder(secretName, schemaName, databaseName).UpdateValue(newValue.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not update value of secret: %s", q)
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

	q := newSecretBuilder(secretName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "secret"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
