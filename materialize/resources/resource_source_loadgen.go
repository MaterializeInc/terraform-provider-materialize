package resources

import (
	"context"
	"fmt"
	"log"
	"sort"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/validation"
	"github.com/jmoiron/sqlx"
)

var sourceLoadgenSchema = map[string]*schema.Schema{
	"name": {
		Description: "The identifier for the source.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"schema_name": {
		Description: "The identifier for the source schema.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "public",
	},
	"database_name": {
		Description: "The identifier for the source database.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "materialize",
	},
	"qualified_name": {
		Description: "The fully qualified name of the source.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"cluster_name": {
		Description:   "The cluster to maintain this source. If not specified, the size option must be specified.",
		Type:          schema.TypeString,
		Optional:      true,
		ConflictsWith: []string{"size"},
	},
	"size": {
		Description:   "The size of the source.",
		Type:          schema.TypeString,
		Optional:      true,
		ForceNew:      true,
		ValidateFunc:  validation.StringInSlice(append(sourceSizes, localSizes...), true),
		ConflictsWith: []string{"cluster_name"},
	},
	"load_generator_type": {
		Description:  fmt.Sprintf("The load generator types: %s.", loadGeneratorTypes),
		Type:         schema.TypeString,
		Required:     true,
		ValidateFunc: validation.StringInSlice(loadGeneratorTypes, true),
	},
	"tick_interval": {
		Description: "The interval at which the next datum should be emitted. Defaults to one second.",
		Type:        schema.TypeString,
		Optional:    true,
		ForceNew:    true,
	},
	"scale_factor": {
		Description: "The scale factor for the TPCH generator. Defaults to 0.01 (~ 10MB).",
		Type:        schema.TypeFloat,
		Optional:    true,
		Default:     0.01,
		ForceNew:    true,
	},
	"tables": {
		Description: "Creates subsources for specific tables in the load generator.",
		Type:        schema.TypeMap,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
		ForceNew:    true,
	},
}

func SourceLoadgen() *schema.Resource {
	return &schema.Resource{
		Description: "A source describes an external system you want Materialize to read data from, and provides details about how to decode and interpret that data.",

		CreateContext: sourceLoadgenCreate,
		ReadContext:   sourceLoadgenRead,
		UpdateContext: sourceLoadgenUpdate,
		DeleteContext: sourceLoadgenDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourceLoadgenSchema,
	}
}

type SourceLoadgenBuilder struct {
	sourceName        string
	schemaName        string
	databaseName      string
	clusterName       string
	size              string
	loadGeneratorType string
	tickInterval      string
	scaleFactor       float64
	tables            map[string]string
}

