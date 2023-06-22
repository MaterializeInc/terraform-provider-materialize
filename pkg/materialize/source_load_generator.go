package materialize

import (
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type CounterOptions struct {
	TickInterval   string
	ScaleFactor    float64
	MaxCardinality int
}

func GetCounterOptionsStruct(v interface{}) CounterOptions {
	var o CounterOptions
	u := v.([]interface{})[0].(map[string]interface{})
	if v, ok := u["tick_interval"]; ok {
		o.TickInterval = v.(string)
	}

	if v, ok := u["scale_factor"]; ok {
		o.ScaleFactor = v.(float64)
	}

	if v, ok := u["max_cardinality"]; ok {
		o.MaxCardinality = v.(int)
	}
	return o
}

type AuctionOptions struct {
	TickInterval string
	ScaleFactor  float64
	Table        []TableStruct
}

func GetAuctionOptionsStruct(v interface{}) AuctionOptions {
	var o AuctionOptions
	u := v.([]interface{})[0].(map[string]interface{})
	if v, ok := u["tick_interval"]; ok {
		o.TickInterval = v.(string)
	}

	if v, ok := u["scale_factor"]; ok {
		o.ScaleFactor = v.(float64)
	}

	if v, ok := u["table"]; ok {
		o.Table = GetTableStruct(v.([]interface{}))
	}
	return o
}

type MarketingOptions struct {
	TickInterval string
	ScaleFactor  float64
	Table        []TableStruct
}

func GetMarketingOptionsStruct(v interface{}) MarketingOptions {
	var o MarketingOptions
	u := v.([]interface{})[0].(map[string]interface{})
	if v, ok := u["tick_interval"]; ok {
		o.TickInterval = v.(string)
	}

	if v, ok := u["scale_factor"]; ok {
		o.ScaleFactor = v.(float64)
	}

	if v, ok := u["table"]; ok {
		o.Table = GetTableStruct(v.([]interface{}))
	}
	return o
}

type TPCHOptions struct {
	TickInterval string
	ScaleFactor  float64
	Table        []TableStruct
}

func GetTPCHOptionsStruct(v interface{}) TPCHOptions {
	var o TPCHOptions
	u := v.([]interface{})[0].(map[string]interface{})
	if v, ok := u["tick_interval"]; ok {
		o.TickInterval = v.(string)
	}

	if v, ok := u["scale_factor"]; ok {
		o.ScaleFactor = v.(float64)
	}

	if v, ok := u["table"]; ok {
		o.Table = GetTableStruct(v.([]interface{}))
	}
	return o
}

type SourceLoadgenBuilder struct {
	Source
	clusterName       string
	size              string
	loadGeneratorType string
	counterOptions    CounterOptions
	auctionOptions    AuctionOptions
	marketingOptions  MarketingOptions
	tpchOptions       TPCHOptions
}

func NewSourceLoadgenBuilder(conn *sqlx.DB, obj ObjectSchemaStruct) *SourceLoadgenBuilder {
	b := Builder{conn, BaseSource}
	return &SourceLoadgenBuilder{
		Source: Source{b, obj.Name, obj.SchemaName, obj.DatabaseName},
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

func (b *SourceLoadgenBuilder) MarketingOptions(m MarketingOptions) *SourceLoadgenBuilder {
	b.marketingOptions = m
	return b
}

func (b *SourceLoadgenBuilder) TPCHOptions(t TPCHOptions) *SourceLoadgenBuilder {
	b.tpchOptions = t
	return b
}

func (b *SourceLoadgenBuilder) Create() error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE SOURCE %s`, b.QualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, QuoteIdentifier(b.clusterName)))
	}

	q.WriteString(fmt.Sprintf(` FROM LOAD GENERATOR %s`, b.loadGeneratorType))

	// Optional Parameters
	var p []string

	for _, t := range []string{b.counterOptions.TickInterval, b.auctionOptions.TickInterval, b.marketingOptions.TickInterval, b.tpchOptions.TickInterval} {
		if t != "" {
			p = append(p, fmt.Sprintf(`TICK INTERVAL %s`, QuoteString(t)))
		}
	}

	for _, t := range []float64{b.counterOptions.ScaleFactor, b.auctionOptions.ScaleFactor, b.marketingOptions.ScaleFactor, b.tpchOptions.ScaleFactor} {
		if t != 0 {
			p = append(p, fmt.Sprintf(`SCALE FACTOR %.2f`, t))
		}
	}

	if b.counterOptions.MaxCardinality != 0 {
		s := fmt.Sprintf(`MAX CARDINALITY %d`, b.counterOptions.MaxCardinality)
		p = append(p, s)
	}

	if len(p) != 0 {
		p := strings.Join(p[:], ", ")
		q.WriteString(fmt.Sprintf(` (%s)`, p))
	}

	// Table Mapping
	if b.loadGeneratorType == "COUNTER" {
		// Tables do not apply to COUNTER
	} else if len(b.auctionOptions.Table) > 0 || len(b.marketingOptions.Table) > 0 || len(b.tpchOptions.Table) > 0 {

		var ot []TableStruct
		if len(b.auctionOptions.Table) > 0 {
			ot = b.auctionOptions.Table
		} else if len(b.marketingOptions.Table) > 0 {
			ot = b.marketingOptions.Table
		} else {
			ot = b.tpchOptions.Table
		}

		var tables []string
		for _, t := range ot {
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
	return b.ddl.exec(q.String())
}
