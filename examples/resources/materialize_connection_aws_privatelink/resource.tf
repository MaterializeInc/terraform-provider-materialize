# # Create a AWS Private Connection
# Note: you need the max_aws_privatelink_connections increased for this to work:
# show max_aws_privatelink_connections;
resource "materialize_connection_aws_privatelink" "example_privatelink_connection" {
  name               = "example_privatelink_connection"
  schema_name        = "public"
  service_name       = "com.amazonaws.us-east-1.materialize.example"
  availability_zones = ["use1-az2", "use1-az6"]
}

# CREATE CONNECTION example_privatelink_connection TO AWS PRIVATELINK (
#     SERVICE NAME 'com.amazonaws.us-east-1.materialize.example',
#     AVAILABILITY ZONES ('use1-az2', 'use1-az6')
# );
