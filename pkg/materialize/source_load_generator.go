package materialize

import (
	"fmt"
	"strings"
)

type TableLoadgen struct {
	Name  string
	Alias string
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
	table             []TableLoadgen
}

func (b *SourceLoadgenBuilder) qualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.sourceName)
}

func NewSourceLoadgenBuilder(sourceName, schemaName, databaseName string) *SourceLoadgenBuilder {
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

func (b *SourceLoadgenBuilder) Table(t []TableLoadgen) *SourceLoadgenBuilder {
	b.table = t
	return b
}

func (b *SourceLoadgenBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s`, b.qualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, QuoteIdentifier(b.clusterName)))
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
	} else if len(b.table) > 0 {

		var tables []string
		for _, t := range b.table {
			if t.Alias == "" {
				t.Alias = t.Name
			}
			s := fmt.Sprintf(`%s AS %s`, t.Name, t.Alias)
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
	return ReadSourceId(b.sourceName, b.schemaName, b.databaseName)
}
