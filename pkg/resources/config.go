package resources

var loadGeneratorTypes = []string{
	"AUCTION",
	"MARKETING",
	"COUNTER",
	"TPCH",
}

var localSizes = []string{
	"1",
	"2",
	"2-1",
	"2-2",
	"2-4",
	"4",
	"4-1",
	"4-4",
	"8",
	"8-1",
	"8-8",
	"16",
	"16-1",
	"16-16",
	"32",
	"32-1",
	"32-32",
}

// https://materialize.com/docs/sql/create-cluster-replica/#sizes
var replicaSizes = []string{
	"3xsmall",
	"2xsmall",
	"xsmall",
	"small",
	"medium",
	"large",
	"xlarge",
	"x2large",
	"x3large",
	"x4large",
	"x5large",
	"x6large",
}

var saslMechanisms = []string{
	"PLAIN",
	"SCRAM-SHA-256",
	"SCRAM-SHA-512",
}

var sourceSizes = []string{
	"3xsmall",
	"2xsmall",
	"xsmall",
	"small",
	"medium",
	"large",
	"xlarge",
}

var strategy = []string{
	"INLINE",
	"ID",
	"LATEST",
}

var aliases = map[string]string{
	"int8":    "bigint",
	"bool":    "boolean",
	"float":   "double precision",
	"float8":  "double precision",
	"double":  "double precision",
	"int":     "integer",
	"int4":    "integer",
	"json":    "jsonb",
	"decimal": "numeric",
	"real":    "float4",
	"int2":    "smallint",
	"uint":    "uint4",
}
