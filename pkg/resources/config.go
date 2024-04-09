package resources

var loadGeneratorTypes = []string{
	"AUCTION",
	"MARKETING",
	"COUNTER",
	"TPCH",
	"KEY VALUE",
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
	"2xlarge",
	"3xlarge",
	"4xlarge",
	"5xlarge",
	"6xlarge",
	"25cc",
	"50cc",
	"100cc",
	"200cc",
	"300cc",
	"400cc",
	"600cc",
	"800cc",
	"1200cc",
	"1600cc",
	"3200cc",
	"6400cc",
	"128C",
	"256C",
	"512C",
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

var securityProtocols = []string{
	"PLAINTEXT",
	"SASL_PLAINTEXT",
	"SSL",
	"SASL_SSL",
}

var compressionTypes = []string{
	"none",
	"gzip",
	"snappy",
	"lz4",
	"ztsd",
}

var ssoConfigTypes = []string{
	"saml",
	"oidc",
}

var scim2ConfigSources = []string{
	"okta",
	"azure-ad",
	"other",
}

var mysqlSSLMode = []string{
	"disabled",
	"required",
	"verify-ca",
	"verify-identity",
}
