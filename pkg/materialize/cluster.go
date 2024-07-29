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

func (b *ClusterBuilder) Create() error {
	q := strings.Builder{}

	q.WriteString(fmt.Sprintf(`CREATE CLUSTER %s`, b.QualifiedName()))
	// Only create empty clusters, manage replicas with separate resource if size is not set
	if b.size != "" {
		q.WriteString(fmt.Sprintf(` SIZE %s`, QuoteString(b.size)))

		var p []string

		if b.disk {
			i := fmt.Sprintf(` DISK`)
			p = append(p, i)
		}

		if b.replicationFactor != nil {
			i := fmt.Sprintf(` REPLICATION FACTOR %v`, *b.replicationFactor)
			p = append(p, i)
		}

		if len(b.availabilityZones) > 0 {
			var az []string
			for _, z := range b.availabilityZones {
				f := QuoteString(z)
				az = append(az, f)
			}
			a := fmt.Sprintf(` AVAILABILITY ZONES = [%s]`, strings.Join(az[:], ","))
			p = append(p, a)
		}

		if b.introspectionInterval != "" {
			i := fmt.Sprintf(` INTROSPECTION INTERVAL = %s`, QuoteString(b.introspectionInterval))
			p = append(p, i)
		}

		if b.introspectionDebugging {
			p = append(p, ` INTROSPECTION DEBUGGING = TRUE`)
		}

		if b.schedulingConfig.OnRefresh.Enabled {
			scheduleStatement := " SCHEDULE = ON REFRESH"
			if b.schedulingConfig.OnRefresh.HydrationTimeEstimate != "" {
				scheduleStatement += fmt.Sprintf(" (HYDRATION TIME ESTIMATE = %s)", QuoteString(b.schedulingConfig.OnRefresh.HydrationTimeEstimate))
			} else if b.schedulingConfig.OnRefresh.RehydrationTimeEstimate != "" {
				scheduleStatement += fmt.Sprintf(" (HYDRATION TIME ESTIMATE = %s)", QuoteString(b.schedulingConfig.OnRefresh.RehydrationTimeEstimate))
			}
			p = append(p, scheduleStatement)
		}

		if len(p) > 0 {
			p := strings.Join(p[:], ",")
			q.WriteString(fmt.Sprintf(`,%s`, p))
		}
	} else {
		q.WriteString(` REPLICAS ()`)
	}

	q.WriteString(`;`)

	return b.ddl.exec(q.String())
}

func (b *ClusterBuilder) Drop() error {
	qn := b.QualifiedName()
	return b.ddl.drop(qn)
}

func (b *ClusterBuilder) Resize(newSize string) error {
	q := fmt.Sprintf(`ALTER CLUSTER %s SET (SIZE '%s');`, b.QualifiedName(), newSize)
	return b.ddl.exec(q)
}

func (b *ClusterBuilder) SetDisk(disk bool) error {
	q := fmt.Sprintf(`ALTER CLUSTER %s SET (DISK %t);`, b.QualifiedName(), disk)
	return b.ddl.exec(q)
}

func (b *ClusterBuilder) SetReplicationFactor(newReplicationFactor int) error {
	q := fmt.Sprintf(`ALTER CLUSTER %s SET (REPLICATION FACTOR %d);`, b.QualifiedName(), newReplicationFactor)
	return b.ddl.exec(q)
}

func (b *ClusterBuilder) SetAvailabilityZones(availabilityZones []string) error {
	az := strings.Join(availabilityZones[:], ",")
	q := fmt.Sprintf(`ALTER CLUSTER %s SET (AVAILABILITY ZONES = [%s]);`, b.QualifiedName(), az)
	return b.ddl.exec(q)
}

func (b *ClusterBuilder) SetIntrospectionInterval(introspectionInterval string) error {
	q := fmt.Sprintf(`ALTER CLUSTER %s SET (INTROSPECTION INTERVAL %s);`, b.QualifiedName(), QuoteString(introspectionInterval))
	return b.ddl.exec(q)
}

func (b *ClusterBuilder) SetIntrospectionDebugging(introspectionDebugging bool) error {
	q := fmt.Sprintf(`ALTER CLUSTER %s SET (INTROSPECTION DEBUGGING %t);`, b.QualifiedName(), introspectionDebugging)
	return b.ddl.exec(q)
}

func (b *ClusterBuilder) SetSchedulingConfig(s interface{}) error {
	schedulingConfig := GetSchedulingConfig(s)
	var scheduleStatement string
	var q string

	// Check if the scheduling is enabled and set the appropriate SQL command.
	if schedulingConfig.OnRefresh.Enabled {
		scheduleStatement = "SCHEDULE = ON REFRESH"
		if schedulingConfig.OnRefresh.HydrationTimeEstimate != "" {
			scheduleStatement += fmt.Sprintf(" (HYDRATION TIME ESTIMATE = %s)", QuoteString(schedulingConfig.OnRefresh.HydrationTimeEstimate))
		} else if schedulingConfig.OnRefresh.RehydrationTimeEstimate != "" {
			scheduleStatement += fmt.Sprintf(" (HYDRATION TIME ESTIMATE = %s)", QuoteString(schedulingConfig.OnRefresh.RehydrationTimeEstimate))
		}
		q = fmt.Sprintf("ALTER CLUSTER %s SET (%s);", b.QualifiedName(), scheduleStatement)
	} else {
		// Reset the schedule settings if not enabled.
		q = fmt.Sprintf("ALTER CLUSTER %s RESET (SCHEDULE);", b.QualifiedName())
	}

	return b.ddl.exec(q)
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
