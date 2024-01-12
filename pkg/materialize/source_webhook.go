package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type FieldStruct struct {
	Body    bool
	Headers bool
	Secret  IdentifierSchemaStruct
}

type HeaderStruct struct {
	Header string
	Alias  string
	Bytes  bool
}

type IncludeHeadersStruct struct {
	All  bool
	Only []string
	Not  []string
}

type CheckOptionsStruct struct {
	Field FieldStruct
	Alias string
	Bytes bool
}

type SourceWebhookBuilder struct {
	Source
	clusterName     string
	bodyFormat      string
	includeHeader   []HeaderStruct
	includeHeaders  IncludeHeadersStruct
	checkOptions    []CheckOptionsStruct
	checkExpression string
}

func NewSourceWebhookBuilder(conn *sqlx.DB, obj MaterializeObject) *SourceWebhookBuilder {
	b := Builder{conn, BaseSource}
	return &SourceWebhookBuilder{
		Source: Source{b, obj.Name, obj.SchemaName, obj.DatabaseName},
	}
}

func (b *SourceWebhookBuilder) ClusterName(c string) *SourceWebhookBuilder {
	b.clusterName = c
	return b
}

func (b *SourceWebhookBuilder) BodyFormat(f string) *SourceWebhookBuilder {
	b.bodyFormat = f
	return b
}

func (b *SourceWebhookBuilder) IncludeHeader(h []HeaderStruct) *SourceWebhookBuilder {
	b.includeHeader = h
	return b
}

func (b *SourceWebhookBuilder) IncludeHeaders(h IncludeHeadersStruct) *SourceWebhookBuilder {
	b.includeHeaders = h
	return b
}

func (b *SourceWebhookBuilder) CheckOptions(o []CheckOptionsStruct) *SourceWebhookBuilder {
	b.checkOptions = o
	return b
}

func (b *SourceWebhookBuilder) CheckExpression(e string) *SourceWebhookBuilder {
	b.checkExpression = e
	return b
}

func (b *SourceWebhookBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s`, b.QualifiedName()))
	q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, QuoteIdentifier(b.clusterName)))
	q.WriteString(` FROM WEBHOOK`)
	q.WriteString(fmt.Sprintf(` BODY FORMAT %s`, b.bodyFormat))

	if len(b.includeHeader) > 0 {
		for _, h := range b.includeHeader {
			q.WriteString(fmt.Sprintf(` INCLUDE HEADER %s AS %s`, QuoteString(h.Header), h.Alias))
			if h.Bytes {
				q.WriteString(` BYTES`)
			}
		}
	}

	if b.includeHeaders.All || len(b.includeHeaders.Only) > 0 || len(b.includeHeaders.Not) > 0 {
		q.WriteString(` INCLUDE HEADERS`)

		var headers []string
		for _, h := range b.includeHeaders.Only {
			headers = append(headers, QuoteString(h))
		}
		for _, h := range b.includeHeaders.Not {
			headers = append(headers, fmt.Sprintf("NOT %s", QuoteString(h)))
		}
		if len(headers) > 0 {
			q.WriteString(fmt.Sprintf(` (%s)`, strings.Join(headers, ", ")))
		}
	}

	if len(b.checkOptions) > 0 || b.checkExpression != "" {
		var options []string
		for _, option := range b.checkOptions {
			var o string
			if option.Field.Body {
				o = "BODY"
			}
			if option.Field.Headers {
				o = "HEADERS"
			}
			if option.Field.Secret.Name != "" {
				o = "SECRET " + option.Field.Secret.QualifiedName()
			}
			if option.Alias != "" {
				o += fmt.Sprintf(" AS %s", option.Alias)
			}
			if option.Bytes {
				o += " BYTES"
			}
			options = append(options, o)
		}

		q.WriteString(" CHECK (")
		if len(options) > 0 {
			q.WriteString(fmt.Sprintf(" WITH (%s)", strings.Join(options, ", ")))
		}
		if b.checkExpression != "" {
			q.WriteString(fmt.Sprintf(" %s", b.checkExpression))
		}
		q.WriteString(")")
	}

	return b.ddl.exec(q.String())
}
