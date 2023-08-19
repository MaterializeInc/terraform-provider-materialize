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

type CheckOptionsStruct struct {
	Field FieldStruct
	Alias string
}

type SourceWebhookBuilder struct {
	Source
	clusterName     string
	size            string // TODO: size is not supported for webhook sources
	bodyFormat      string
	includeHeaders  bool
	checkOptions    []CheckOptionsStruct
	checkExpression string
}

func NewSourceWebhookBuilder(conn *sqlx.DB, obj ObjectSchemaStruct) *SourceWebhookBuilder {
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

func (b *SourceWebhookBuilder) IncludeHeaders(h bool) *SourceWebhookBuilder {
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

func (b *SourceWebhookBuilder) Size(s string) *SourceWebhookBuilder {
	b.size = s
	return b
}

func (b *SourceWebhookBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s`, b.QualifiedName()))
	q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, QuoteIdentifier(b.clusterName)))
	q.WriteString(` FROM WEBHOOK`)
	q.WriteString(fmt.Sprintf(` BODY FORMAT %s`, b.bodyFormat))

	if b.includeHeaders {
		q.WriteString(` INCLUDE HEADERS`)
	}

	if len(b.checkOptions) > 0 || b.checkExpression != "" {
		q.WriteString(` CHECK (`)
		if len(b.checkOptions) > 0 {
			q.WriteString(` WITH (`)
			var options []string
			for _, option := range b.checkOptions {
				if option.Field.Body {
					options = append(options, "BODY")
				}
				if option.Field.Headers {
					options = append(options, "HEADERS")
				}
				if option.Field.Secret.Name != "" {
					options = append(options, "SECRET "+option.Field.Secret.QualifiedName())
				}
				if option.Alias != "" {
					options[len(options)-1] += " AS " + option.Alias
				}
			}
			q.WriteString(strings.Join(options, ", "))
			q.WriteString(`) `)
		}
		if b.checkExpression != "" {
			q.WriteString(b.checkExpression)
		}
		q.WriteString(`)`)
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}
