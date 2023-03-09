package resources

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var indexSchema = map[string]*schema.Schema{
	"name": {
		Description:  "The identifier for the index.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		ExactlyOneOf: []string{"name", "default"},
	},
	"default": {
		Description:  "Creates a default index using a set of columns that uniquely identify each row. If this set of columns can’t be inferred, all columns are used.",
		Type:         schema.TypeBool,
		Optional:     true,
		ForceNew:     true,
		ExactlyOneOf: []string{"name", "default"},
	},
	"obj_name": {
		Description: "The name of the source, view, or materialized view on which you want to create an index..",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"cluster_name": {
		Description: "The cluster to maintain this index. If not specified, defaults to the active cluster.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"method": {
		Description:  "The name of the index method to use.",
		Type:         schema.TypeString,
		Optional:     true,
		ForceNew:     true,
		Default:      "ARRANGEMENT",
		ValidateFunc: validation.StringInSlice([]string{"ARRANGEMENT"}, true),
	},
	"col_expr": {
		Description: "The expressions to use as the key for the index.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"field": {
					Description: "The name of the option you want to set.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"val": {
					Description: "The value for the option.",
					Type:        schema.TypeString,
					Optional:    true,
				},
			},
		},
		Optional: true,
		ForceNew: true,
	},
}

func Index() *schema.Resource {
	return &schema.Resource{
		Description: "An in-memory index on a source, view, or materialized view.",

		CreateContext: indexCreate,
		ReadContext:   indexRead,
		DeleteContext: indexDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: indexSchema,
	}
}

type IndexColumn struct {
	field string
	val   string
}

type IndexBuilder struct {
	indexName    string
	indexDefault bool
	objName      string
	clusterName  string
	method       string
	colExpr      []IndexColumn
}

func newIndexBuilder(indexName string) *IndexBuilder {
	return &IndexBuilder{
		indexName: indexName,
	}
}

func (b *IndexBuilder) IndexDefault() *IndexBuilder {
	b.indexDefault = true
	return b
}

func (b *IndexBuilder) ObjName(o string) *IndexBuilder {
	b.objName = o
	return b
}

func (b *IndexBuilder) ClusterName(c string) *IndexBuilder {
	b.clusterName = c
	return b
}

func (b *IndexBuilder) Method(m string) *IndexBuilder {
	b.method = m
	return b
}

func (b *IndexBuilder) ColExpr(c []IndexColumn) *IndexBuilder {
	b.colExpr = c
	return b
}

func (b *IndexBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(`CREATE`)

	if b.indexDefault {
		q.WriteString(` DEFAULT`)
	} else {
		q.WriteString(fmt.Sprintf(` INDEX %s`, b.indexName))
	}

	q.WriteString(fmt.Sprintf(` IN CLUSTER %s ON %s USING %s`, b.clusterName, b.objName, b.method))

	if len(b.colExpr) > 0 {
		var columns []string

		for _, c := range b.colExpr {
			s := strings.Builder{}

			s.WriteString(fmt.Sprintf(`%s %s`, c.field, c.val))
			o := s.String()
			columns = append(columns, o)

		}
		p := strings.Join(columns[:], ", ")
		q.WriteString(fmt.Sprintf(` (%s)`, p))
	} else {
		q.WriteString(` ()`)
	}

	q.WriteString(`;`)
	return q.String()
}

func (b *IndexBuilder) Drop() string {
	return fmt.Sprintf(`DROP INDEX %s;`, b.indexName)
}

func (b *IndexBuilder) ReadId() string {
	return fmt.Sprintf(`SELECT id FROM mz_indexes WHERE name = '%s';`, b.indexName)
}

func readIndexParams(id string) string {
	return fmt.Sprintf("SELECT name FROM mz_indexes WHERE id = '%s';", id)
}

func indexRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readIndexParams(i)

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

func indexCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	indexName := d.Get("name").(string)

	builder := newIndexBuilder(indexName)

	if v, ok := d.GetOk("default"); ok && v.(bool) {
		builder.IndexDefault()
	}

	if v, ok := d.GetOk("obj_name"); ok {
		builder.ObjName(v.(string))
	}

	if v, ok := d.GetOk("cluster_name"); ok {
		builder.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("method"); ok {
		builder.Method(v.(string))
	}

	if v, ok := d.GetOk("col_expr"); ok {
		var colExprs []IndexColumn
		for _, colExpr := range v.([]interface{}) {
			b := colExpr.(map[string]interface{})
			colExprs = append(colExprs, IndexColumn{
				field: b["field"].(string),
				val:   b["val"].(string),
			})
		}
		builder.ColExpr(colExprs)
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "index"); err != nil {
		return diag.FromErr(err)
	}
	return indexRead(ctx, d, meta)
}

func indexDelete(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	indexName := d.Get("name").(string)

	q := newIndexBuilder(indexName).Drop()

	if err := dropResource(conn, d, q, "index"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
