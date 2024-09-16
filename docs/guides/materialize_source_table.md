---
page_title: "Source versioning: migrating to `materialize_source_table` Resource"
subcategory: ""
description: |-

---

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

The same approach was used for other source types such as Postgres and the load generator sources.

## New Approach

The new approach separates source definitions and table definitions. You will now create the source without specifying the tables, and then define each table using the `materialize_source_table_mysql` resource.

## Manual Migration Process

This manual migration process requires users to create new source tables using the new `materialize_source_table_{source}` resource and then remove the old ones. In this example, we will use MySQL as the source type.

### Step 1: Define `materialize_source_table_mysql` Resources

Before making any changes to your existing source resources, create new `materialize_source_table_mysql` resources for each table that is currently defined within your sources. This ensures that the tables are preserved during the migration:

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

### Step 2: Apply the Changes

Run `terraform plan` and `terraform apply` to create the new `materialize_source_table_mysql` resources. This step ensures that the tables are defined separately from the source and are not removed from Materialize.

> **Note:** This will start an ingestion process for the newly created source tables.

### Step 3: Remove Table Blocks from Source Resources

Once the new `materialize_source_table_mysql` resources are successfully created, you can safely remove the `table` blocks from your existing source resources:

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

This will drop the old tables from the source resources.

### Step 4: Update Terraform State

After removing the `table` blocks from your source resources, run `terraform plan` and `terraform apply` again to update the Terraform state and apply the changes.

### Step 5: Verify the Migration

After applying the changes, verify that your tables are still correctly set up in Materialize by checking the table definitions using Materialize’s SQL commands.

During the migration, you can use both the old `table` blocks and the new `materialize_source_table_{source}` resources simultaneously. This allows for a gradual transition until the old method is fully deprecated.

The same approach can be used for other source types such as Postgres, eg. `materialize_source_table_postgres`.

## Automated Migration Process (TBD)

> **Note:** This will still not work as the previous source tables are considered subsources of the source and are missing from the `mz_tables` table in Materialize so we can't import them directly without recreating them.

Once the migration on the Materialize side has been implemented, a more automated migration process will be available. The steps will include:

### Step 1: Define `materialize_source_table_{source}` Resources

First, define the new `materialize_source_table_mysql` resources for each table:

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

### Step 2: Modify the Existing Source Resource

Next, modify the existing source resource by removing the `table` blocks and adding an `ignore_changes` directive for the `table` attribute. This prevents Terraform from trying to delete the tables:

```hcl
resource "materialize_source_mysql" "mysql_source" {
  name         = "mysql_source"
  cluster_name = "cluster_name"

  mysql_connection {
    name = materialize_connection_mysql.mysql_connection.name
  }

  lifecycle {
    ignore_changes = [table]
  }
}
```

- **`lifecycle { ignore_changes = [table] }`**: This directive tells Terraform to ignore changes to the `table` attribute, preventing it from trying to delete tables that were previously defined in the source resource.

### Step 3: Import the Existing Tables

You can then import the existing tables into the new `materialize_source_table_mysql` resources without disrupting your existing setup:

```bash
terraform import materialize_source_table_mysql.mysql_table_from_source <region>:<table_id>
```

Replace `<region>` with the actual region and `<table_id>` with the table ID. You can find the table ID by querying the `mz_tables` table.

### Step 4: Run Terraform Plan and Apply

Finally, run `terraform plan` and `terraform apply` to ensure that everything is correctly set up without triggering any unwanted deletions.

This approach allows you to migrate your tables safely without disrupting your existing setup.

## Importing Existing Tables

To import existing tables into your Terraform state using the manual migration process, use the following command:

```bash
terraform import materialize_source_table_mysql.table_name <region>:<table_id>
```

Ensure you replace `<region>` with the region where the table is located and `<table_id>` with the ID of the table.

> **Note:** The `upstream_name` and `upstream_schema_name` attributes are not yet implemented on the Materialize side, so the import process will not work until these changes are made.

## Future Improvements

The Kafka and Webhooks sources are currently being implemented. Once these changes, the migration process will be updated to include them.
