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

  aws_connection {
    name          = "aws_connection"
    database_name = "materialize"
    schema_name   = "public"
  }

  key              = ["id"]
  key_not_enforced = true
  commit_interval  = "10s"
}
