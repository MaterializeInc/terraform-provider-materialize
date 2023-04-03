package materialize

import (
	"fmt"
	"strings"
)

type TableLoadgen struct {
	Name  string
	Alias string
}

type CounterOptions struct {
	TickInterval   string
	MaxCardinality bool
}

type AuctionOptions struct {
	TickInterval string
	ScaleFactor  float64
	Table        []TableLoadgen
}

type TPCHOptions struct {
	TickInterval string
	ScaleFactor  float64
	Table        []TableLoadgen
}

type SourceLoadgenBuilder struct {
	Source
	clusterName       string
	size              string
	loadGeneratorType string
	counterOptions    CounterOptions
	auctionOptions    AuctionOptions
	tpchOptions       TPCHOptions
}

func NewSourceLoadgenBuilder(sourceName, schemaName, databaseName string) *SourceLoadgenBuilder {
	return &SourceLoadgenBuilder{
		Source: Source{sourceName, schemaName, databaseName},
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

func (b *SourceLoadgenBuilder) CounterOptions(c CounterOptions) *SourceLoadgenBuilder {
	b.counterOptions = c
	return b
}

func (b *SourceLoadgenBuilder) AuctionOptions(a AuctionOptions) *SourceLoadgenBuilder {
	b.auctionOptions = a
	return b
}

func (b *SourceLoadgenBuilder) TPCHOptions(t TPCHOptions) *SourceLoadgenBuilder {
	b.tpchOptions = t
	return b
}

func (b *SourceLoadgenBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s`, b.QualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, QuoteIdentifier(b.clusterName)))
	}

	q.WriteString(fmt.Sprintf(` FROM LOAD GENERATOR %s`, b.loadGeneratorType))

	// Optional Parameters
	var p []string

	if b.loadGeneratorType == "COUNTER" {
		if b.counterOptions.TickInterval != "" {
			t := fmt.Sprintf(`TICK INTERVAL %s`, QuoteString(b.counterOptions.TickInterval))
			p = append(p, t)
		}

		if b.counterOptions.MaxCardinality {
			p = append(p, `MAX CARDINALITY`)
		}
	} else if b.loadGeneratorType == "AUCTION" {
		if b.auctionOptions.TickInterval != "" {
			t := fmt.Sprintf(`TICK INTERVAL %s`, QuoteString(b.auctionOptions.TickInterval))
			p = append(p, t)
		}

		if b.auctionOptions.ScaleFactor != 0 {
			s := fmt.Sprintf(`SCALE FACTOR %.2f`, b.auctionOptions.ScaleFactor)
			p = append(p, s)
		}
	} else if b.loadGeneratorType == "TPCH" {
		if b.tpchOptions.TickInterval != "" {
			t := fmt.Sprintf(`TICK INTERVAL %s`, QuoteString(b.tpchOptions.TickInterval))
			p = append(p, t)
		}

		if b.tpchOptions.ScaleFactor != 0 {
			s := fmt.Sprintf(`SCALE FACTOR %.2f`, b.tpchOptions.ScaleFactor)
			p = append(p, s)
		}
	} else {
		panic("Not valid load generator type")
	}

	if len(p) != 0 {
		p := strings.Join(p[:], ", ")
		q.WriteString(fmt.Sprintf(` (%s)`, p))
	}

	// Table Mapping
	if b.loadGeneratorType == "COUNTER" {
		// Tables do not apply to COUNTER
	} else if len(b.auctionOptions.Table) > 0 || len(b.tpchOptions.Table) > 0 {
		var tables []string
		for _, t := range append(b.auctionOptions.Table, b.tpchOptions.Table...) {
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

	// Size
	if b.size != "" {
		q.WriteString(fmt.Sprintf(` WITH (SIZE = %s)`, QuoteString(b.size)))
	}

	q.WriteString(`;`)
	return q.String()
}

func (b *SourceLoadgenBuilder) Rename(newName string) string {
	n := QualifiedName(b.DatabaseName, b.SchemaName, newName)
	return fmt.Sprintf(`ALTER SOURCE %s RENAME TO %s;`, b.QualifiedName(), n)
}

func (b *SourceLoadgenBuilder) UpdateSize(newSize string) string {
	return fmt.Sprintf(`ALTER SOURCE %s SET (SIZE = %s);`, b.QualifiedName(), QuoteString(newSize))
}

func (b *SourceLoadgenBuilder) Drop() string {
	return fmt.Sprintf(`DROP SOURCE %s;`, b.QualifiedName())
}

func (b *SourceLoadgenBuilder) ReadId() string {
	return ReadSourceId(b.SourceName, b.SchemaName, b.DatabaseName)
}
