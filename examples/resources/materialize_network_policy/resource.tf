resource "materialize_network_policy" "office_policy" {
  name = "office_access_policy"

  rule {
    name      = "minnesota"
    action    = "allow"
    direction = "ingress"
    address   = "2.3.4.5/32"
  }

  rule {
    name      = "new_york"
    action    = "allow"
    direction = "ingress"
    address   = "1.2.3.4/28"
  }

  comment = "Network policy for office locations"
}
