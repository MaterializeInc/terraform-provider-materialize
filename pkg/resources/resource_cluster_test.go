package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

var inCluster = map[string]interface{}{
	"name":                    "cluster",
	"size":                    "3xsmall",
	"replication_factor":      2,
	"availability_zones":      []interface{}{"use1-az1", "use1-az2"},
	"introspection_interval":  "10s",
	"introspection_debugging": true,
	"scheduling": []interface{}{
		map[string]interface{}{
			"on_refresh": []interface{}{
				map[string]interface{}{
					"enabled":                 true,
					"hydration_time_estimate": "2 hours",
				},
			},
		},
	},
	"ownership_role": "joe",
}

func TestResourceClusterCreate(t *testing.T) {
	r := require.New(t)

	d := schema.TestResourceDataRaw(t, Cluster().Schema, inCluster)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`
			CREATE CLUSTER "cluster" \(SIZE '3xsmall',
			REPLICATION FACTOR 2,
			AVAILABILITY ZONES = \['use1-az1','use1-az2'\],
			INTROSPECTION INTERVAL = '10s',
			INTROSPECTION DEBUGGING = TRUE,
			SCHEDULE = ON REFRESH \(HYDRATION TIME ESTIMATE = '2 hours'\)\);
		`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Ownership
		mock.ExpectExec(`ALTER CLUSTER "cluster" OWNER TO "joe";`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_clusters.name = 'cluster'`
		testhelpers.MockClusterScan(mock, ip)

		// Query Params
		pp := `WHERE mz_clusters.id = 'u1'`
		testhelpers.MockClusterScan(mock, pp)
		testhelpers.MockClusterReconfigurationScan(mock, "", 0)
		testhelpers.MockClusterAutoScalingScan(mock, "", 0)

		if err := clusterCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceClusterAutoScalingCreate(t *testing.T) {
	r := require.New(t)

	in := map[string]interface{}{
		"name": "cluster",
		"size": "3xsmall",
		"auto_scaling_strategy": []interface{}{
			map[string]interface{}{
				"on_hydration": []interface{}{
					map[string]interface{}{
						"hydration_size":  "800cc",
						"linger_duration": "15s",
					},
				},
			},
		},
	}
	d := schema.TestResourceDataRaw(t, Cluster().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`
			CREATE CLUSTER "cluster" \(SIZE '3xsmall',
			INTROSPECTION INTERVAL = '1s',
			AUTO SCALING STRATEGY = \(ON HYDRATION \(HYDRATION SIZE = '800cc', LINGER DURATION = '15s'\)\)\);
		`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_clusters.name = 'cluster'`
		testhelpers.MockClusterScan(mock, ip)

		// Query Params
		pp := `WHERE mz_clusters.id = 'u1'`
		testhelpers.MockClusterScan(mock, pp)
		testhelpers.MockClusterReconfigurationScan(mock, "", 0)
		testhelpers.MockClusterAutoScalingScan(mock, "800cc", 15)

		if err := clusterCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		strategy := d.Get("auto_scaling_strategy").([]interface{})
		r.Len(strategy, 1)
	})
}

// Confirm id is updated with region and type prefix
func TestResourceClusterReadIdMigration(t *testing.T) {
	utils.SetDefaultRegion("aws/us-east-1")
	r := require.New(t)

	testCases := []struct {
		name           string
		identifyByName bool
		initialId      string
		expectedId     string
		mockId         string
		mockName       string
	}{
		{
			name:           "Migrate to ID-based identifier",
			identifyByName: false,
			initialId:      "u1",
			expectedId:     "aws/us-east-1:id:u1",
			mockId:         "u1",
			mockName:       "cluster",
		},
		{
			name:           "Migrate to name-based identifier",
			identifyByName: true,
			initialId:      "u1",
			expectedId:     "aws/us-east-1:name:cluster",
			mockId:         "u1",
			mockName:       "cluster",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			in := map[string]interface{}{
				"name":             "cluster",
				"identify_by_name": tc.identifyByName,
			}
			d := schema.TestResourceDataRaw(t, Cluster().Schema, in)
			r.NotNil(d)

			// Set id before migration
			d.SetId(tc.initialId)

			testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
				// Query Params
				pp := `WHERE mz_clusters.id = '` + tc.mockId + `'`
				testhelpers.MockClusterScan(mock, pp)
				testhelpers.MockClusterReconfigurationScan(mock, "", 0)
				testhelpers.MockClusterAutoScalingScan(mock, "", 0)

				if err := clusterRead(context.TODO(), d, db); err != nil {
					t.Fatal(err)
				}

				if d.Id() != tc.expectedId {
					t.Fatalf("unexpected id of %s, expected %s", d.Id(), tc.expectedId)
				}

				// Verify that the name is set correctly
				if name := d.Get("name").(string); name != tc.mockName {
					t.Fatalf("unexpected name of %s, expected %s", name, tc.mockName)
				}
			})
		})
	}
}

func TestResourceClusterZeroReplicationCreate(t *testing.T) {
	r := require.New(t)

	var inClusterZeroReplication = map[string]interface{}{
		"name":               "cluster",
		"size":               "3xsmall",
		"replication_factor": 0,
	}
	d := schema.TestResourceDataRaw(t, Cluster().Schema, inClusterZeroReplication)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(`
			CREATE CLUSTER "cluster" \(SIZE '3xsmall',
			REPLICATION FACTOR 0,
			INTROSPECTION INTERVAL = '1s'\);
		`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_clusters.name = 'cluster'`
		testhelpers.MockClusterScan(mock, ip)

		// Query Params
		pp := `WHERE mz_clusters.id = 'u1'`
		testhelpers.MockClusterScan(mock, pp)
		testhelpers.MockClusterReconfigurationScan(mock, "", 0)
		testhelpers.MockClusterAutoScalingScan(mock, "", 0)

		if err := clusterCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceClusterDelete(t *testing.T) {
	r := require.New(t)

	d := schema.TestResourceDataRaw(t, Cluster().Schema, inCluster)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP CLUSTER "cluster";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := clusterDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

// fakeChanges reports a fixed set of attributes as changed.
type fakeChanges map[string]bool

func (f fakeChanges) HasChange(key string) bool { return f[key] }

// Materialize rejects WAIT unless the same statement also changes something
// that builds new replicas, so we only send it for those attributes.
func TestWaitUntilReadySupported(t *testing.T) {
	tests := []struct {
		changed string
		want    bool
	}{
		{"size", true},
		{"availability_zones", true},
		{"introspection_interval", true},
		{"introspection_debugging", true},
		// The reported case: autoscaling resizes in place, nothing to wait on.
		{"auto_scaling_strategy", false},
		// Same restriction, and more likely to be hit in practice.
		{"replication_factor", false},
	}

	for _, tt := range tests {
		t.Run(tt.changed, func(t *testing.T) {
			require.Equal(t, tt.want, waitUntilReadySupported(fakeChanges{tt.changed: true}))
		})
	}

	t.Run("a supported change alongside an unsupported one still waits", func(t *testing.T) {
		require.True(t, waitUntilReadySupported(fakeChanges{"size": true, "replication_factor": true}))
	})

	t.Run("no changes", func(t *testing.T) {
		require.False(t, waitUntilReadySupported(fakeChanges{}))
	})
}
