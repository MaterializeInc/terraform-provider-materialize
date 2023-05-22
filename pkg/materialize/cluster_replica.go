package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type ClusterReplicaBuilder struct {
	conn                       *sqlx.DB
	replicaName                string
	clusterName                string
	size                       string
	availabilityZone           string
	introspectionInterval      string
	introspectionDebugging     bool
	idleArrangementMergeEffort int
}

func NewClusterReplicaBuilder(conn *sqlx.DB, replicaName, clusterName string) *ClusterReplicaBuilder {
	return &ClusterReplicaBuilder{
		conn:        conn,
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

func (b *ClusterReplicaBuilder) Create() error {
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

	_, err := b.conn.Exec(q.String())
	if err != nil {
		return err
	}

	return nil
}

func (b *ClusterReplicaBuilder) Drop() error {
	q := fmt.Sprintf(`DROP CLUSTER REPLICA %s;`, b.QualifiedName())

	_, err := b.conn.Exec(q)
	if err != nil {
		return err
	}

	return nil
}

type ClusterReplicaParams struct {
	ReplicaId        sql.NullString `db:"id"`
	ReplicaName      sql.NullString `db:"replica_name"`
	ClusterName      sql.NullString `db:"cluster_name"`
	Size             sql.NullString `db:"size"`
	AvailabilityZone sql.NullString `db:"availability_zone"`
}

var clusterReplicaQuery = `
	SELECT
		mz_cluster_replicas.id,
		mz_cluster_replicas.name AS replica_name,
		mz_clusters.name AS cluster_name,
		mz_cluster_replicas.size,
		mz_cluster_replicas.availability_zone
	FROM mz_cluster_replicas
	JOIN mz_clusters
		ON mz_cluster_replicas.cluster_id = mz_clusters.id
`

func (b *ClusterReplicaBuilder) Id() (string, error) {
	q := NewBaseQuery(clusterReplicaQuery)
	p := map[string]string{
		"mz_cluster_replicas.name": b.replicaName,
		"mz_clusters.name":         b.clusterName,
	}

	var s ClusterReplicaParams
	if err := b.conn.Get(&s, q.queryPredicate(p)); err != nil {
		return "", err
	}

	return s.ReplicaId.String, nil
}

func (b *ClusterReplicaBuilder) Params(id string) (ClusterReplicaParams, error) {
	q := NewBaseQuery(clusterReplicaQuery)
	p := map[string]string{
		"mz_cluster_replicas.id": id,
	}

	var s ClusterReplicaParams
	if err := b.conn.Get(&s, q.queryPredicate(p)); err != nil {
		return s, err
	}

	return s, nil
}

func (b *ClusterReplicaBuilder) DataSource() ([]ClusterReplicaParams, error) {
	q := NewBaseQuery(clusterReplicaQuery)
	p := map[string]string{}

	var s []ClusterReplicaParams
	if err := b.conn.Select(&s, q.queryPredicate(p)); err != nil {
		return s, err
	}

	return s, nil
}
