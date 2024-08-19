package materialize

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/jmoiron/sqlx"
	"github.com/lib/pq"
)

// DDL
type ClusterBuilder struct {
	ddl                    Builder
	clusterName            string
	replicationFactor      *int
	size                   string
	disk                   bool
	availabilityZones      []string
	introspectionInterval  string
	introspectionDebugging bool
	schedulingConfig       SchedulingConfig
}

func NewClusterBuilder(conn *sqlx.DB, obj MaterializeObject) *ClusterBuilder {
	return &ClusterBuilder{
		ddl:         Builder{conn, Cluster},
		clusterName: obj.Name,
	}
}

type SchedulingConfig struct {
	OnRefresh OnRefreshConfig
	IsSet     bool
}

type OnRefreshConfig struct {
	Enabled                 bool
	HydrationTimeEstimate   string
	RehydrationTimeEstimate string
}

func GetSchedulingConfig(v interface{}) SchedulingConfig {
	if v == nil {
		return SchedulingConfig{}
	}

	configSlice, ok := v.([]interface{})
	if !ok || len(configSlice) == 0 {
		return SchedulingConfig{}
	}

	configMap, ok := configSlice[0].(map[string]interface{})
	if !ok {
		return SchedulingConfig{}
	}

	onRefreshSlice, ok := configMap["on_refresh"].([]interface{})
	if !ok || len(onRefreshSlice) == 0 {
		return SchedulingConfig{}
	}

	onRefreshMap, ok := onRefreshSlice[0].(map[string]interface{})
	if !ok {
		return SchedulingConfig{}
	}

	config := SchedulingConfig{
		OnRefresh: OnRefreshConfig{
			Enabled: false,
		},
		IsSet: true,
	}

	if enabled, ok := onRefreshMap["enabled"].(bool); ok {
		config.OnRefresh.Enabled = enabled
	}

	if hydrationTimeEstimate, ok := onRefreshMap["hydration_time_estimate"].(string); ok {
		config.OnRefresh.HydrationTimeEstimate = hydrationTimeEstimate
	}

	if rehydrationTimeEstimate, ok := onRefreshMap["rehydration_time_estimate"].(string); ok {
		config.OnRefresh.RehydrationTimeEstimate = rehydrationTimeEstimate
	}

	return config
}

func (b *ClusterBuilder) QualifiedName() string {
	return QualifiedName(b.clusterName)
}

func (b *ClusterBuilder) ReplicationFactor(r *int) *ClusterBuilder {
	b.replicationFactor = r
	return b
}

func (b *ClusterBuilder) Size(s string) *ClusterBuilder {
	b.size = s
	return b
}

func (b *ClusterBuilder) Disk(disk bool) *ClusterBuilder {
	b.disk = disk
	return b
}

func (b *ClusterBuilder) AvailabilityZones(z []string) *ClusterBuilder {
	b.availabilityZones = z
	return b
}

func (b *ClusterBuilder) IntrospectionInterval(i string) *ClusterBuilder {
	b.introspectionInterval = i
	return b
}

func (b *ClusterBuilder) IntrospectionDebugging() *ClusterBuilder {
	b.introspectionDebugging = true
	return b
}

func (b *ClusterBuilder) Scheduling(v []interface{}) *ClusterBuilder {
	b.schedulingConfig = GetSchedulingConfig(v)
	return b
}

func (b *ClusterBuilder) GenerateClusterOptions() string {
	var p []string
	if b.size != "" {
		i := fmt.Sprintf(`SIZE %s`, QuoteString(b.size))
		p = append(p, i)
	}

	if b.disk {
		i := fmt.Sprintf(`DISK`)
		p = append(p, i)
	}

	if b.replicationFactor != nil {
		i := fmt.Sprintf(`REPLICATION FACTOR %v`, *b.replicationFactor)
		p = append(p, i)
	}

	if len(b.availabilityZones) > 0 {
		var az []string
		for _, z := range b.availabilityZones {
			f := QuoteString(z)
			az = append(az, f)
		}
		a := fmt.Sprintf(`AVAILABILITY ZONES = [%s]`, strings.Join(az[:], ","))
		p = append(p, a)
	}

	if b.introspectionInterval != "" {
		i := fmt.Sprintf(`INTROSPECTION INTERVAL = %s`, QuoteString(b.introspectionInterval))
		p = append(p, i)
	}

	if b.introspectionDebugging {
		p = append(p, `INTROSPECTION DEBUGGING = TRUE`)
	}

	if b.schedulingConfig.IsSet {
		// We skip setting this in alter cluster because it
		// will be handled independently
		p = append(p, b.GenSchedulingConfigSql(b.schedulingConfig))
	}

	if len(p) > 0 {
		return strings.Join(p[:], ", ")
	}
	return ""
}

