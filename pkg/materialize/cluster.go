package materialize

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// DDL
type ClusterBuilder struct {
	ddl               Builder
	clusterName       string
	replicationFactor int
	size              string
}

func NewClusterBuilder(conn *sqlx.DB, obj ObjectSchemaStruct) *ClusterBuilder {
	return &ClusterBuilder{
		ddl:         Builder{conn, Cluster},
		clusterName: obj.Name,
	}
}

func (b *ClusterBuilder) QualifiedName() string {
	return QualifiedName(b.clusterName)
}

func (b *ClusterBuilder) ReplicationFactor(r int) *ClusterBuilder {
	b.replicationFactor = r
	return b
}

func (b *ClusterBuilder) Size(s string) *ClusterBuilder {
	b.size = s
	return b
}

func (b *ClusterBuilder) Create() error {
	// Only create empty clusters, manage replicas with separate resource if replication factor is not set
	if b.replicationFactor == 0 {
		q := fmt.Sprintf(`CREATE CLUSTER %s REPLICAS ();`, b.QualifiedName())
		return b.ddl.exec(q)
	}
	q := fmt.Sprintf(`CREATE CLUSTER %s SIZE '%s', REPLICATION FACTOR %d;`, b.QualifiedName(), b.size, b.replicationFactor)
	return b.ddl.exec(q)
}

func (b *ClusterBuilder) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

func (b *ClusterBuilder) Resize(newSize string) error {
	q := fmt.Sprintf(`ALTER CLUSTER %s SET (SIZE '%s');`, b.QualifiedName(), newSize)
	return b.ddl.exec(q)
}

func (b *ClusterBuilder) ResizeReplicationFactor(newReplicationFactor int) error {
	q := fmt.Sprintf(`ALTER CLUSTER %s SET (REPLICATION FACTOR %d);`, b.QualifiedName(), newReplicationFactor)
	return b.ddl.exec(q)
}

// DML
type ClusterParams struct {
	ClusterId         sql.NullString `db:"id"`
	ClusterName       sql.NullString `db:"name"`
	OwnerName         sql.NullString `db:"owner_name"`
	Privileges        sql.NullString `db:"privileges"`
	ReplicationFactor sql.NullInt64  `db:"replication_factor"`
	Size              sql.NullString `db:"size"`
	Managed           sql.NullBool   `db:"managed"`
}

var clusterQuery = NewBaseQuery(`
	SELECT
		mz_clusters.id,
		mz_clusters.name,
		mz_clusters.replication_factor,
		mz_clusters.size,
		mz_clusters.managed,
		mz_roles.name AS owner_name,
		mz_clusters.privileges
	FROM mz_clusters
	JOIN mz_roles
		ON mz_clusters.owner_id = mz_roles.id`)

func ClusterId(conn *sqlx.DB, obj ObjectSchemaStruct) (string, error) {
	q := clusterQuery.QueryPredicate(map[string]string{"mz_clusters.name": obj.Name})

	var c ClusterParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return c.ClusterId.String, nil
}

func ScanCluster(conn *sqlx.DB, id string) (ClusterParams, error) {
	q := clusterQuery.QueryPredicate(map[string]string{"mz_clusters.id": id})

	var c ClusterParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}

func ListClusters(conn *sqlx.DB) ([]ClusterParams, error) {
	q := clusterQuery.QueryPredicate(map[string]string{})

	var c []ClusterParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
