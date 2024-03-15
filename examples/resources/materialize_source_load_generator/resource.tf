resource "materialize_source_load_generator" "example_source_load_generator" {
  name         = "source_load_generator"
  schema_name  = "schema"
  cluster_name = "quickstart"

  load_generator_type = "COUNTER"

  counter_options {
    tick_interval = "500ms"
  }
}

# CREATE SOURCE schema.source_load_generator
#   FROM LOAD GENERATOR COUNTER
#   (TICK INTERVAL '500ms');