func (b *ClusterBuilder) Create() error {
	q := strings.Builder{}

	q.WriteString(fmt.Sprintf(`CREATE CLUSTER %s`, b.QualifiedName()))
	// Size is the only required option for managed clusters. A "REPLICAS ()" option indicates
	// and is required for unmanaged clusters. No other options can be provided for managed
	// clusters.
	if b.size != "" {
		q.WriteString(fmt.Sprintf(` (%s)`, b.GenerateClusterOptions()))
	} else {
		q.WriteString(` (REPLICAS ())`)
	}

	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}

func (b *ClusterBuilder) AlterClusterScheduling(s SchedulingConfig) error {
	q := strings.Builder{}
	q.WriteString(fmt.Sprintf(`ALTER CLUSTER %s`, b.QualifiedName()))
	q.WriteString(fmt.Sprintf(` SET (%s)`, b.GenSchedulingConfigSql(s)))
	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}

func (b *ClusterBuilder) AlterCluster() error {
	q := strings.Builder{}

	q.WriteString(fmt.Sprintf(`ALTER CLUSTER %s`, b.QualifiedName()))
	// The only alterations to unmanaged clusters should be to
	// move them to maanged clusters, we will assume here that we are only
	// dealing with managed clusters
	q.WriteString(fmt.Sprintf(` SET (%s)`, b.GenerateClusterOptions()))
	q.WriteString(`;`)
	return b.ddl.exec(q.String())
}

func (b *ClusterBuilder) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

func (b *ClusterBuilder) SetSize(newSize string) {
	b.size = newSize
}

func (b *ClusterBuilder) SetDisk(disk bool) {
	b.disk = disk
}

func (b *ClusterBuilder) SetReplicationFactor(newReplicationFactor int) {
	b.replicationFactor = &newReplicationFactor
}

func (b *ClusterBuilder) SetAvailabilityZones(availabilityZones []string) {
	b.availabilityZones = availabilityZones
}

func (b *ClusterBuilder) SetIntrospectionInterval(introspectionInterval string) {
	b.introspectionInterval = introspectionInterval
}

func (b *ClusterBuilder) SetIntrospectionDebugging(introspectionDebugging bool) {
	b.introspectionDebugging = introspectionDebugging
}

func (b *ClusterBuilder) SetSchedulingConfig(s interface{}) {
	b.schedulingConfig = GetSchedulingConfig(s)
}

func (b *ClusterBuilder) GenSchedulingConfigSql(s SchedulingConfig) string {
	statement := "SCHEDULE = "
	if !s.IsSet || !s.OnRefresh.Enabled {
		statement += "MANUAL"
		return statement
	} else {
		statement += "ON REFRESH"
		if s.OnRefresh.HydrationTimeEstimate != "" {
			statement += fmt.Sprintf(" (HYDRATION TIME ESTIMATE = %s)", QuoteString(b.schedulingConfig.OnRefresh.HydrationTimeEstimate))
		} else if s.OnRefresh.RehydrationTimeEstimate != "" {
			statement += fmt.Sprintf(" (HYDRATION TIME ESTIMATE = %s)", QuoteString(b.schedulingConfig.OnRefresh.RehydrationTimeEstimate))
		}
	}
	return statement
}

// DML
type ClusterParams struct {
	ClusterId         sql.NullString `db:"id"`
	ClusterName       sql.NullString `db:"name"`
	Managed           sql.NullBool   `db:"managed"`
	Size              sql.NullString `db:"size"`
	ReplicationFactor sql.NullInt64  `db:"replication_factor"`
	Disk              sql.NullBool   `db:"disk"`
	AvailabilityZones pq.StringArray `db:"availability_zones"`
	Comment           sql.NullString `db:"comment"`
	OwnerName         sql.NullString `db:"owner_name"`
	Privileges        pq.StringArray `db:"privileges"`
}

var clusterQuery = NewBaseQuery(`
	SELECT
		mz_clusters.id,
		mz_clusters.name,
		mz_clusters.managed,
		mz_clusters.size,
		mz_clusters.replication_factor,
		mz_clusters.disk,
		mz_clusters.availability_zones,
		comments.comment AS comment,
		mz_roles.name AS owner_name,
		mz_clusters.privileges
	FROM mz_clusters
	JOIN mz_roles
		ON mz_clusters.owner_id = mz_roles.id
	LEFT JOIN (
		SELECT id, comment
		FROM mz_internal.mz_comments
		WHERE object_type = 'cluster'
	) comments
		ON mz_clusters.id = comments.id`)

func ClusterId(conn *sqlx.DB, obj MaterializeObject) (string, error) {
	q := clusterQuery.QueryPredicate(map[string]string{"mz_clusters.name": obj.Name})

	var c ClusterParams
	if err := conn.Get(&c, q); err != nil {
		return "", err
	}

	return c.ClusterId.String, nil
}

func ScanCluster(conn *sqlx.DB, identifier string, byName bool) (ClusterParams, error) {
	var predicate map[string]string
	if byName {
		predicate = map[string]string{"mz_clusters.name": identifier}
	} else {
		predicate = map[string]string{"mz_clusters.id": identifier}
	}
	q := clusterQuery.QueryPredicate(predicate)

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
