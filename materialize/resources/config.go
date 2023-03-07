package resources

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
