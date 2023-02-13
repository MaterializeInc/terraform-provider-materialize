package resources

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var databaseSchema = map[string]*schema.Schema{
	"name": {
		Description: "The identifier for the database.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
}

func Database() *schema.Resource {
	return &schema.Resource{
		Description: "The highest level namespace hierarchy in Materialize.",

		CreateContext: databaseCreate,
		ReadContext:   databaseRead,
		DeleteContext: databaseDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: databaseSchema,
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

func (b *DatabaseBuilder) ReadId() string {
	return fmt.Sprintf(`SELECT id FROM mz_databases WHERE name = '%s';`, b.databaseName)
}

func (b *DatabaseBuilder) Drop() string {
	return fmt.Sprintf(`DROP DATABASE %s;`, b.databaseName)
}

func readDatabaseParams(id string) string {
	return fmt.Sprintf("SELECT name FROM mz_databases WHERE id = '%s';", id)
}

//lint:ignore U1000 Ignore unused function temporarily for debugging
type _database struct {
	name sql.NullString `db:"name"`
}

func databaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readDatabaseParams(i)

	readResource(conn, d, i, q, _database{}, "database")
	return nil
}

func databaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	databaseName := d.Get("name").(string)

	builder := newDatabaseBuilder(databaseName)
	qc := builder.Create()
	qr := builder.ReadId()

	createResource(conn, d, qc, qr, "database")
	return databaseRead(ctx, d, meta)
}

func databaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	databaseName := d.Get("name").(string)

	builder := newDatabaseBuilder(databaseName)
	q := builder.Drop()

	dropResource(conn, d, q, "database")
	return nil
}
