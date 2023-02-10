package resources

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Secret() *schema.Resource {
	return &schema.Resource{
		Description: "A secret securely stores sensitive credentials (like passwords and SSL keys) in Materialize’s secret management system.",

		CreateContext: resourceSecretCreate,
		ReadContext:   resourceSecretRead,
		UpdateContext: resourceSecretUpdate,
		DeleteContext: resourceSecretDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The identifier for the secret.",
				Type:        schema.TypeString,
				Required:    true,
			},
			"schema_name": {
				Description: "The identifier for the secret schema.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "public",
			},
			"database_name": {
				Description: "The identifier for the secret database.",
				Type:        schema.TypeString,
				Optional:    true,
				Default:     "materialize",
			},
			"value": {
				Description: "The value for the secret. The value expression may not reference any relations, and must be implicitly castable to bytea.",
				Type:        schema.TypeString,
				Required:    true,
				Sensitive:   true,
			},
		},
	}
}

type SecretBuilder struct {
	secretName   string
	schemaName   string
	databaseName string
}

func newSecretBuilder(secretName, schemaName, databaseName string) *SecretBuilder {
	return &SecretBuilder{
		secretName:   secretName,
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (b *SecretBuilder) Create(value string) string {
	return fmt.Sprintf(`CREATE SECRET %s.%s.%s AS %s;`, b.databaseName, b.schemaName, b.secretName, value)
}

func (b *SecretBuilder) Read() string {
	return fmt.Sprintf(`
		SELECT
			mz_secrets.id,
			mz_secrets.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_secrets
		JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_secrets.name = '%s'
		AND mz_schemas.name = '%s'
		AND mz_databases.name = '%s';
	`, b.secretName, b.schemaName, b.databaseName)
}

func (b *SecretBuilder) Rename(newName string) string {
	return fmt.Sprintf(`ALTER SECRET %s.%s.%s RENAME TO %s.%s.%s;`, b.databaseName, b.schemaName, b.secretName, b.databaseName, b.schemaName, newName)
}

func (b *SecretBuilder) UpdateValue(newValue string) string {
	return fmt.Sprintf(`ALTER SECRET %s.%s.%s AS %s;`, b.databaseName, b.schemaName, b.secretName, newValue)
}

func (b *SecretBuilder) Drop() string {
	return fmt.Sprintf(`DROP SECRET %s.%s.%s;`, b.databaseName, b.schemaName, b.secretName)
}

func resourceSecretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)
	secretName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newSecretBuilder(secretName, schemaName, databaseName)
	q := builder.Read()

	var id, name, schema_name, database_name string
	conn.QueryRow(q).Scan(&id, &name, &schema_name, &database_name)

	d.SetId(id)

	return diags
}

func resourceSecretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sql.DB)
	secretName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	value := d.Get("value").(string)

	builder := newSecretBuilder(secretName, schemaName, databaseName)
	q := builder.Create(value)

	if err := ExecResource(conn, q); err != nil {
		log.Printf("[ERROR] could not execute query: %s", q)
		return diag.FromErr(err)
	}
	return resourceSecretRead(ctx, d, meta)
}

func resourceSecretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sql.DB)
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")

		builder := newSecretBuilder(oldName.(string), schemaName, databaseName)
		q := builder.Rename(newName.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("value") {
		oldValue, newValue := d.GetChange("value")

		builder := newSecretBuilder(oldValue.(string), schemaName, databaseName)
		q := builder.UpdateValue(newValue.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return resourceSecretRead(ctx, d, meta)
}

func resourceSecretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)
	secretName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newSecretBuilder(secretName, schemaName, databaseName)
	q := builder.Drop()

	if err := ExecResource(conn, q); err != nil {
		log.Printf("[ERROR] could not execute query: %s", q)
		return diag.FromErr(err)
	}
	return diags
}
