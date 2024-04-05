package materialize

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/jmoiron/sqlx"
)

type CounterOptions struct {
	TickInterval   string
	MaxCardinality int
}

func GetCounterOptionsStruct(v interface{}) CounterOptions {
	var o CounterOptions
	u := v.([]interface{})[0].(map[string]interface{})
	if v, ok := u["tick_interval"]; ok {
		o.TickInterval = v.(string)
	}

	if v, ok := u["max_cardinality"]; ok {
		o.MaxCardinality = v.(int)
	}
	return o
}

type AuctionOptions struct {
	TickInterval string
}

func GetAuctionOptionsStruct(v interface{}) AuctionOptions {
	var o AuctionOptions
	u := v.([]interface{})[0].(map[string]interface{})
	if v, ok := u["tick_interval"]; ok {
		o.TickInterval = v.(string)
	}

	return o
}

type MarketingOptions struct {
	TickInterval string
}

func GetMarketingOptionsStruct(v interface{}) MarketingOptions {
	var o MarketingOptions
	u := v.([]interface{})[0].(map[string]interface{})
	if v, ok := u["tick_interval"]; ok {
		o.TickInterval = v.(string)
	}

	return o
}

type TPCHOptions struct {
	TickInterval string
	ScaleFactor  float64
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
	return o
}

type KeyValueOptions struct {
	Keys                  int
	SnapshotRounds        int
	TransactionalSnapshot bool
	ValueSize             int
	TickInterval          string
	Seed                  uint8
	Partitions            int
	BatchSize             int
}

func GetKeyValueOptionsStruct(v interface{}) KeyValueOptions {
	var o KeyValueOptions
	u := v.([]interface{})[0].(map[string]interface{})

	if val, ok := u["keys"]; ok {
		o.Keys = val.(int)
	}

	if val, ok := u["snapshot_rounds"]; ok {
		o.SnapshotRounds = val.(int)
	}

	if val, ok := u["transactional_snapshot"]; ok {
		o.TransactionalSnapshot = val.(bool)
	}

	if val, ok := u["value_size"]; ok {
		o.ValueSize = val.(int)
	}

	if val, ok := u["tick_interval"]; ok {
		o.TickInterval = val.(string)
	}

	if val, ok := u["seed"]; ok {
		o.Seed = uint8(val.(int))
	}

	if val, ok := u["partitions"]; ok {
		o.Partitions = val.(int)
	}

	if val, ok := u["batch_size"]; ok {
		o.BatchSize = val.(int)
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
	keyValueOptions   KeyValueOptions
	exposeProgress    IdentifierSchemaStruct
}

func NewSourceLoadgenBuilder(conn *sqlx.DB, obj MaterializeObject) *SourceLoadgenBuilder {
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

func (b *SourceLoadgenBuilder) ExposeProgress(e IdentifierSchemaStruct) *SourceLoadgenBuilder {
	b.exposeProgress = e
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

func (b *SourceLoadgenBuilder) KeyValueOptions(k KeyValueOptions) *SourceLoadgenBuilder {
	b.keyValueOptions = k
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

	for _, t := range []float64{b.tpchOptions.ScaleFactor} {
		if t != 0 {
			p = append(p, fmt.Sprintf(`SCALE FACTOR %.2f`, t))
		}
	}

	if b.loadGeneratorType == "KEY_VALUE" {
		// Add KEY VALUE specific parameters
		if b.keyValueOptions.Keys != 0 {
			p = append(p, fmt.Sprintf(`KEYS %d`, b.keyValueOptions.Keys))
		}
		if b.keyValueOptions.SnapshotRounds != 0 {
			p = append(p, fmt.Sprintf(`SNAPSHOT ROUNDS %d`, b.keyValueOptions.SnapshotRounds))
		}
		if b.keyValueOptions.TransactionalSnapshot {
			p = append(p, fmt.Sprintf(`TRANSACTIONAL SNAPSHOT %s`, strconv.FormatBool(b.keyValueOptions.TransactionalSnapshot)))
		}
		if b.keyValueOptions.ValueSize != 0 {
			p = append(p, fmt.Sprintf(`VALUE SIZE %d`, b.keyValueOptions.ValueSize))
		}
		if b.keyValueOptions.TickInterval != "" {
			p = append(p, fmt.Sprintf(`TICK INTERVAL %s`, QuoteString(b.keyValueOptions.TickInterval)))
		}
		if b.keyValueOptions.Seed != 0 {
			p = append(p, fmt.Sprintf(`SEED %d`, b.keyValueOptions.Seed))
		}
		if b.keyValueOptions.Partitions != 0 {
			p = append(p, fmt.Sprintf(`PARTITIONS %d`, b.keyValueOptions.Partitions))
		}
		if b.keyValueOptions.BatchSize != 0 {
			p = append(p, fmt.Sprintf(`BATCH SIZE %d`, b.keyValueOptions.BatchSize))
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

	// Include for multi-output sources
	if b.loadGeneratorType == "AUCTION" || b.loadGeneratorType == "MARKETING" || b.loadGeneratorType == "TPCH" {
		q.WriteString(` FOR ALL TABLES`)
	}

	if b.exposeProgress.Name != "" {
		q.WriteString(fmt.Sprintf(` EXPOSE PROGRESS AS %s`, b.exposeProgress.QualifiedName()))
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}
