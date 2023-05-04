data "materialize_egress_ips" "all" {}

output "ips" {
  value = data.materialize_egress_ips.all.egress_ips
}
