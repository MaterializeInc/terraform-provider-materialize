package materialize

import "github.com/jmoiron/sqlx"

// Any Materialize Database Object. Will contain name and optionally database and schema
// Cluster name only applies to cluster replicas
type MaterializeObject struct {
	ObjectType   EntityType
	Name         string
	SchemaName   string
	DatabaseName string
	ClusterName  string
}

func GetMaterializeObject(v interface{}) MaterializeObject {
	var conn MaterializeObject
	u := v.([]interface{})[0].(map[string]interface{})
	if v, ok := u["name"]; ok {
		conn.Name = v.(string)
	}
	if v, ok := u["schema_name"]; ok && v.(string) != "" {
		conn.SchemaName = v.(string)
	}

	if v, ok := u["database_name"]; ok && v.(string) != "" {
		conn.DatabaseName = v.(string)
	}

	if v, ok := u["cluster_name"]; ok && v.(string) != "" {
		conn.DatabaseName = v.(string)
	}

	if v, ok := u["object_type"]; ok && v.(string) != "" {
		conn.DatabaseName = v.(string)
	}
	return conn
}

func (g *MaterializeObject) QualifiedName() string {
	fields := []string{}

	if g.ClusterName != "" {
		fields = append(fields, g.ClusterName)
	} else {
		if g.DatabaseName != "" {
			fields = append(fields, g.DatabaseName)
		}

		if g.SchemaName != "" {
			fields = append(fields, g.SchemaName)
		}
	}

	fields = append(fields, g.Name)
	return QualifiedName(fields...)
}

func ObjectId(conn *sqlx.DB, object MaterializeObject) (string, error) {
	var i string
	var e error

	switch object.ObjectType {
	case Database:
		i, e = DatabaseId(conn, object)

	case Schema:
		i, e = SchemaId(conn, object)

	case Table:
		i, e = TableId(conn, object)

	case View:
		i, e = ViewId(conn, object)

	case MaterializedView:
		i, e = MaterializedViewId(conn, object)

	case BaseType:
		i, e = TypeId(conn, object)

	case BaseSource:
		i, e = SourceId(conn, object)

	case BaseConnection:
		i, e = ConnectionId(conn, object)

	case Secret:
		i, e = SecretId(conn, object)

	case Cluster:
		i, e = ClusterId(conn, object)

	case NetworkPolicy:
		i, e = NetworkPolicyId(conn, object)
	}

	if e != nil {
		return "", e
	}

	return i, nil
}
