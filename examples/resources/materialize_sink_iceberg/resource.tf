# Create an Iceberg sink to export data to AWS S3 Tables
resource "materialize_sink_iceberg" "example" {
  name         = "iceberg_sink"
  cluster_name = "quickstart"

  from {
    name          = "my_materialized_view"
    database_name = "materialize"
    schema_name   = "public"
  }

  iceberg_catalog_connection {
    name          = "iceberg_catalog_connection"
    database_name = "materialize"
    schema_name   = "public"
  }

  namespace = "my_namespace"
  table     = "my_table"

  key              = ["id"]
  key_not_enforced = true
  commit_interval  = "10s"
}

# Append every change as a row instead of keeping rows current by key. Takes no
# key. Databricks Unity Catalog tables only accept this mode.
resource "materialize_sink_iceberg" "append" {
  name         = "iceberg_sink_append"
  cluster_name = "quickstart"

  from {
    name          = "my_materialized_view"
    database_name = "materialize"
    schema_name   = "public"
  }

  iceberg_catalog_connection {
    name          = "databricks_catalog_connection"
    database_name = "materialize"
    schema_name   = "public"
  }

  namespace       = "my_schema"
  table           = "my_table"
  mode            = "append"
  commit_interval = "1m"
}
