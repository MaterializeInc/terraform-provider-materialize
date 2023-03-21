package resources

import (
	"context"
	"terraform-materialize/pkg/materialize"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var databaseSchema = map[string]*schema.Schema{
	"name": {
		Description: "The identifier for the database.",
		Type:        schema.TypeString,
		Required:    true,
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

func databaseRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := materialize.ReadDatabaseParams(i)

	var name string
	if err := conn.QueryRowx(q).Scan(&name); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func databaseCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	databaseName := d.Get("name").(string)

	builder := materialize.NewDatabaseBuilder(databaseName)
	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "database"); err != nil {
		return diag.FromErr(err)
	}
	return databaseRead(ctx, d, meta)
}

func databaseDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	databaseName := d.Get("name").(string)

	q := materialize.NewDatabaseBuilder(databaseName).Drop()

	if err := dropResource(conn, d, q, "database"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
