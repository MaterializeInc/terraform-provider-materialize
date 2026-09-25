package resources

import (
	"context"
	"testing"

	"github.com/MaterializeInc/terraform-provider-materialize/pkg/testhelpers"
	"github.com/MaterializeInc/terraform-provider-materialize/pkg/utils"

	sqlmock "github.com/DATA-DOG/go-sqlmock"
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/require"
)

var inSinkIceberg = map[string]interface{}{
	"name":          "iceberg_sink",
	"schema_name":   "schema",
	"database_name": "database",
	"cluster_name":  "my_cluster",
	"from": []interface{}{
		map[string]interface{}{
			"name":          "my_view",
			"schema_name":   "public",
			"database_name": "database",
		},
	},
	"iceberg_catalog_connection": []interface{}{
		map[string]interface{}{
			"name":          "iceberg_catalog",
			"schema_name":   "public",
			"database_name": "materialize",
		},
	},
	"namespace": "my_namespace",
	"table":     "my_table",
	"aws_connection": []interface{}{
		map[string]interface{}{
			"name":          "aws_conn",
			"schema_name":   "public",
			"database_name": "materialize",
		},
	},
	"key":              []interface{}{"id"},
	"key_not_enforced": false,
	"commit_interval":  "10s",
}

func TestResourceSinkIcebergCreate(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SinkIceberg().Schema, inSinkIceberg)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."iceberg_sink" IN CLUSTER "my_cluster" FROM "database"."public"."my_view" INTO ICEBERG CATALOG CONNECTION "materialize"."public"."iceberg_catalog" \(NAMESPACE = 'my_namespace', TABLE = 'my_table'\) USING AWS CONNECTION "materialize"."public"."aws_conn" KEY \(id\) MODE UPSERT WITH \(COMMIT INTERVAL = '10s'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sinks.name = 'iceberg_sink'`
		testhelpers.MockSinkScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sinks.id = 'u1'`
		testhelpers.MockSinkScan(mock, pp)

		if err := sinkIcebergCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSinkIcebergCreateWithKeyNotEnforced(t *testing.T) {
	r := require.New(t)
	inSinkIcebergWithNotEnforced := map[string]interface{}{
		"name":          "iceberg_sink",
		"schema_name":   "schema",
		"database_name": "database",
		"cluster_name":  "my_cluster",
		"from": []interface{}{
			map[string]interface{}{
				"name":          "my_view",
				"schema_name":   "public",
				"database_name": "database",
			},
		},
		"iceberg_catalog_connection": []interface{}{
			map[string]interface{}{
				"name":          "iceberg_catalog",
				"schema_name":   "public",
				"database_name": "materialize",
			},
		},
		"namespace": "my_namespace",
		"table":     "my_table",
		"aws_connection": []interface{}{
			map[string]interface{}{
				"name":          "aws_conn",
				"schema_name":   "public",
				"database_name": "materialize",
			},
		},
		"key":              []interface{}{"id"},
		"key_not_enforced": true,
		"commit_interval":  "30s",
	}
	d := schema.TestResourceDataRaw(t, SinkIceberg().Schema, inSinkIcebergWithNotEnforced)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."iceberg_sink" IN CLUSTER "my_cluster" FROM "database"."public"."my_view" INTO ICEBERG CATALOG CONNECTION "materialize"."public"."iceberg_catalog" \(NAMESPACE = 'my_namespace', TABLE = 'my_table'\) USING AWS CONNECTION "materialize"."public"."aws_conn" KEY \(id\) NOT ENFORCED MODE UPSERT WITH \(COMMIT INTERVAL = '30s'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sinks.name = 'iceberg_sink'`
		testhelpers.MockSinkScan(mock, ip)

		// Query Params
		pp := `WHERE mz_sinks.id = 'u1'`
		testhelpers.MockSinkScan(mock, pp)

		if err := sinkIcebergCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSinkIcebergCreateAppend(t *testing.T) {
	r := require.New(t)
	in := map[string]interface{}{
		"name":          "iceberg_sink",
		"schema_name":   "schema",
		"database_name": "database",
		"cluster_name":  "my_cluster",
		"from": []interface{}{
			map[string]interface{}{"name": "my_view", "schema_name": "public", "database_name": "database"},
		},
		"iceberg_catalog_connection": []interface{}{
			map[string]interface{}{"name": "iceberg_catalog", "schema_name": "public", "database_name": "materialize"},
		},
		"namespace":       "my_namespace",
		"table":           "my_table",
		"mode":            "append",
		"commit_interval": "1m",
	}
	d := schema.TestResourceDataRaw(t, SinkIceberg().Schema, in)
	r.NotNil(d)

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Create: no KEY and no USING AWS CONNECTION
		mock.ExpectExec(
			`CREATE SINK "database"."schema"."iceberg_sink" IN CLUSTER "my_cluster" FROM "database"."public"."my_view" INTO ICEBERG CATALOG CONNECTION "materialize"."public"."iceberg_catalog" \(NAMESPACE = 'my_namespace', TABLE = 'my_table'\) MODE APPEND WITH \(COMMIT INTERVAL = '1m'\);`,
		).WillReturnResult(sqlmock.NewResult(1, 1))

		// Query Id
		ip := `WHERE mz_databases.name = 'database' AND mz_schemas.name = 'schema' AND mz_sinks.name = 'iceberg_sink'`
		testhelpers.MockSinkIcebergScan(mock, ip, "append")

		// Query Params
		pp := `WHERE mz_sinks.id = 'u1'`
		testhelpers.MockSinkIcebergScan(mock, pp, "append")

		if err := sinkIcebergCreate(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
	})
}

func TestResourceSinkIcebergReadSetsModeFromEnvelope(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SinkIceberg().Schema, inSinkIceberg)
	r.NotNil(d)
	d.SetId("u1")
	r.Equal("upsert", d.Get("mode"), "schema default")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		pp := `WHERE mz_sinks.id = 'u1'`
		testhelpers.MockSinkIcebergScan(mock, pp, "append")

		if err := sinkIcebergRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
		r.Equal("append", d.Get("mode"))
	})
}

func TestResourceSinkIcebergRead(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SinkIceberg().Schema, inSinkIceberg)
	r.NotNil(d)

	// Set id before read
	d.SetId("u1")

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		// Query Params
		pp := `WHERE mz_sinks.id = 'u1'`
		testhelpers.MockSinkScan(mock, pp)

		if err := sinkRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}

		if d.Id() != "aws/us-east-1:u1" {
			t.Fatalf("unexpected id of %s", d.Id())
		}
	})
}

func TestResourceSinkIcebergKeyModePlanCheck(t *testing.T) {
	r := require.New(t)
	res := SinkIceberg()
	plan := func(extra map[string]interface{}) error {
		cfg := map[string]interface{}{
			"name":                       "iceberg_sink",
			"from":                       []interface{}{map[string]interface{}{"name": "my_view"}},
			"iceberg_catalog_connection": []interface{}{map[string]interface{}{"name": "iceberg_catalog"}},
			"namespace":                  "ns",
			"table":                      "tbl",
			"commit_interval":            "10s",
		}
		for k, v := range extra {
			cfg[k] = v
		}
		_, err := res.Diff(context.TODO(), nil, terraform.NewResourceConfigRaw(cfg), nil)
		return err
	}

	r.NoError(plan(map[string]interface{}{"key": []interface{}{"id"}}))
	r.NoError(plan(map[string]interface{}{"key": []interface{}{"id"}, "key_not_enforced": true}))
	r.NoError(plan(map[string]interface{}{"mode": "append"}))
	r.ErrorContains(plan(map[string]interface{}{}), "key is required")
	r.ErrorContains(plan(map[string]interface{}{"mode": "append", "key": []interface{}{"id"}}), "key is not allowed")
	r.ErrorContains(plan(map[string]interface{}{"mode": "append", "key_not_enforced": true}), "key_not_enforced has no effect")
}

func TestResourceSinkIcebergRemovingAwsConnectionDoesNotReplace(t *testing.T) {
	r := require.New(t)
	res := SinkIceberg()
	cfg := map[string]interface{}{
		"name":                       "iceberg_sink",
		"from":                       []interface{}{map[string]interface{}{"name": "my_view"}},
		"iceberg_catalog_connection": []interface{}{map[string]interface{}{"name": "iceberg_catalog"}},
		"namespace":                  "ns",
		"table":                      "tbl",
		"key":                        []interface{}{"id"},
		"commit_interval":            "10s",
	}
	withAws := map[string]interface{}{}
	for k, v := range cfg {
		withAws[k] = v
	}
	withAws["aws_connection"] = []interface{}{map[string]interface{}{"name": "aws_conn"}}

	// state as written by an older provider: the block is set
	state := &terraform.InstanceState{ID: "u1", Attributes: map[string]string{
		"name": "iceberg_sink", "schema_name": "public", "database_name": "materialize",
		"from.#": "1", "from.0.name": "my_view", "from.0.schema_name": "public", "from.0.database_name": "materialize",
		"iceberg_catalog_connection.#": "1", "iceberg_catalog_connection.0.name": "iceberg_catalog", "iceberg_catalog_connection.0.schema_name": "public", "iceberg_catalog_connection.0.database_name": "materialize",
		"aws_connection.#": "1", "aws_connection.0.name": "aws_conn", "aws_connection.0.schema_name": "public", "aws_connection.0.database_name": "materialize",
		"namespace": "ns", "table": "tbl", "key.#": "1", "key.0": "id", "key_not_enforced": "false", "mode": "upsert", "commit_interval": "10s",
	}}

	// dropping the block from the configuration is a no-op
	diff, err := res.Diff(context.TODO(), state, terraform.NewResourceConfigRaw(cfg), nil)
	r.NoError(err)
	r.True(diff == nil || diff.Empty(), "removing aws_connection must not produce a diff, got %v", diff)

	// adding the block where there was none is still a real change
	delete(state.Attributes, "aws_connection.#")
	delete(state.Attributes, "aws_connection.0.name")
	delete(state.Attributes, "aws_connection.0.schema_name")
	delete(state.Attributes, "aws_connection.0.database_name")
	diff, err = res.Diff(context.TODO(), state, terraform.NewResourceConfigRaw(withAws), nil)
	r.NoError(err)
	r.False(diff == nil || diff.Empty(), "adding aws_connection should still show up in the plan")
}

// Older Materialize versions may not report an Iceberg envelope; the read must
// leave mode as it is rather than blank it.
func TestResourceSinkIcebergReadKeepsModeOnUnknownEnvelope(t *testing.T) {
	r := require.New(t)
	d := schema.TestResourceDataRaw(t, SinkIceberg().Schema, inSinkIceberg)
	d.SetId("u1")
	r.NoError(d.Set("mode", "append"))

	testhelpers.WithMockProviderMeta(t, func(db *utils.ProviderMeta, mock sqlmock.Sqlmock) {
		testhelpers.MockSinkIcebergScan(mock, `WHERE mz_sinks.id = 'u1'`, "")
		if err := sinkIcebergRead(context.TODO(), d, db); err != nil {
			t.Fatal(err)
		}
		r.Equal("append", d.Get("mode"))
	})
}

func icebergSinkState(extra map[string]string) *terraform.InstanceState {
	a := map[string]string{
		"name": "iceberg_sink", "schema_name": "public", "database_name": "materialize",
		"from.#": "1", "from.0.name": "my_view", "from.0.schema_name": "public", "from.0.database_name": "materialize",
		"iceberg_catalog_connection.#": "1", "iceberg_catalog_connection.0.name": "iceberg_catalog", "iceberg_catalog_connection.0.schema_name": "public", "iceberg_catalog_connection.0.database_name": "materialize",
		"namespace": "ns", "table": "tbl", "key.#": "1", "key.0": "id", "key_not_enforced": "false", "commit_interval": "10s",
	}
	for k, v := range extra {
		a[k] = v
	}
	return &terraform.InstanceState{ID: "u1", Attributes: a}
}

var icebergSinkUpsertConfig = map[string]interface{}{
	"name":                       "iceberg_sink",
	"from":                       []interface{}{map[string]interface{}{"name": "my_view"}},
	"iceberg_catalog_connection": []interface{}{map[string]interface{}{"name": "iceberg_catalog"}},
	"namespace":                  "ns",
	"table":                      "tbl",
	"key":                        []interface{}{"id"},
	"commit_interval":            "10s",
}

// State written before mode existed has no mode. Planning it without a refresh
// must not replace the sink, while a real change to append still does.
func TestResourceSinkIcebergUpgradeWithoutModeDoesNotReplace(t *testing.T) {
	r := require.New(t)
	res := SinkIceberg()

	diff, err := res.Diff(context.TODO(), icebergSinkState(nil), terraform.NewResourceConfigRaw(icebergSinkUpsertConfig), nil)
	r.NoError(err)
	r.False(diff != nil && diff.RequiresNew(), "an upsert sink from an older provider must not be replaced, got %v", diff)

	appendCfg := map[string]interface{}{}
	for k, v := range icebergSinkUpsertConfig {
		if k != "key" {
			appendCfg[k] = v
		}
	}
	appendCfg["mode"] = "append"
	diff, err = res.Diff(context.TODO(), icebergSinkState(map[string]string{"mode": "upsert"}), terraform.NewResourceConfigRaw(appendCfg), nil)
	r.NoError(err)
	r.True(diff != nil && diff.RequiresNew(), "switching an existing sink to append must replace")
}

// Only removing the deprecated block is suppressed; pointing it at a
// connection named "0" is a real change.
func TestResourceSinkIcebergAwsConnectionNamedZeroIsNotSuppressed(t *testing.T) {
	r := require.New(t)
	state := icebergSinkState(map[string]string{
		"mode": "upsert", "aws_connection.#": "1", "aws_connection.0.name": "aws_conn",
		"aws_connection.0.schema_name": "public", "aws_connection.0.database_name": "materialize",
	})
	cfg := map[string]interface{}{}
	for k, v := range icebergSinkUpsertConfig {
		cfg[k] = v
	}
	cfg["aws_connection"] = []interface{}{map[string]interface{}{"name": "0"}}
	diff, err := SinkIceberg().Diff(context.TODO(), state, terraform.NewResourceConfigRaw(cfg), nil)
	r.NoError(err)
	r.False(diff == nil || diff.Empty(), "renaming aws_connection to 0 must show up in the plan")
}
