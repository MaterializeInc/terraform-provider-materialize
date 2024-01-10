data "materialize_region" "all" {}

output "region" {
  value = data.materialize_region.all
}
