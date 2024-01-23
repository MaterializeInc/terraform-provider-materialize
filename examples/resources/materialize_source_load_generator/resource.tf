resource "materialize_source_load_generator" "example_source_load_generator" {
  name         = "source_load_generator"
  schema_name  = "schema"
  cluster_name = "quickstart"

  load_generator_type = "COUNTER"

  counter_options {
    tick_interval = "500ms"
    scale_factor  = 0.01
  }
}

# CREATE SOURCE schema.source_load_generator
#   FROM LOAD GENERATOR COUNTER
#   (TICK INTERVAL '500ms' SCALE FACTOR 0.01)
#   WITH (SIZE = '3xsmall');
