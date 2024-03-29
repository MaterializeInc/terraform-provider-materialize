resource "materialize_source_webhook" "example_webhook" {
  name             = "example_webhook"
  cluster_name     = materialize_cluster.cluster.name
  body_format      = "json"
  check_expression = "headers->'x-mz-api-key' = secret"
  include_headers {
    not = ["x-mz-api-key"]
  }

  check_options {
    field {
      headers = true
    }
  }

  check_options {
    field {
      secret {
        name          = materialize_secret.password.name
        database_name = materialize_secret.password.database_name
        schema_name   = materialize_secret.password.schema_name
      }
    }
    alias = "secret"
  }
}

# CREATE SOURCE example_webhook IN CLUSTER cluster FROM WEBHOOK
#   BODY FORMAT json
#   INCLUDE HEADERS ( NOT 'x-mz-api-key' )
#   CHECK (
#     WITH ( HEADERS, SECRET materialize.public.password AS secret)
#     headers->'x-mz-api-key' = secret
#   );