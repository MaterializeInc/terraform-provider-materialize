package resources

import (
	"context"
	"fmt"
	"log"
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
	"source_type": {
		Description: "The type of source.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"cluster_name": {
		Description:  "The cluster to maintain this source. If not specified, the size option must be specified.",
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ExactlyOneOf: []string{"cluster_name", "size"},
	},
	"size": {
		Description:  "The size of the source.",
		Type:         schema.TypeString,
		Optional:     true,
		Computed:     true,
		ExactlyOneOf: []string{"cluster_name", "size"},
		ValidateFunc: validation.StringInSlice(append(sourceSizes, localSizes...), true),
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
	"max_cardinality": {
		Description: "Valid for the COUNTER generator. Causes the generator to delete old values to keep the collection at most a given size. Defaults to unlimited.",
		Type:        schema.TypeBool,
		Optional:    true,
		ForceNew:    true,
	},
	"tables": {
		Description: "Creates subsources for specific tables.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"name": {
					Description: "The name of the table.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"alias": {
					Description: "The alias of the table.",
					Type:        schema.TypeString,
					Optional:    true,
				},
			},
		},
		Optional: true,
		MinItems: 1,
		ForceNew: true,
	},
}

func SourceLoadgen() *schema.Resource {
	return &schema.Resource{
		Description: "A source describes an external system you want Materialize to read data from, and provides details about how to decode and interpret that data.",

		CreateContext: sourceLoadgenCreate,
		ReadContext:   SourceRead,
		UpdateContext: sourceLoadgenUpdate,
		DeleteContext: sourceLoadgenDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourceLoadgenSchema,
	}
}

type TableLoadgen struct {
	name  string
	alias string
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
	maxCardinality    bool
	tables            []TableLoadgen
}

func (b *SourceLoadgenBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.sourceName)
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

func (b *SourceLoadgenBuilder) MaxCardinality(m bool) *SourceLoadgenBuilder {
	b.maxCardinality = m
	return b
}

func (b *SourceLoadgenBuilder) Tables(t []TableLoadgen) *SourceLoadgenBuilder {
	b.tables = t
	return b
}

func (b *SourceLoadgenBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s`, b.qualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, b.clusterName))
	}

	q.WriteString(fmt.Sprintf(` FROM LOAD GENERATOR %s`, b.loadGeneratorType))

	if b.tickInterval != "" || b.scaleFactor != 0 || b.maxCardinality {
		var p []string
		if b.tickInterval != "" {
			t := fmt.Sprintf(`TICK INTERVAL %s`, QuoteString(b.tickInterval))
			p = append(p, t)
		}

		if b.scaleFactor != 0 {
			s := fmt.Sprintf(`SCALE FACTOR %.2f`, b.scaleFactor)
			p = append(p, s)
		}

		if b.maxCardinality {
			p = append(p, ` MAX CARDINALITY`)
		}

		if len(p) != 0 {
			p := strings.Join(p[:], ", ")
			q.WriteString(fmt.Sprintf(` (%s)`, p))
		}
	}

	if b.loadGeneratorType == "COUNTER" {
		// Tables do not apply to COUNTER
	} else if len(b.tables) > 0 {

		var tables []string
		for _, t := range b.tables {
			if t.alias == "" {
				t.alias = t.name
			}
			s := fmt.Sprintf(`%s AS %s`, t.name, t.alias)
			tables = append(tables, s)
		}
		o := strings.Join(tables[:], ", ")
		q.WriteString(fmt.Sprintf(` FOR TABLES (%s)`, o))
	} else {
		q.WriteString(` FOR ALL TABLES`)
	}

	if b.size != "" {
		q.WriteString(fmt.Sprintf(` WITH (SIZE = %s)`, QuoteString(b.size)))
	}

	q.WriteString(`;`)
	return q.String()
}

func (b *SourceLoadgenBuilder) Rename(newName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newName)
	return fmt.Sprintf(`ALTER SOURCE %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *SourceLoadgenBuilder) UpdateSize(newSize string) string {
	return fmt.Sprintf(`ALTER SOURCE %s SET (SIZE = %s);`, b.qualifiedName(), QuoteString(newSize))
}

func (b *SourceLoadgenBuilder) Drop() string {
	return fmt.Sprintf(`DROP SOURCE %s;`, b.qualifiedName())
}

func (b *SourceLoadgenBuilder) ReadId() string {
	return readSourceId(b.sourceName, b.schemaName, b.databaseName)
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

	if v, ok := d.GetOk("max_cardinality"); ok {
		builder.MaxCardinality(v.(bool))
	}

	if v, ok := d.GetOk("tables"); ok {
		var tables []TableLoadgen
		for _, table := range v.([]interface{}) {
			t := table.(map[string]interface{})
			tables = append(tables, TableLoadgen{
				name:  t["name"].(string),
				alias: t["alias"].(string),
			})
		}
		builder.Tables(tables)
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "source"); err != nil {
		return diag.FromErr(err)
	}
	return SourceRead(ctx, d, meta)
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

	return SourceRead(ctx, d, meta)
}

func sourceLoadgenDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := newSourceLoadgenBuilder(sourceName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "source"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
