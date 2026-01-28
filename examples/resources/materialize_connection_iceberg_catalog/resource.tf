# Create an AWS connection for Iceberg catalog authentication
resource "materialize_secret" "aws_secret_access_key" {
  name  = "aws_secret_access_key"
  value = "YOUR_SECRET_ACCESS_KEY"
}

resource "materialize_connection_aws" "aws_connection" {
  name = "aws_connection"
  access_key_id {
    text = "YOUR_ACCESS_KEY_ID"
  }
  secret_access_key {
    name          = materialize_secret.aws_secret_access_key.name
    database_name = materialize_secret.aws_secret_access_key.database_name
    schema_name   = materialize_secret.aws_secret_access_key.schema_name
  }
}

# Create an Iceberg catalog connection using AWS S3 Tables
resource "materialize_connection_iceberg_catalog" "example" {
  name         = "iceberg_catalog_connection"
  catalog_type = "s3tablesrest"
  url          = "https://s3tables.us-east-1.amazonaws.com/iceberg"
  warehouse    = "arn:aws:s3tables:us-east-1:123456789012:bucket/my-bucket"
  aws_connection {
    name          = materialize_connection_aws.aws_connection.name
    database_name = materialize_connection_aws.aws_connection.database_name
    schema_name   = materialize_connection_aws.aws_connection.schema_name
  }
}

# CREATE CONNECTION iceberg_catalog_connection TO ICEBERG CATALOG (
#   CATALOG TYPE = 's3tablesrest',
#   URL = 'https://s3tables.us-east-1.amazonaws.com/iceberg',
#   WAREHOUSE = 'arn:aws:s3tables:us-east-1:123456789012:bucket/my-bucket',
#   AWS CONNECTION = aws_connection
# );
