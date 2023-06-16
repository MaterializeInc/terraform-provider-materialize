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
	Permissions []string
}

// https://materialize.com/docs/sql/grant-privilege/#details
var ObjectPermissions = map[string]ObjectType{
	"DATABASE": {
		Permissions: []string{"U", "C"},
	},
	"SCHEMA": {
		Permissions: []string{"U", "C"},
	},
	"TABLE": {
		Permissions: []string{"a", "r", "w", "d"},
	},
	"VIEW": {
		Permissions: []string{"r"},
	},
	"MATERIALIZED VIEW": {
		Permissions: []string{"r"},
	},
	"INDEX": {
		Permissions: []string{},
	},
	"TYPE": {
		Permissions: []string{"U"},
	},
	"SOURCE": {
		Permissions: []string{"r"},
	},
	"SINK": {
		Permissions: []string{},
	},
	"CONNECTION": {
		Permissions: []string{"U"},
	},
	"SECRET": {
		Permissions: []string{"U"},
	},
	"CLUSTER": {
		Permissions: []string{"U", "C"},
	},
}

var TableMapping = []string{"SOURCE", "VIEW", "MATERIALIZED VIEW"}
