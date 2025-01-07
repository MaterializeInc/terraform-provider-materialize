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

var inSourceTableWebhook = map[string]interface{}{
	"name":          "webhook_table",
	"schema_name":   "schema",
	"database_name": "database",
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

func TestResourceSourceTableWebhookCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTableWebhook().Schema, inSourceTableWebhook)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE TABLE "database"."schema"."webhook_table" FROM WEBHOOK BODY FORMAT JSON INCLUDE HEADERS CHECK \( WITH \(BODY AS bytes\, HEADERS AS headers\) check_expression\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_tables.name = 'webhook_table'`
		testhelpers.MockSourceTableWebhookScan(mock, ip)

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableWebhookScan(mock, pp)

		if err := sourceTableWebhookCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableWebhookDelete(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTableWebhook().Schema, inSourceTableWebhook)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP TABLE "database"."schema"."webhook_table";`).WillReturnResult(sqlmock.NewResult(1, 1))

		if err := sourceTableDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableWebhookUpdate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTableWebhook().Schema, inSourceTableWebhook)
	d.SetId("u1")
	d.Set("name", "webhook_table")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER TABLE "database"."schema"."" RENAME TO "database"."schema"."webhook_table"`).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableWebhookScan(mock, pp)

		if err := sourceTableWebhookUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSourceTableWebhookRead(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTableWebhook().Schema, inSourceTableWebhook)
	d.SetId("u1")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableWebhookScan(mock, pp)

		if err := sourceTableWebhookRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
