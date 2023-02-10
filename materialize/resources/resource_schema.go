package resources

import (
	"context"
	"database/sql"
	"fmt"
	"log"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Schema() *schema.Resource {
	return &schema.Resource{
		Description: "The highest level namespace hierarchy in Materialize.",

		CreateContext: resourceSchemaCreate,
		ReadContext:   resourceSchemaRead,
		DeleteContext: resourceSchemaDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The name of the schema.",
				Type:        schema.TypeString,
				Required:    true,
				ForceNew:    true,
			},
			"database_name": {
				Description: "The name of the database.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
				Default:     "materialize",
			},
		},
	}
}

type SchemaBuilder struct {
	schemaName   string
	databaseName string
}

func newSchemaBuilder(schemaName, databaseName string) *SchemaBuilder {
	return &SchemaBuilder{
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (b *SchemaBuilder) Create() string {
	return fmt.Sprintf(`CREATE SCHEMA %s.%s;`, b.databaseName, b.schemaName)
}

func (b *SchemaBuilder) Read() string {
	return fmt.Sprintf(`
		SELECT
			mz_schemas.id,
			mz_schemas.name,
			mz_databases.name
		FROM mz_schemas JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_schemas.name = '%s'
		AND mz_databases.name = '%s';	
	`, b.schemaName, b.databaseName)
}

func (b *SchemaBuilder) Drop() string {
	return fmt.Sprintf(`DROP SCHEMA %s.%s;`, b.databaseName, b.schemaName)
}

func resourceSchemaRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newSchemaBuilder(schemaName, databaseName)
	q := builder.Read()

	var id, name, database string
	conn.QueryRow(q).Scan(&id, &name, &database)

	d.SetId(id)

	return diags
}

func resourceSchemaCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sql.DB)
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newSchemaBuilder(schemaName, databaseName)
	q := builder.Create()

	if err := ExecResource(conn, q); err != nil {
		log.Printf("[ERROR] could not execute query: %s", q)
		return diag.FromErr(err)
	}
	return resourceSchemaRead(ctx, d, meta)
}

func resourceSchemaDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newSchemaBuilder(schemaName, databaseName)
	q := builder.Drop()

	if err := ExecResource(conn, q); err != nil {
		log.Printf("[ERROR] could not execute query: %s", q)
		return diag.FromErr(err)
	}
	return diags
}
