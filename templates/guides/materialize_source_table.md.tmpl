# Source versioning: migrating to `materialize_source_table_{source}` Resource

In previous versions of the Materialize Terraform provider, source tables were defined within the source resource itself and were considered subsources of the source rather than separate entities.

This guide will walk you through the process of migrating your existing source table definitions to the new `materialize_source_table_{source}` resource.

For each source type (e.g., MySQL, Postgres, etc.), you will need to create a new `materialize_source_table_{source}` resource for each table that was previously defined within the source resource. This ensures that the tables are preserved during the migration process. For Kafka sources, you will need to create at least one `materialize_source_table_kafka` table to hold data for the kafka topic.

## Old Approach

Previously, source tables were defined directly within the source resource:

### Example: MySQL Source

```hcl
resource "materialize_source_mysql" "mysql_source" {
  name         = "mysql_source"
  cluster_name = "cluster_name"

  mysql_connection {
    name = materialize_connection_mysql.mysql_connection.name
  }

  table {
    upstream_name        = "mysql_table1"
    upstream_schema_name = "shop"
    name                 = "mysql_table1_local"
  }
}
```

### Example: Kafka Source

```hcl
resource "materialize_source_kafka" "example_source_kafka_format_text" {
  name         = "source_kafka_text"
  comment      = "source kafka comment"
  cluster_name = materialize_cluster.cluster_source.name
  topic        = "topic1"

  kafka_connection {
    name          = materialize_connection_kafka.kafka_connection.name
    schema_name   = materialize_connection_kafka.kafka_connection.schema_name
    database_name = materialize_connection_kafka.kafka_connection.database_name
  }
  key_format {
    text = true
  }
  value_format {
    text = true
  }
}
```

## New Approach

The new approach separates source definitions and table definitions. You will now create the source without specifying the tables, and then define each table using the `materialize_source_table_{source}` resource.

## Manual Migration Process

This manual migration process requires users to create new source tables using the new `materialize_source_table_{source}` resource and then remove the old ones. We'll cover examples for both MySQL and Kafka sources.

### Step 1: Define `materialize_source_table_{source}` Resources

Before making any changes to your existing source resources, create new `materialize_source_table_{source}` resources for each table that is currently defined within your sources.

#### MySQL Example:

```hcl
resource "materialize_source_table_mysql" "mysql_table_from_source" {
  name           = "mysql_table1_from_source"
  schema_name    = "public"
  database_name  = "materialize"

  source {
    name = materialize_source_mysql.mysql_source.name
    // Define the schema and database for the source if needed
  }

  upstream_name        = "mysql_table1"
  upstream_schema_name = "shop"

  ignore_columns = ["about"]
}
```

#### Kafka Example:

```hcl
resource "materialize_source_table_kafka" "kafka_table_from_source" {
  name           = "kafka_table_from_source"
  schema_name    = "public"
  database_name  = "materialize"

  source_name {
    name = materialize_source_kafka.kafka_source.name
  }

  key_format {
    text = true
  }

  value_format {
    text = true
  }

}
```

### Step 2: Apply the Changes

Run `terraform plan` and `terraform apply` to create the new `materialize_source_table_{source}` resources.

### Step 3: Remove Table Blocks from Source Resources

Once the new `materialize_source_table_{source}` resources are successfully created, remove all the deprecated and table-specific attributes from your source resources.

#### MySQL Example:

For MySQL sources, remove the `table` block and any table-specific attributes from the source resource:

```hcl
resource "materialize_source_mysql" "mysql_source" {
  name         = "mysql_source"
  cluster_name = "cluster_name"

  mysql_connection {
    name = materialize_connection_mysql.mysql_connection.name
  }

  // Remove the table blocks from here
  - table {
  -   upstream_name        = "mysql_table1"
  -   upstream_schema_name = "shop"
  -   name                 = "mysql_table1_local"
  -
  -   ignore_columns = ["about"]
  -
  ...
}
```

#### Kafka Example:

For Kafka sources, remove the `format`, `include_key`, `include_headers`, and other table-specific attributes from the source resource:

```hcl
resource "materialize_source_kafka" "kafka_source" {
  name         = "kafka_source"
  cluster_name = "cluster_name"

  kafka_connection {
    name = materialize_connection_kafka.kafka_connection.name
  }

  topic = "example_topic"

  lifecycle {
    ignore_changes = [
      include_key,
      include_headers,
      format,
      ...
    ]
  }
  // Remove the format, include_key, include_headers, and other table-specific attributes
}
```

In the `lifecycle` block, add the `ignore_changes` meta-argument to prevent Terraform from trying to update these attributes during subsequent applies, that way Terraform won't try to update these values based on incomplete information from the state as they will no longer be defined in the source resource itself but in the new `materialize_source_table_{source}` resources.

> Note: We will make the changes to those attributes a no-op, so the `ignore_changes` block will not be necessary.

### Step 4: Update Terraform State

After removing the `table` blocks and the table/topic specific attributes from your source resources, run `terraform plan` and `terraform apply` again to update the Terraform state and apply the changes.

### Step 5: Verify the Migration

After applying the changes, verify that your tables are still correctly set up in Materialize by checking the table definitions using Materialize's SQL commands.

## Importing Existing Tables

To import existing tables into your Terraform state, use the following command:

```bash
terraform import materialize_source_table_{source}.table_name <region>:<table_id>
```

Replace `{source}` with the appropriate source type (e.g., `mysql`, `kafka`), `<region>` with the actual region, and `<table_id>` with the table ID.

### Important Note on Importing

Due to limitations in the current read function, not all properties of the source tables are available when importing. To work around this, you'll need to use the `ignore_changes` lifecycle meta-argument for certain attributes that can't be read back from the state.

For example, for a Kafka source table:

```hcl
resource "materialize_source_table_kafka" "kafka_table_from_source" {
  name           = "kafka_table_from_source"
  schema_name    = "public"
  database_name  = "materialize"

  source_name = materialize_source_kafka.kafka_source.name

  include_key     = true
  include_headers = true

  envelope {
    upsert = true
  }

  lifecycle {
    ignore_changes = [
      include_key,
      include_headers,
      envelope
      ... Add other attributes here as needed
    ]
  }
}
```

This `ignore_changes` block tells Terraform to ignore changes to these attributes during subsequent applies, preventing Terraform from trying to update these values based on incomplete information from the state.

After importing, you may need to manually update these ignored attributes in your Terraform configuration to match the actual state in Materialize.

## Future Improvements

The Kafka and Webhooks sources are currently being implemented. Once these changes, the migration process will be updated to include them.
