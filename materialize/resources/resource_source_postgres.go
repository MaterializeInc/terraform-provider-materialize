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
	"source_type": {
		Description: "The type of source.",
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

func SourcePostgres() *schema.Resource {
	return &schema.Resource{
		Description: "A source describes an external system you want Materialize to read data from, and provides details about how to decode and interpret that data.",

		CreateContext: sourcePostgresCreate,
		ReadContext:   SourceRead,
		UpdateContext: sourcePostgresUpdate,
		DeleteContext: sourcePostgresDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: sourcePostgresSchema,
	}
}

type TablePostgres struct {
	name  string
	alias string
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
	tables             []TablePostgres
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

func (b *SourcePostgresBuilder) Tables(t []TablePostgres) *SourcePostgresBuilder {
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

	if len(b.tables) > 0 {
		q.WriteString(` FOR TABLES (`)
		for i, t := range b.tables {
			if t.alias == "" {
				t.alias = t.name
			}
			q.WriteString(fmt.Sprintf(`%s AS %s`, t.name, t.alias))
			if i < len(b.tables)-1 {
				q.WriteString(`, `)
			}
		}
		q.WriteString(`)`)
	} else {
		q.WriteString(` FOR ALL TABLES`)
	}

	if b.size != "" {
		q.WriteString(fmt.Sprintf(` WITH (SIZE = '%s')`, b.size))
	}

	q.WriteString(`;`)
	return q.String()
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

func (b *SourcePostgresBuilder) ReadId() string {
	return readSourceId(b.sourceName, b.schemaName, b.databaseName)
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

	if v, ok := d.GetOk("postgres_connection"); ok {
		builder.PostgresConnection(v.(string))
	}

	if v, ok := d.GetOk("publication"); ok {
		builder.Publication(v.(string))
	}

	if v, ok := d.GetOk("tables"); ok {
		var tables []TablePostgres
		for _, table := range v.([]interface{}) {
			t := table.(map[string]interface{})
			tables = append(tables, TablePostgres{
				name:  t["name"].(string),
				alias: t["alias"].(string),
			})
		}
		builder.Tables(tables)
	}

	if v, ok := d.GetOk("textColumns"); ok {
		builder.TextColumns(v.([]string))
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "source"); err != nil {
		return diag.FromErr(err)
	}
	return SourceRead(ctx, d, meta)
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

	return SourceRead(ctx, d, meta)
}

func sourcePostgresDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	sourceName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := newSourcePostgresBuilder(sourceName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "source"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
