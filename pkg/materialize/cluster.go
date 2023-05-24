package materialize

import (
	"database/sql"
	"fmt"

	"github.com/jmoiron/sqlx"
)

// DDL
type ClusterBuilder struct {
	ddl         Builder
	clusterName string
}

func NewClusterBuilder(conn *sqlx.DB, clusterName string) *ClusterBuilder {
	return &ClusterBuilder{
		ddl:         Builder{conn, Cluster},
		clusterName: clusterName,
	}
}

func (b *ClusterBuilder) QualifiedName() string {
	return QualifiedName(b.clusterName)
}

func (b *ClusterBuilder) Create() error {
	// Only create empty clusters, manage replicas with separate resource
	q := fmt.Sprintf(`CREATE CLUSTER %s REPLICAS ();`, b.QualifiedName())
	return b.ddl.exec(q)
}

func (b *ClusterBuilder) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

// DML
type ClusterParams struct {
	ClusterId   sql.NullString `db:"id"`
	ClusterName sql.NullString `db:"name"`
}

var clusterQuery = NewBaseQuery(`SELECT id, name FROM mz_clusters`)

func ClusterId(conn *sqlx.DB, clusterName string) (string, error) {
	q := clusterQuery.QueryPredicate(map[string]string{"name": clusterName})

	var c ClusterParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return c.ClusterId.String, nil
}

func ScanCluster(conn *sqlx.DB, id string) (ClusterParams, error) {
	q := clusterQuery.QueryPredicate(map[string]string{"id": id})

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
