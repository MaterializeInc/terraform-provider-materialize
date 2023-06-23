package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
)

type MaterializedViewBuilder struct {
	ddl                  Builder
	materializedViewName string
	schemaName           string
	databaseName         string
	clusterName          string
	selectStmt           string
}

func NewMaterializedViewBuilder(conn *sqlx.DB, obj ObjectSchemaStruct) *MaterializedViewBuilder {
	return &MaterializedViewBuilder{
		ddl:                  Builder{conn, MaterializedView},
		materializedViewName: obj.Name,
		schemaName:           obj.SchemaName,
		databaseName:         obj.DatabaseName,
	}
}

func (b *MaterializedViewBuilder) QualifiedName() string {
	return QualifiedName(b.databaseName, b.schemaName, b.materializedViewName)
}

func (b *MaterializedViewBuilder) ClusterName(clusterName string) *MaterializedViewBuilder {
	b.clusterName = clusterName
	return b
}

func (b *MaterializedViewBuilder) SelectStmt(selectStmt string) *MaterializedViewBuilder {
	b.selectStmt = selectStmt
	return b
}

func (b *MaterializedViewBuilder) Create() error {
	q := strings.Builder{}

	q.WriteString(fmt.Sprintf(`CREATE MATERIALIZED VIEW %s`, b.QualifiedName()))

	if b.clusterName != "" {
		q.WriteString(fmt.Sprintf(` IN CLUSTER %s`, QuoteIdentifier(b.clusterName)))
	}

	q.WriteString(fmt.Sprintf(` AS %s;`, b.selectStmt))
	return b.ddl.exec(q.String())
}

func (b *MaterializedViewBuilder) Rename(newMaterializedViewName string) error {
	old := b.QualifiedName()
	new := QualifiedName(newMaterializedViewName)
	return b.ddl.rename(old, new)
}

func (b *MaterializedViewBuilder) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

type MaterializedViewParams struct {
	MaterializedViewId   sql.NullString `db:"id"`
	MaterializedViewName sql.NullString `db:"materialized_view_name"`
	SchemaName           sql.NullString `db:"schema_name"`
	DatabaseName         sql.NullString `db:"database_name"`
	Cluster              sql.NullString `db:"cluster_name"`
	OwnerName            sql.NullString `db:"owner_name"`
	Privileges           sql.NullString `db:"privileges"`
}

var materializedViewQuery = NewBaseQuery(`
	SELECT
		mz_materialized_views.id,
		mz_materialized_views.name AS materialized_view_name,
		mz_schemas.name AS schema_name,
		mz_databases.name AS database_name,
		mz_clusters.name AS cluster_name,
		mz_roles.name AS owner_name,
		mz_materialized_views.privileges
	FROM mz_materialized_views
	JOIN mz_schemas
		ON mz_materialized_views.schema_id = mz_schemas.id
	JOIN mz_databases
		ON mz_schemas.database_id = mz_databases.id
	LEFT JOIN mz_clusters
		ON mz_materialized_views.cluster_id = mz_clusters.id
	JOIN mz_roles
		ON mz_materialized_views.owner_id = mz_roles.id`)

func MaterializedViewId(conn *sqlx.DB, obj ObjectSchemaStruct) (string, error) {
	p := map[string]string{
		"mz_materialized_views.name": obj.Name,
		"mz_schemas.name":            obj.SchemaName,
		"mz_databases.name":          obj.DatabaseName,
	}
	q := materializedViewQuery.QueryPredicate(p)

	var c MaterializedViewParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return c.MaterializedViewId.String, nil
}

func ScanMaterializedView(conn *sqlx.DB, id string) (MaterializedViewParams, error) {
	p := map[string]string{
		"mz_materialized_views.id": id,
	}
	q := materializedViewQuery.QueryPredicate(p)

	var c MaterializedViewParams
	if err := conn.Get(&c, q); err != nil {
		return c, err
	}

	return c, nil
}

func ListMaterializedViews(conn *sqlx.DB, schemaName, databaseName string) ([]MaterializedViewParams, error) {
	p := map[string]string{
		"mz_schemas.name":   schemaName,
		"mz_databases.name": databaseName,
	}
	q := materializedViewQuery.QueryPredicate(p)

	var c []MaterializedViewParams
	if err := conn.Select(&c, q); err != nil {
		return c, err
	}

	return c, nil
}