func newSourceLoadgenBuilder(sourceName, schemaName, databaseName string) *SourceLoadgenBuilder {
	return &SourceLoadgenBuilder{
		sourceName:   sourceName,
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (b *SourceLoadgenBuilder) ClusterName(c string) *SourceLoadgenBuilder {
	b.clusterName = c
	return b
}

func (b *SourceLoadgenBuilder) Size(s string) *SourceLoadgenBuilder {
	b.size = s
	return b
}

func (b *SourceLoadgenBuilder) LoadGeneratorType(l string) *SourceLoadgenBuilder {
	b.loadGeneratorType = l
	return b
}

func (b *SourceLoadgenBuilder) TickInterval(t string) *SourceLoadgenBuilder {
	b.tickInterval = t
	return b
}

func (b *SourceLoadgenBuilder) ScaleFactor(s float64) *SourceLoadgenBuilder {
	b.scaleFactor = s
	return b
}

func (b *SourceLoadgenBuilder) Tables(t map[string]string) *SourceLoadgenBuilder {
	b.tables = t
	return b
}

func (b *SourceLoadgenBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s.%s.%s`, b.databaseName, b.schemaName, b.sourceName))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, b.clusterName))
	}

	q.WriteString(fmt.Sprintf(` FROM LOAD GENERATOR %s`, b.loadGeneratorType))

	if b.tickInterval != "" || b.scaleFactor != 0 {
		var p []string
		if b.tickInterval != "" {
			t := fmt.Sprintf(`TICK INTERVAL '%s'`, b.tickInterval)
			p = append(p, t)
		}

		if b.scaleFactor != 0 {
			s := fmt.Sprintf(`SCALE FACTOR %.2f`, b.scaleFactor)
			p = append(p, s)
		}

		if len(p) != 0 {
			p := strings.Join(p[:], ", ")
			q.WriteString(fmt.Sprintf(` (%s)`, p))
		}
	}

	var o []string
	if b.loadGeneratorType == "COUNTER" {

	} else if len(b.tables) > 0 {
		// Need to sort tables to ensure order for tests
		var keys []string
		for k := range b.tables {
			keys = append(keys, k)
		}
		sort.Strings(keys)

		for _, k := range keys {
			s := fmt.Sprintf(`%s AS %s`, k, b.tables[k])
			o = append(o, s)
		}
		o := strings.Join(o[:], ", ")
		q.WriteString(fmt.Sprintf(` FOR TABLES (%s)`, o))
	} else {
		q.WriteString(` FOR ALL TABLES`)
	}

	if b.size != "" {
		q.WriteString(fmt.Sprintf(` WITH (SIZE = '%s')`, b.size))
	}

	q.WriteString(`;`)
	return q.String()
}

func (b *SourceLoadgenBuilder) ReadId() string {
	return readsourceId(b.sourceName, b.schemaName, b.databaseName)
}

func (b *SourceLoadgenBuilder) Rename(newName string) string {
	return fmt.Sprintf(`ALTER SOURCE %s.%s.%s RENAME TO %s.%s.%s;`, b.databaseName, b.schemaName, b.sourceName, b.databaseName, b.schemaName, newName)
}

func (b *SourceLoadgenBuilder) UpdateSize(newSize string) string {
	return fmt.Sprintf(`ALTER SOURCE %s.%s.%s SET (SIZE = '%s');`, b.databaseName, b.schemaName, b.sourceName, newSize)
}

func (b *SourceLoadgenBuilder) Drop() string {
	return fmt.Sprintf(`DROP SOURCE %s.%s.%s;`, b.databaseName, b.schemaName, b.sourceName)
}

func sourceLoadgenRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readSourceParams(i)

	readResource(conn, d, i, q, _source{}, "source")
	setQualifiedName(d)
	return nil
}

func sourceLoadgenCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newSourceLoadgenBuilder(sourceName, schemaName, databaseName)

	if v, ok := d.GetOk("cluster_name"); ok {
		builder.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("size"); ok {
		builder.Size(v.(string))
	}

	if v, ok := d.GetOk("load_generator_type"); ok {
		builder.LoadGeneratorType(v.(string))
	}

	if v, ok := d.GetOk("tick_interval"); ok {
		builder.TickInterval(v.(string))
	}

	if v, ok := d.GetOk("scale_factor"); ok {
		builder.ScaleFactor(v.(float64))
	}

	if v, ok := d.GetOk("tables"); ok {
		builder.Tables(v.(map[string]string))
	}

	qc := builder.Create()
	qr := builder.ReadId()

	createResource(conn, d, qc, qr, "source")
	return sourceLoadgenRead(ctx, d, meta)
}

func sourceLoadgenUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")

		builder := newSourceLoadgenBuilder(oldName.(string), schemaName, databaseName)
		q := builder.Rename(newName.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("size") {
		sourceName := d.Get("sourceName").(string)
		_, newSize := d.GetChange("size")

		builder := newSourceLoadgenBuilder(sourceName, schemaName, databaseName)
		q := builder.UpdateSize(newSize.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return sourceLoadgenRead(ctx, d, meta)
}

func sourceLoadgenDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newSourceLoadgenBuilder(sourceName, schemaName, databaseName)
	q := builder.Drop()

	dropResource(conn, d, q, "source")
	return nil
}
