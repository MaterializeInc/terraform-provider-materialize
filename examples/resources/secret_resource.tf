resource "materialize_secret" "example_secret" {
  name  = "secret"
  value = "decode('c2VjcmV0Cg==', 'base64')"
}