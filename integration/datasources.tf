data "materialize_egress_ips" "all" {}

# Get the egress IPs from a specific region
data "materialize_egress_ips" "us_west" {
  region = "aws/us-west-2"
}
