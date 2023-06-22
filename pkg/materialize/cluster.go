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

func NewClusterBuilder(conn *sqlx.DB, obj ObjectSchemaStruct) *ClusterBuilder {
	return &ClusterBuilder{
		ddl:         Builder{conn, Cluster},
		clusterName: obj.Name,
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
	OwnerName   sql.NullString `db:"owner_name"`
	Privileges  sql.NullString `db:"privileges"`
}

var clusterQuery = NewBaseQuery(`
	SELECT
		mz_clusters.id,
		mz_clusters.name,
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
