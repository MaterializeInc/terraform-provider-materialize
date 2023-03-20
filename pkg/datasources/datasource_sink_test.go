package datasources

// func TestSinkDatasource(t *testing.T) {
// 	r := require.New(t)

// 	in := map[string]interface{}{
// 		"schema_name":   "schema",
// 		"database_name": "database",
// 	}
// 	d := schema.TestResourceDataRaw(t, Sink().Schema, in)
// 	r.NotNil(d)

// 	testhelpers.WithMockDb(t, func(db *sqlx.DB, mock sqlmock.Sqlmock) {
// 		ir := mock.NewRows([]string{"id", "name", "schema", "database"}).
// 			AddRow("u1", "view", "schema", "database")
// 		mock.ExpectQuery(`
// 		SELECT
// 			mz_sinks.id,
// 			mz_sinks.name,
// 			mz_schemas.name,
// 			mz_databases.name,
// 			mz_sinks.type,
// 			mz_sinks.size,
// 			mz_sinks.envelope_type,
// 			mz_connections.name as connection_name,
// 			mz_clusters.name as cluster_name
// 		FROM mz_sinks
// 		JOIN mz_schemas
// 			ON mz_sinks.schema_id = mz_schemas.id
// 		JOIN mz_databases
// 			ON mz_schemas.database_id = mz_databases.id
// 		LEFT JOIN mz_connections
// 			ON mz_sinks.connection_id = mz_connections.id
// 		LEFT JOIN mz_clusters
// 			ON mz_sinks.cluster_id = mz_clusters.id`).WillReturnRows(ir)

// 		if err := sinkRead(context.TODO(), d, db); err != nil {
// 			t.Fatal(err)
// 		}
// 	})

// }
