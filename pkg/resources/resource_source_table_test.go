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

var inSourceTable = map[string]interface{}{
	"name":          "table",
	"schema_name":   "schema",
	"database_name": "database",
	"source": []interface{}{
		map[string]interface{}{
			"name":          "source",
			"schema_name":   "public",
			"database_name": "materialize",
		},
	},
	"region": "aws/us-east-1",
}

func TestSourceTableRead(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"name":          {Type: schema.TypeString, Required: true},
		"schema_name":   {Type: schema.TypeString, Optional: true},
		"database_name": {Type: schema.TypeString, Optional: true},
		"source": {
			Type:     schema.TypeList,
			Required: true,
			Elem: &schema.Resource{
				Schema: map[string]*schema.Schema{
					"name":          {Type: schema.TypeString, Required: true},
					"schema_name":   {Type: schema.TypeString, Required: true},
					"database_name": {Type: schema.TypeString, Required: true},
				},
			},
		},
		"ownership_role": {Type: schema.TypeString, Optional: true},
		"comment":        {Type: schema.TypeString, Optional: true},
		"region":         {Type: schema.TypeString, Optional: true},
	}, inSourceTable)
	d.SetId("aws/us-east-1:u1")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableScan(mock, pp)

		if err := sourceTableRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		r.Equal("table", d.Get("name").(string))
		r.Equal("schema", d.Get("schema_name").(string))
		r.Equal("database", d.Get("database_name").(string))
		r.Equal("materialize", d.Get("ownership_role").(string))
		r.Equal("comment", d.Get("comment").(string))
	})
}

func TestSourceTableUpdateRename(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SourceTablePostgres().Schema, inSourceTable)
	d.SetId("u1")
	d.Set("name", "old_table")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`ALTER TABLE "database"."schema"."" RENAME TO "database"."schema"."table";`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		pp := `WHERE mz_tables.id = 'u1'`
		testhelpers.MockSourceTableScan(mock, pp)

		if err := sourceTableUpdate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestSourceTableDelete(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, map[string]*schema.Schema{
		"name":          {Type: schema.TypeString, Required: true},
		"schema_name":   {Type: schema.TypeString, Optional: true},
		"database_name": {Type: schema.TypeString, Optional: true},
		"region":        {Type: schema.TypeString, Optional: true},
	}, inSourceTable)
	d.SetId("aws/us-east-1:u1")
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		mock.ExpectExec(`DROP TABLE "database"."schema"."table";`).
			WillReturnResult(sqlmock.NewResult(1, 1))

		if err := sourceTableDelete(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}
