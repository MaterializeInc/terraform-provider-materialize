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

# Create an Iceberg catalog connection to Databricks Unity Catalog over the
# Iceberg REST API. Unity Catalog only hands out storage credentials through
# credential vending, so access_delegation is required there.
resource "materialize_secret" "databricks_oauth" {
  name  = "databricks_oauth"
  value = "<client_id>:<client_secret>"
}

resource "materialize_connection_iceberg_catalog" "databricks" {
  name              = "databricks_catalog_connection"
  catalog_type      = "rest"
  url               = "https://<workspace>.cloud.databricks.com/api/2.1/unity-catalog/iceberg-rest"
  warehouse         = "<catalog_name>"
  oauth2_server_url = "https://<workspace>.cloud.databricks.com/oidc/v1/token"
  scope             = "all-apis"
  access_delegation = "vended-credentials"
  credential {
    secret {
      name          = materialize_secret.databricks_oauth.name
      database_name = materialize_secret.databricks_oauth.database_name
      schema_name   = materialize_secret.databricks_oauth.schema_name
    }
  }
}

# CREATE CONNECTION databricks_catalog_connection TO ICEBERG CATALOG (
#   CATALOG TYPE = 'rest',
#   URL = 'https://<workspace>.cloud.databricks.com/api/2.1/unity-catalog/iceberg-rest',
#   WAREHOUSE = '<catalog_name>',
#   CREDENTIAL = SECRET databricks_oauth,
#   OAUTH2 SERVER URL = 'https://<workspace>.cloud.databricks.com/oidc/v1/token',
#   SCOPE = 'all-apis',
#   ACCESS DELEGATION = 'vended-credentials'
# );
