package materialize

import (
	"fmt"
)

type ClusterBuilder struct {
	clusterName string
}

func NewClusterBuilder(clusterName string) *ClusterBuilder {
	return &ClusterBuilder{
		clusterName: clusterName,
	}
}

func (b *ClusterBuilder) Create() string {
	// Only create empty clusters, manage replicas with separate resource
	return fmt.Sprintf(`CREATE CLUSTER %s REPLICAS ();`, QuoteIdentifier(b.clusterName))
}

func (b *ClusterBuilder) Drop() string {
	return fmt.Sprintf(`DROP CLUSTER %s;`, QuoteIdentifier(b.clusterName))
}

func (b *ClusterBuilder) ReadId() string {
	return fmt.Sprintf(`SELECT id FROM mz_clusters WHERE name = %s;`, QuoteString(b.clusterName))
}

func ReadClusterParams(id string) string {
	return fmt.Sprintf("SELECT name AS cluster_name FROM mz_clusters WHERE id = %s;", QuoteString(id))
}

func ReadClusterDatasource() string {
	return `SELECT id, name FROM mz_clusters;`
}
