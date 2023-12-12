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
