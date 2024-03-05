resource "materialize_source_mysql" "test" {
  name          = "source_mysql"
  schema_name   = materialize_schema.test.name
  database_name = materialize_database.test.name

  cluster_name = "quickstart"

  mysql_connection {
    name = materialize_connection_mysql.test.name
  }

  table {
    name  = "shop.mysql_table1"
    alias = "alias_mysql_table1"
  }
  table {
    name  = "shop.mysql_table2"
    alias = "alias_mysql_table2"
  }
}

# CREATE SOURCE schema.source_mysql
#   FROM MYSQL CONNECTION "database"."schema"."pg_connection" (PUBLICATION 'mz_source')
#   FOR TABLES (schema1.table_1 AS s1_table_1, schema2_table_1 AS s2_table_1);
