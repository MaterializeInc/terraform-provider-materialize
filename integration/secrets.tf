resource "materialize_secret" "example_secret" {
  name  = "example"
  value = "decode('c2VjcmV0Cg==', 'base64')"
}