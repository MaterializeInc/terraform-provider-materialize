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

var inSourceWebhook = map[string]interface{}{
	"name":          "webhook_source",
	"schema_name":   "schema",
	"database_name": "database",
	"cluster_name":  "cluster",
	"body_format":   "JSON",
	"include_headers": []interface{}{
		map[string]interface{}{
			"all": true,
		},
	},
	"check_options": []interface{}{
		map[string]interface{}{
			"field": []interface{}{map[string]interface{}{
				"body": true,
			}},
			"alias": "bytes",
		},
		map[string]interface{}{
			"field": []interface{}{map[string]interface{}{
				"headers": true,
			}},
			"alias": "headers",
		},
	},
	"check_expression": "check_expression",
}

func TestResourceSourceWebhookCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceWebhook().Schema, inSourceWebhook)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SOURCE "database"."schema"."webhook_source" IN CLUSTER "cluster" FROM WEBHOOK BODY FORMAT JSON INCLUDE HEADERS CHECK \( WITH \(BODY AS bytes\, HEADERS AS headers\) check_expression\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sources.name = 'webhook_source'`
		testhelpers.MockSourceScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sources.id = 'u1'`
		testhelpers.MockSourceScan(mock, pp)

		// Query Subsources
		ps := `WHERE mz_object_dependencies.referenced_object_id = 'u1' AND mz_objects.type = 'source'`
		testhelpers.MockSubsourceScan(mock, ps)

		if err := sourceWebhookCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
