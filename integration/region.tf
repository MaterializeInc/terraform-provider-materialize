data "materialize_region" "all" {}

output "region" {
  value = data.materialize_region.all
}

resource "materialize_region" "example_region" {
  region_id = "aws/us-east-1"
}
