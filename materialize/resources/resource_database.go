package resources

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
)

func Database() *schema.Resource {
	return &schema.Resource{
		Description: "The highest level namespace hierarchy in Materialize.",

		CreateContext: resourceDatabaseCreate,
		ReadContext:   resourceDatabaseRead,
		DeleteContext: resourceDatabaseDelete,

		Schema: map[string]*schema.Schema{
			"name": {
				Description: "The identifier for the database.",
				Type:        schema.TypeString,
				Optional:    true,
				ForceNew:    true,
			},
		},
	}
}

type DatabaseBuilder struct {
	databaseName string
}

func newDatabaseBuilder(databaseName string) *DatabaseBuilder {
	return &DatabaseBuilder{
		databaseName: databaseName,
	}
}

func (b *DatabaseBuilder) Create() string {
	return fmt.Sprintf(`CREATE DATABASE %s;`, b.databaseName)
}

func (b *DatabaseBuilder) Read() string {
	return fmt.Sprintf(`SELECT id, name FROM mz_databases WHERE name = '%s';`, b.databaseName)
}

func (b *DatabaseBuilder) Drop() string {
	return fmt.Sprintf(`DROP DATABASE %s;`, b.databaseName)
}

func resourceDatabaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)
	databaseName := d.Get("name").(string)

	builder := newDatabaseBuilder(databaseName)
	q := builder.Read()

	var id, name string
	conn.QueryRow(q).Scan(&id, &name)

	d.SetId(id)

	return diags
}

func resourceDatabaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sql.DB)
	databaseName := d.Get("name").(string)

	builder := newDatabaseBuilder(databaseName)
	q := builder.Create()

	if err := ExecResource(conn, q); err != nil {
		return diag.FromErr(err)
	}
	return resourceDatabaseRead(ctx, d, meta)
}

func resourceDatabaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	var diags diag.Diagnostics

	conn := meta.(*sql.DB)
	databaseName := d.Get("name").(string)

	builder := newDatabaseBuilder(databaseName)
	q := builder.Drop()

	if err := ExecResource(conn, q); err != nil {
		return diag.FromErr(err)
	}
	return diags
}
