resource "materialize_source_table_load_generator" "load_generator_table_from_source" {
  name           = "load_generator_table_from_source"
  schema_name    = "public"
  database_name  = "materialize"

  # The load generator source must be of type: `auction_options`, `marketing_options`, and `tpch_options` load generator sources.
  source {
    name          = materialize_source_load_generator.example.name
    schema_name   = materialize_source_load_generator.example.schema_name
    database_name = materialize_source_load_generator.example.database_name
  }

  upstream_name         = "load_generator_table_name" # The name of the table from the load generator

}
