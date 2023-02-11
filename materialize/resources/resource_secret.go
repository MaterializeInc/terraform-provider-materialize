package resources

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

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
				Optional:    true,
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

func (b *SecretBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_secrets.id
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

func ReadParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_secrets.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_secrets
		JOIN mz_schemas
			ON mz_secrets.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_secrets.id = '%s';
	`, id)
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
type secret struct {
	name          sql.NullString `db:"name"`
	schema_name   sql.NullString `db:"schema_name"`
	database_name sql.NullString `db:"database_name"`
}

func secretRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()

	q := ReadParams(i)

	readResource(conn, d, i, q, secret{})
	return nil
}

func secretCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	secretName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)
	value := d.Get("value").(string)

	builder := newSecretBuilder(secretName, schemaName, databaseName)

	createResource(conn, d, builder.Create(value), builder.ReadId())
	return secretRead(ctx, d, meta)
}

func secretUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
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

	return secretRead(ctx, d, meta)
}

func secretDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	secretName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newSecretBuilder(secretName, schemaName, databaseName)

	dropResource(conn, d, builder.Drop())
	return nil
}
