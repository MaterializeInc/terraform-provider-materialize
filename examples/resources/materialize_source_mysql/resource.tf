resource "materialize_source_mysql" "test" {
  name          = "source_mysql"
  schema_name   = materialize_schema.test.name
  database_name = materialize_database.test.name

  cluster_name = "quickstart"

  mysql_connection {
    name = materialize_connection_mysql.test.name
  }

  table {
    upstream_name        = "mysql_table1"
    upstream_schema_name = "shop"
    name                 = "mysql_table1_local"
  }

  table {
    upstream_name        = "mysql_table2"
    upstream_schema_name = "shop"
    name                 = "mysql_table2_local"
  }
}
