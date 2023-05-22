package materialize

import (
	"fmt"
	"strings"
)

type ClusterReplicaBuilder struct {
	replicaName                string
	clusterName                string
	size                       string
	availabilityZone           string
	introspectionInterval      string
	introspectionDebugging     bool
	idleArrangementMergeEffort int
}

func NewClusterReplicaBuilder(replicaName, clusterName string) *ClusterReplicaBuilder {
	return &ClusterReplicaBuilder{
		replicaName: replicaName,
		clusterName: clusterName,
	}
}

func (b *ClusterReplicaBuilder) QualifiedName() string {
	return QualifiedName(b.clusterName, b.replicaName)
}

func (b *ClusterReplicaBuilder) Size(s string) *ClusterReplicaBuilder {
	b.size = s
	return b
}

func (b *ClusterReplicaBuilder) AvailabilityZone(z string) *ClusterReplicaBuilder {
	b.availabilityZone = z
	return b
}

func (b *ClusterReplicaBuilder) IntrospectionInterval(i string) *ClusterReplicaBuilder {
	b.introspectionInterval = i
	return b
}

func (b *ClusterReplicaBuilder) IntrospectionDebugging() *ClusterReplicaBuilder {
	b.introspectionDebugging = true
	return b
}

func (b *ClusterReplicaBuilder) IdleArrangementMergeEffort(e int) *ClusterReplicaBuilder {
	b.idleArrangementMergeEffort = e
	return b
}

func (b *ClusterReplicaBuilder) Create() string {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`CREATE CLUSTER REPLICA %s`, b.QualifiedName()))

	var p []string
	if b.size != "" {
		s := fmt.Sprintf(` SIZE = %s`, QuoteString(b.size))
		p = append(p, s)
	}

	if b.availabilityZone != "" {
		a := fmt.Sprintf(` AVAILABILITY ZONE = %s`, QuoteString(b.availabilityZone))
		p = append(p, a)
	}

	if b.introspectionInterval != "" {
		i := fmt.Sprintf(` INTROSPECTION INTERVAL = %s`, QuoteString(b.introspectionInterval))
		p = append(p, i)
	}

	if b.introspectionDebugging {
		p = append(p, ` INTROSPECTION DEBUGGING = TRUE`)
	}

	if b.idleArrangementMergeEffort != 0 {
		m := fmt.Sprintf(` IDLE ARRANGEMENT MERGE EFFORT = %d`, b.idleArrangementMergeEffort)
		p = append(p, m)
	}

	if len(p) > 0 {
		p := strings.Join(p[:], ",")
		q.WriteString(p)
	}

	q.WriteString(`;`)
	return q.String()
}

func (b *ClusterReplicaBuilder) Drop() string {
	return fmt.Sprintf(`DROP CLUSTER REPLICA %s;`, b.QualifiedName())
}

func (b *ClusterReplicaBuilder) ReadId() string {
	return fmt.Sprintf(`
		SELECT mz_cluster_replicas.id
		FROM mz_cluster_replicas
		JOIN mz_clusters
			ON mz_cluster_replicas.cluster_id = mz_clusters.id
		WHERE mz_cluster_replicas.name = %s
		AND mz_clusters.name = %s;`, QuoteString(b.replicaName), QuoteString(b.clusterName))
}

func ReadClusterReplicaParams(id string) string {
	return fmt.Sprintf(`
		SELECT
			mz_cluster_replicas.name AS replica_name,
			mz_clusters.name AS cluster_name,
			mz_cluster_replicas.size,
			mz_cluster_replicas.availability_zone
		FROM mz_cluster_replicas
		JOIN mz_clusters
			ON mz_cluster_replicas.cluster_id = mz_clusters.id
		WHERE mz_cluster_replicas.id = %s;`, QuoteString(id))
}

func ReadClusterReplicaDatasource() string {
	return `
		SELECT
			mz_cluster_replicas.id,
			mz_cluster_replicas.name,
			mz_clusters.name,
			mz_cluster_replicas.size,
			mz_cluster_replicas.availability_zone
		FROM mz_cluster_replicas
		JOIN mz_clusters
			ON mz_cluster_replicas.cluster_id = mz_clusters.id;
	`
}
