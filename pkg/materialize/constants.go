package materialize

var Permissions = map[string]string{
	"r": "SELECT",
	"a": "INSERT",
	"w": "UPDATE",
	"d": "DELETE",
	"C": "CREATE",
	"U": "USAGE",
}

type ObjectType struct {
	Permissions  []string
	CatalogTable string
}

// https://materialize.com/docs/sql/grant-privilege/#details
var ObjectPermissions = map[string]ObjectType{
	"DATABASE": {
		Permissions:  []string{"U", "C"},
		CatalogTable: "mz_databases",
	},
	"SCHEMA": {
		Permissions:  []string{"U", "C"},
		CatalogTable: "mz_schemas",
	},
	"TABLE": {
		Permissions:  []string{"a", "r", "w", "d"},
		CatalogTable: "mz_tables",
	},
	"VIEW": {
		Permissions:  []string{"r"},
		CatalogTable: "mz_views",
	},
	"MATERIALIZED VIEW": {
		Permissions:  []string{"r"},
		CatalogTable: "mz_materialized_views",
	},
	"INDEX": {
		Permissions:  []string{},
		CatalogTable: "mz_indexes",
	},
	"TYPE": {
		Permissions:  []string{"U"},
		CatalogTable: "mz_types",
	},
	"SOURCE": {
		Permissions:  []string{"r"},
		CatalogTable: "mz_sources",
	},
	"SINK": {
		Permissions:  []string{},
		CatalogTable: "mz_sinks",
	},
	"CONNECTION": {
		Permissions:  []string{"U"},
		CatalogTable: "mz_connections",
	},
	"SECRET": {
		Permissions:  []string{"U"},
		CatalogTable: "mz_secrets",
	},
	"CLUSTER": {
		Permissions:  []string{"U", "C"},
		CatalogTable: "mz_clusters",
	},
}

var TableMapping = []string{"SOURCE", "VIEW", "MATERIALIZED VIEW"}
