package resources

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/hashicorp/terraform-plugin-sdk/v2/diag"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/jmoiron/sqlx"
)

var tableSchema = map[string]*schema.Schema{
	"name": {
		Description: "The identifier for the table.",
		Type:        schema.TypeString,
		Required:    true,
	},
	"schema_name": {
		Description: "The identifier for the table schema.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "public",
		ForceNew:    true,
	},
	"database_name": {
		Description: "The identifier for the table database.",
		Type:        schema.TypeString,
		Optional:    true,
		Default:     "materialize",
		ForceNew:    true,
	},
	"qualified_name": {
		Description: "The fully qualified name of the table.",
		Type:        schema.TypeString,
		Computed:    true,
	},
	"columns": {
		Description: "Columns of the table.",
		Type:        schema.TypeList,
		Elem: &schema.Resource{
			Schema: map[string]*schema.Schema{
				"col_name": {
					Description: "The name of the column to be created in the table.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"col_type": {
					Description: "The data type of the column indicated by col_name.",
					Type:        schema.TypeString,
					Required:    true,
				},
				"not_null": {
					Description: "	Do not allow the column to contain NULL values. Columns without this constraint can contain NULL values.",
					Type:     schema.TypeBool,
					Optional: true,
				},
			},
		},
		Optional: true,
		MinItems: 1,
		ForceNew: true,
	},
}

func Table() *schema.Resource {
	return &schema.Resource{
		Description: "A table persists in durable storage and can be written to, updated and seamlessly joined with other tables, views or sources.",

		CreateContext: tableCreate,
		ReadContext:   tableRead,
		UpdateContext: tableUpdate,
		DeleteContext: tableDelete,

		Importer: &schema.ResourceImporter{
			StateContext: schema.ImportStatePassthroughContext,
		},

		Schema: tableSchema,
	}
}

type TableColumn struct {
	colName string
	colType string
	notNull bool
}

type TableBuilder struct {
	tableName    string
	schemaName   string
	databaseName string
	columns      []TableColumn
}

func (b *TableBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.tableName)
}

func newTableBuilder(tableName, schemaName, databaseName string) *TableBuilder {
	return &TableBuilder{
		tableName:    tableName,
		schemaName:   schemaName,
		databaseName: databaseName,
	}
}

func (b *TableBuilder) Columns(c []TableColumn) *TableBuilder {
	b.columns = c
	return b
}

func (b *TableBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(`CREATE`)

	q.WriteString(fmt.Sprintf(` TABLE %s`, b.qualifiedName()))

	var columns []string
	for _, c := range b.columns {
		s := strings.Builder{}

		s.WriteString(fmt.Sprintf(`%s %s`, c.colName, c.colType))
		if c.notNull {
			s.WriteString(` NOT NULL`)
		}
		o := s.String()
		columns = append(columns, o)

	}
	p := strings.Join(columns[:], ", ")
	q.WriteString(fmt.Sprintf(` (%s);`, p))
	return q.String()
}

func (b *TableBuilder) Rename(newName string) string {
	n := QualifiedName(b.databaseName, b.schemaName, newName)
	return fmt.Sprintf(`ALTER TABLE %s RENAME TO %s;`, b.qualifiedName(), n)
}

func (b *TableBuilder) Drop() string {
	return fmt.Sprintf(`DROP TABLE %s;`, b.qualifiedName())
}

func (b *TableBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_tables.id
		FROM mz_tables
		JOIN mz_schemas
			ON mz_tables.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_tables.name = %s
		AND mz_schemas.name = %s
		AND mz_databases.name = %s;
	`, QuoteString(b.tableName), QuoteString(b.schemaName), QuoteString(b.databaseName))
}

func readTableParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_tables.name,
			mz_schemas.name,
			mz_databases.name
		FROM mz_tables
		JOIN mz_schemas
			ON mz_tables.schema_id = mz_schemas.id
		JOIN mz_databases
			ON mz_schemas.database_id = mz_databases.id
		WHERE mz_tables.id = %s;`, QuoteString(id))
}

func tableRead(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	i := d.Id()
	q := readTableParams(i)

	var name, schema, database *string
	if err := conn.QueryRowx(q).Scan(&name, &schema, &database); err != nil {
		return diag.FromErr(err)
	}

	d.SetId(i)

	if err := d.Set("name", name); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("schema_name", schema); err != nil {
		return diag.FromErr(err)
	}

	if err := d.Set("database_name", database); err != nil {
		return diag.FromErr(err)
	}

	return nil
}

func tableCreate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)

	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	builder := newTableBuilder(tableName, schemaName, databaseName)

	if v, ok := d.GetOk("columns"); ok {
		var columns []TableColumn
		for _, column := range v.([]interface{}) {
			c := column.(map[string]interface{})
			columns = append(columns, TableColumn{
				colName: c["col_name"].(string),
				colType: c["col_type"].(string),
				notNull: c["not_null"].(bool),
			})
		}
		builder.Columns(columns)
	}

	qc := builder.Create()
	qr := builder.ReadId()

	if err := createResource(conn, d, qc, qr, "table"); err != nil {
		return diag.FromErr(err)
	}
	return tableRead(ctx, d, meta)
}

func tableUpdate(ctx context.Context, d *schema.ResourceData, meta interface{}) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	if d.HasChange("name") {
		_, newName := d.GetChange("name")

		q := newSecretBuilder(tableName, schemaName, databaseName).Rename(newName.(string))

		if err := ExecResource(conn, q); err != nil {
			log.Printf("[ERROR] could not rename table: %s", q)
			return diag.FromErr(err)
		}
	}

	return tableRead(ctx, d, meta)
}

func tableDelete(ctx context.Context, d *schema.ResourceData, meta any) diag.Diagnostics {
	conn := meta.(*sqlx.DB)
	tableName := d.Get("name").(string)
	schemaName := d.Get("schema_name").(string)
	databaseName := d.Get("database_name").(string)

	q := newTableBuilder(tableName, schemaName, databaseName).Drop()

	if err := dropResource(conn, d, q, "table"); err != nil {
		return diag.FromErr(err)
	}
	return nil
}
