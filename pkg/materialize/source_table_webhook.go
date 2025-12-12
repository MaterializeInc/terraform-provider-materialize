package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

// SourceTableWebhookParams contains the parameters for a webhook source table
type SourceTableWebhookParams struct {
	SourceTableParams
}

// Query to get webhook source table information
var sourceTableWebhookQuery = NewBaseQuery(`
	SELECT
		mz_tables.id,
		mz_tables.name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		comments.comment AS comment,
		mz_roles.name AS owner_name,
		mz_tables.privileges
	FROM mz_tables
	JOIN mz_schemas
		ON mz_tables.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	JOIN mz_roles
		ON mz_tables.owner_id = mz_roles.id
	LEFT JOIN (
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'table'
		AND object_sub_id IS NULL
	) comments
		ON mz_tables.id = comments.id`)

// SourceTableWebhookId retrieves the ID of a webhook source table
func SourceTableWebhookId(conn *sqlx.DB, obj MaterializeObject) (string, error) {
	p := map[string]string{
		"mz_tables.name":    obj.Name,
		"mz_schemas.name":   obj.SchemaName,
		"mz_databases.name": obj.DatabaseName,
	}
	q := sourceTableWebhookQuery.QueryPredicate(p)

	var t SourceTableParams
	if err := conn.Get(&t, q); err != nil {
		return "", err
	}

	return t.TableId.String, nil
}

// ScanSourceTableWebhook scans a webhook source table by ID
func ScanSourceTableWebhook(conn *sqlx.DB, id string) (SourceTableWebhookParams, error) {
	q := sourceTableWebhookQuery.QueryPredicate(map[string]string{"mz_tables.id": id})

	var params SourceTableWebhookParams
	if err := conn.Get(&params, q); err != nil {
		return params, err
	}

	return params, nil
}

// SourceTableWebhookBuilder builds webhook source tables
type SourceTableWebhookBuilder struct {
	ddl             Builder
	tableName       string
	schemaName      string
	databaseName    string
	bodyFormat      string
	includeHeader   []HeaderStruct
	includeHeaders  IncludeHeadersStruct
	checkOptions    []CheckOptionsStruct
	checkExpression string
}

// NewSourceTableWebhookBuilder creates a new webhook source table builder
func NewSourceTableWebhookBuilder(conn *sqlx.DB, obj MaterializeObject) *SourceTableWebhookBuilder {
	return &SourceTableWebhookBuilder{
		ddl:          Builder{conn, Table},
		tableName:    obj.Name,
		schemaName:   obj.SchemaName,
		databaseName: obj.DatabaseName,
	}
}

// QualifiedName returns the fully qualified name of the table
func (b *SourceTableWebhookBuilder) QualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.tableName)
}

// BodyFormat sets the body format
func (b *SourceTableWebhookBuilder) BodyFormat(f string) *SourceTableWebhookBuilder {
	b.bodyFormat = f
	return b
}

// IncludeHeader adds header inclusions
func (b *SourceTableWebhookBuilder) IncludeHeader(h []HeaderStruct) *SourceTableWebhookBuilder {
	b.includeHeader = h
	return b
}

// IncludeHeaders sets headers to include
func (b *SourceTableWebhookBuilder) IncludeHeaders(h IncludeHeadersStruct) *SourceTableWebhookBuilder {
	b.includeHeaders = h
	return b
}

// CheckOptions sets the check options
func (b *SourceTableWebhookBuilder) CheckOptions(o []CheckOptionsStruct) *SourceTableWebhookBuilder {
	b.checkOptions = o
	return b
}

// CheckExpression sets the check expression
func (b *SourceTableWebhookBuilder) CheckExpression(e string) *SourceTableWebhookBuilder {
	b.checkExpression = e
	return b
}

// Drop removes the webhook source table
func (b *SourceTableWebhookBuilder) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

func (b *SourceTableWebhookBuilder) Rename(newName string) error {
	oldName := b.QualifiedName()
	b.tableName = newName
	newName = b.QualifiedName()
	return b.ddl.rename(oldName, newName)
}

// Create creates the webhook source table
func (b *SourceTableWebhookBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE TABLE %s FROM WEBHOOK`, b.QualifiedName()))

	// Add webhook-specific options
	var options []string

	// Body Format
	options = append(options, fmt.Sprintf(`BODY FORMAT %s`, b.bodyFormat))

	// Include Header
	if len(b.includeHeader) > 0 {
		for _, h := range b.includeHeader {
			headerOption := fmt.Sprintf(`INCLUDE HEADER %s`, QuoteString(h.Header))
			if h.Alias != "" {
				headerOption += fmt.Sprintf(` AS %s`, h.Alias)
			}
			if h.Bytes {
				headerOption += ` BYTES`
			}
			options = append(options, headerOption)
		}
	}

	// Include Headers
	if b.includeHeaders.All || len(b.includeHeaders.Only) > 0 || len(b.includeHeaders.Not) > 0 {
		headerOption := `INCLUDE HEADERS`
		var headers []string

		for _, h := range b.includeHeaders.Only {
			headers = append(headers, QuoteString(h))
		}
		for _, h := range b.includeHeaders.Not {
			headers = append(headers, fmt.Sprintf("NOT %s", QuoteString(h)))
		}

		if len(headers) > 0 {
			headerOption += fmt.Sprintf(` (%s)`, strings.Join(headers, ", "))
		}
		options = append(options, headerOption)
	}

	// Check Options and Expression
	// check_expression is required when check_options are provided
	if b.checkExpression != "" {
		var checkOpts []string
		for _, opt := range b.checkOptions {
			var o string
			if opt.Field.Body {
				o = "BODY"
			}
			if opt.Field.Headers {
				o = "HEADERS"
			}
			if opt.Field.Secret.Name != "" {
				o = "SECRET " + opt.Field.Secret.QualifiedName()
			}
			if opt.Alias != "" {
				o += fmt.Sprintf(" AS %s", opt.Alias)
			}
			if opt.Bytes {
				o += " BYTES"
			}
			checkOpts = append(checkOpts, o)
		}

		checkOption := "CHECK ("
		if len(checkOpts) > 0 {
			checkOption += fmt.Sprintf(" WITH (%s)", strings.Join(checkOpts, ", "))
		}
		if len(checkOpts) > 0 {
			checkOption += " "
		}
		checkOption += b.checkExpression
		checkOption += ")"
		options = append(options, checkOption)
	}

	if len(options) > 0 {
		q.WriteString(" ")
		q.WriteString(strings.Join(options, " "))
	}

	q.WriteString(";")
	return b.ddl.exec(q.String())
}
