resource "materialize_source_webhook" "example_webhook" {
  name            = "example_webhook"
  cluster_name    = materialize_cluster.example_cluster.name
  body_format     = "json"
  include_headers = false
}
