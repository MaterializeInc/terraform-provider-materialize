package resources

import (
	"context"
	"testing"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/require"
)

// Confirm id is updated with region for 0.4.0
// All resources share the same read function
func TestResourceGrantPrivilegeReadIdMigration(t *testing.T) {
	utils.SetDefaultRegion("aws/us-east-1")
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":    "joe",
		"privilege":    "CREATE",
		"cluster_name": "materialize",
	}
	d := schema.TestResourceDataRaw(t, GrantCluster().Schema, in)
	r.NotNil(d)

	// Set id before migration
	d.SetId("GRANT|CLUSTER|u1|u1|CREATE")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_clusters.id = 'u1'`
		testhelpers.MockClusterScan(mock, pp)

		if err := grantRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:GRANT|CLUSTER|u1|u1|CREATE" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

// A cluster swap or rename drops the cluster the id in state refers to
func TestResourceGrantPrivilegeReadClusterSwapped(t *testing.T) {
	utils.SetDefaultRegion("aws/us-east-1")
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":    "joe",
		"privilege":    "USAGE",
		"cluster_name": "materialize",
	}
	d := schema.TestResourceDataRaw(t, GrantCluster().Schema, in)
	r.NotNil(d)

	d.SetId("aws/us-east-1:GRANT|CLUSTER|u99|u1|USAGE")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_clusters.id = 'u99'`
		testhelpers.MockClusterScanNoRows(mock, pp)

		// Query Cluster Id
		cp := `WHERE mz_clusters.name = 'materialize'`
		testhelpers.MockClusterScan(mock, cp)

		// Query Params
		np := `WHERE mz_clusters.id = 'u1'`
		testhelpers.MockClusterScan(mock, np)

		if err := grantRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:GRANT|CLUSTER|u1|u1|USAGE" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceGrantPrivilegeReadClusterDropped(t *testing.T) {
	utils.SetDefaultRegion("aws/us-east-1")
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":    "joe",
		"privilege":    "USAGE",
		"cluster_name": "materialize",
	}
	d := schema.TestResourceDataRaw(t, GrantCluster().Schema, in)
	r.NotNil(d)

	d.SetId("aws/us-east-1:GRANT|CLUSTER|u99|u1|USAGE")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_clusters.id = 'u99'`
		testhelpers.MockClusterScanNoRows(mock, pp)

		// Query Cluster Id
		cp := `WHERE mz_clusters.name = 'materialize'`
		testhelpers.MockClusterScanNoRows(mock, cp)

		if err := grantRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

// A cluster id imported into another grant resource has no cluster_name to
// re-resolve from, which must not take the provider down
func TestResourceGrantPrivilegeReadClusterIdOnOtherResource(t *testing.T) {
	utils.SetDefaultRegion("aws/us-east-1")
	r := require.New(t)

	d := schema.TestResourceDataRaw(t, GrantDatabase().Schema, map[string]interface{}{})
	r.NotNil(d)

	d.SetId("aws/us-east-1:GRANT|CLUSTER|u99|u1|USAGE")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		pp := `WHERE mz_clusters.id = 'u99'`
		testhelpers.MockClusterScanNoRows(mock, pp)

		if err := grantRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		r.Empty(d.Id(), "an unresolvable grant should be removed from state")
	})
}

// A grant that is no longer present on the object should leave state, otherwise
// the next plan sees no drift and the grant is never recreated
func TestResourceGrantPrivilegeReadRevoked(t *testing.T) {
	utils.SetDefaultRegion("aws/us-east-1")
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":    "joe",
		"privilege":    "USAGE",
		"cluster_name": "materialize",
	}
	d := schema.TestResourceDataRaw(t, GrantCluster().Schema, in)
	r.NotNil(d)

	// u99 holds no privileges on the cluster
	d.SetId("aws/us-east-1:GRANT|CLUSTER|u1|u99|USAGE")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		testhelpers.MockClusterScan(mock, `WHERE mz_clusters.id = 'u1'`)

		if err := grantRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		r.Empty(d.Id(), "a revoked grant should be removed from state")
	})
}

// The swapped-in cluster does not always carry the grant. Re-resolving must not
// hide that, or the grant is silently missing and no plan ever restores it
func TestResourceGrantPrivilegeReadClusterSwappedWithoutGrant(t *testing.T) {
	utils.SetDefaultRegion("aws/us-east-1")
	r := require.New(t)

	in := map[string]interface{}{
		"role_name":    "joe",
		"privilege":    "USAGE",
		"cluster_name": "materialize",
	}
	d := schema.TestResourceDataRaw(t, GrantCluster().Schema, in)
	r.NotNil(d)

	d.SetId("aws/us-east-1:GRANT|CLUSTER|u99|u77|USAGE")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// The id in state was dropped by the swap
		testhelpers.MockClusterScanNoRows(mock, `WHERE mz_clusters.id = 'u99'`)

		// The name resolves to the new cluster, which holds no grant for u77
		testhelpers.MockClusterScan(mock, `WHERE mz_clusters.name = 'materialize'`)
		testhelpers.MockClusterScan(mock, `WHERE mz_clusters.id = 'u1'`)

		if err := grantRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		r.Empty(d.Id(), "a grant absent from the swapped-in cluster should be removed from state")
	})
}
