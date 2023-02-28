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

var sourcePostgresSchema = map[string]*schema.Schema{
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
	"postgres_connection": {
		Description: "The name of the PostgreSQL connection to use in the source.",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"publication": {
		Description: "The PostgreSQL publication (the replication data set containing the tables to be streamed to Materialize).",
		Type:        schema.TypeString,
		Required:    true,
		ForceNew:    true,
	},
	"text_columns": {
		Description: "Decode data as text for specific columns that contain PostgreSQL types that are unsupported in Materialize.",
		Type:        schema.TypeList,
		Elem:        &schema.Schema{Type: schema.TypeString},
		Optional:    true,
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

func SourcePostgres() *schema.Resource {
	return &schema.Resource{
		Description: "A source describes an external system you want Materialize to read data from, and provides details about how to decode and interpret that data.",

		CreateContext: sourcePostgresCreate,
		ReadContext:   sourcePostgresRead,
		UpdateContext: sourcePostgresUpdate,
		DeleteContext: sourcePostgresDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourcePostgresSchema,
	}
}

type SourcePostgresBuilder struct {
	sourceName         string
	schemaName         string
	databaseName       string
	clusterName        string
	size               string
	postgresConnection string
	publication        string
	textColumns        []string
	tables             map[string]string
}

func newSourcePostgresBuilder(sourceName, schemaName, databaseName string) *SourcePostgresBuilder {
	return &SourcePostgresBuilder{
		sourceName:   sourceName,
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (b *SourcePostgresBuilder) ClusterName(c string) *SourcePostgresBuilder {
	b.clusterName = c
	return b
}

func (b *SourcePostgresBuilder) Size(s string) *SourcePostgresBuilder {
	b.size = s
	return b
}

func (b *SourcePostgresBuilder) PostgresConnection(p string) *SourcePostgresBuilder {
	b.postgresConnection = p
	return b
}

func (b *SourcePostgresBuilder) Publication(p string) *SourcePostgresBuilder {
	b.publication = p
	return b
}

func (b *SourcePostgresBuilder) TextColumns(t []string) *SourcePostgresBuilder {
	b.textColumns = t
	return b
}

func (b *SourcePostgresBuilder) Tables(t map[string]string) *SourcePostgresBuilder {
	b.tables = t
	return b
}

func (b *SourcePostgresBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s.%s.%s`, b.databaseName, b.schemaName, b.sourceName))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, b.clusterName))
	}

	q.WriteString(fmt.Sprintf(` FROM POSTGRES CONNECTION %s`, b.postgresConnection))

	// Publication
	p := fmt.Sprintf(`PUBLICATION '%s'`, b.publication)

	if len(b.textColumns) > 0 {
		s := strings.Join(b.textColumns, ", ")
		p = p + fmt.Sprintf(`, TEXT COLUMNS (%s)`, s)
	}

	q.WriteString(fmt.Sprintf(` (%s)`, p))

	var o []string
	if len(b.tables) > 0 {
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

func (b *SourcePostgresBuilder) ReadId() string {
	return readsourceId(b.sourceName, b.schemaName, b.databaseName)
}

func (b *SourcePostgresBuilder) Rename(newName string) string {
	return fmt.Sprintf(`ALTER SOURCE %s.%s.%s RENAME TO %s.%s.%s;`, b.databaseName, b.schemaName, b.sourceName, b.databaseName, b.schemaName, newName)
}

func (b *SourcePostgresBuilder) UpdateSize(newSize string) string {
	return fmt.Sprintf(`ALTER SOURCE %s.%s.%s SET (SIZE = '%s');`, b.databaseName, b.schemaName, b.sourceName, newSize)
}

func (b *SourcePostgresBuilder) Drop() string {
	return fmt.Sprintf(`DROP SOURCE %s.%s.%s;`, b.databaseName, b.schemaName, b.sourceName)
}

func sourcePostgresRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readSourceParams(i)

	readResource(conn, d, i, q, _source{}, "source")
	setQualifiedName(d)
	return nil
}

func sourcePostgresCreate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newSourcePostgresBuilder(sourceName, schemaName, databaseName)

	if v, ok := d.GetOk("cluster_name"); ok {
		builder.ClusterName(v.(string))
	}

	if v, ok := d.GetOk("size"); ok {
		builder.Size(v.(string))
	}

	if v, ok := d.GetOk("postgresConnection"); ok {
		builder.PostgresConnection(v.(string))
	}

	if v, ok := d.GetOk("publication"); ok {
		builder.Publication(v.(string))
	}

	if v, ok := d.GetOk("tables"); ok {
		builder.Tables(v.(map[string]string))
	}

	if v, ok := d.GetOk("textColumns"); ok {
		builder.TextColumns(v.([]string))
	}

	qc := builder.Create()
	qr := builder.ReadId()

	createResource(conn, d, qc, qr, "source")
	return sourcePostgresRead(ctx, d, meta)
}

func sourcePostgresUpdate(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	schemaName := d.Get("name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		oldName, newName := d.GetChange("name")

		builder := newSourcePostgresBuilder(oldName.(string), schemaName, databaseName)
		q := builder.Rename(newName.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	if d.HasChange("size") {
		sourceName := d.Get("sourceName").(string)
		_, newSize := d.GetChange("size")

		builder := newSourcePostgresBuilder(sourceName, schemaName, databaseName)
		q := builder.UpdateSize(newSize.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not execute query: %s", q)
			return diag.FromErr(err)
		}
	}

	return sourcePostgresRead(ctx, d, meta)
}

func sourcePostgresDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newSourcePostgresBuilder(sourceName, schemaName, databaseName)
	q := builder.Drop()

	dropResource(conn, d, q, "source")
	return nil
}
