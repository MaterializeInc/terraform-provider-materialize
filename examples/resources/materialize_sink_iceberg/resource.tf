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

# CREATE SINK iceberg_sink
#   IN CLUSTER quickstart
#   FROM my_materialized_view
#   INTO ICEBERG CATALOG CONNECTION iceberg_catalog_connection (
#     NAMESPACE = 'my_namespace',
#     TABLE = 'my_table'
#   )
#   USING AWS CONNECTION aws_connection
#   KEY (id) NOT ENFORCED
#   MODE UPSERT
#   WITH (COMMIT INTERVAL = '10s');


# Example with multiple key columns
resource "materialize_sink_iceberg" "multi_key_example" {
  name         = "iceberg_sink_multi_key"
  cluster_name = "quickstart"

  from {
    name          = "orders"
    database_name = "materialize"
    schema_name   = "public"
  }

  iceberg_catalog_connection {
    name          = "iceberg_catalog_connection"
    database_name = "materialize"
    schema_name   = "public"
  }

  namespace = "sales"
  table     = "orders_table"

  aws_connection {
    name          = "aws_connection"
    database_name = "materialize"
    schema_name   = "public"
  }

  key             = ["order_id", "tenant_id"]
  commit_interval = "1m"
}

# CREATE SINK iceberg_sink_multi_key
#   IN CLUSTER quickstart
#   FROM orders
#   INTO ICEBERG CATALOG CONNECTION iceberg_catalog_connection (
#     NAMESPACE = 'sales',
#     TABLE = 'orders_table'
#   )
#   USING AWS CONNECTION aws_connection
#   KEY (order_id, tenant_id)
#   MODE UPSERT
#   WITH (COMMIT INTERVAL = '1m');
