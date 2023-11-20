package resources

var loadGeneratorTypes = []string{
	"AUCTION",
	"MARKETING",
	"COUNTER",
	"TPCH",
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

var sessionVariables = []string{
	"cluster",
	"cluster_replica",
	"database",
	"search_path",
	"transaction_isolation",
	"auto_route_introspection_queries",
	"application_name",
	"client_encoding",
	"client_min_messages",
	"datestyle",
	"emit_introspection_query_notice",
	"emit_timestamp_notice",
	"emit_trace_id_notice",
	"enable_session_rbac_checks",
	"extra_float_digits",
	"failpoints",
	"idle_in_transaction_session_timeout",
	"integer_datetimes",
	"intervalstyle",
	"is_superuser",
	"max_identifier_length",
	"max_query_result_size",
	"mz_version",
	"server_version",
	"server_version_num",
	"sql_safe_updates",
	"standard_conforming_strings",
	"statement_timeout",
	"timezone",
}

var securityProtocols = []string{
	"PLAINTEXT",
	"SASL_PLAINTEXT",
	"SSL",
	"SASL_SSL",
}
