package resources

var loadGeneratorTypes = []string{
	"AUCTION",
	"MARKETING",
	"TPCH",
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

var sqlServerSSLMode = []string{
	"disabled",
	"required",
	"verify",
	"verify-ca",
}

var sinkFormatCompatibilityLevels = []string{
	"BACKWARD",
	"BACKWARD_TRANSITIVE",
	"FORWARD",
	"FORWARD_TRANSITIVE",
	"FULL",
	"FULL_TRANSITIVE",
	"NONE",
}
