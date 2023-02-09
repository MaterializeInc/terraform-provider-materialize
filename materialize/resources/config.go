package resources

var connectionTypes = []string{
	"KAFKA",
	"POSTGRES",
	"LOAD GENERATOR",
	"AWS PRIVATELINK"
	"SSH TUNNEL",
	"CONFLUENT SCHEMA REGISTRY",
}

var envelopes = []string{
	"DEBEZIUM",
	"UPSERT",
	"TPCH",
}

var loadGeneratorTypes = []string{
	"AUCTION",
	"COUNTER",
	"NONE",
}

var regions = []string{
	"us-east-1",
	"eu-west-1",
}

var replicaSizes = []string{
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

var sourceSizes = []string{
	"3xsmall",
	"2xsmall",
	"xsmall",
	"small",
	"medium",
	"large",
	"xlarge",
}
