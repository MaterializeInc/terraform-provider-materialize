---
name: materialize-terraform-provider
description: >-
  Using the Materialize Terraform provider to manage Materialize resources
  declaratively. Covers clusters, sources (Kafka, Postgres, MySQL, SQL
  Server), sinks (Kafka, Iceberg), connections, materialized views,
  indexes, tables, roles, grants, secrets, network policies, and
  cloud-only resources (users, SSO, SCIM, app passwords). Use this skill
  whenever the user asks about writing Terraform for Materialize, creating
  or configuring Materialize resources with Terraform, importing existing
  Materialize objects into Terraform state, configuring the Materialize
  provider for Cloud or self-managed, setting up RBAC or grants via
  Terraform, creating connections or sources in Terraform, or
  troubleshooting Terraform plan/apply issues with Materialize resources.
  Also trigger when the user mentions materialize_cluster,
  materialize_source_kafka, materialize_connection_postgres, or any
  other materialize_* resource type.
---

# Materialize Terraform Provider

The `MaterializeInc/materialize` Terraform provider manages Materialize resources declaratively. It works with both Materialize Cloud (SaaS) and self-managed deployments.

## How to Use This Skill

1. **User wants to create a resource**: Look up the resource type in the catalog below, then read `docs/resources/<resource>.md` for full argument reference.
2. **User wants to list or query resources**: Check the Data Sources section and read `docs/data-sources/<resource>.md`.
3. **User needs provider config help**: See the Provider Configuration section below.
4. **User wants to import existing objects**: See the Import section.
5. **User asks about a specific resource file**: Read `docs/resources/<name>.md` directly.

## Provider Configuration

### Materialize Cloud (SaaS)

```hcl
provider "materialize" {
  password       = var.materialize_password  # app password
  default_region = "aws/us-east-1"
}
```

### Self-Managed

```hcl
provider "materialize" {
  host     = "materialized"
  port     = 6875
  username = "materialize"
  database = "materialize"
  password = var.mz_password
  sslmode  = "disable"
}
```

All arguments have environment variable equivalents: `MZ_HOST`, `MZ_PORT`, `MZ_USER`, `MZ_DATABASE`, `MZ_PASSWORD`, `MZ_SSLMODE`, `MZ_DEFAULT_REGION`.

### OIDC/SSO Authentication (Self-Managed)

```hcl
provider "materialize" {
  host     = "materialized"
  port     = 6875
  username = var.oidc_username
  password = var.oidc_id_token
  database = "materialize"
  options = {
    oidc_auth_enabled = "true"
  }
}
```

### Key Provider Arguments

| Argument | Cloud | Self-Managed | Env Var | Default |
|----------|-------|-------------|---------|---------|
| `password` | Required (app password) | Optional | `MZ_PASSWORD` | |
| `default_region` | Optional | N/A | `MZ_DEFAULT_REGION` | aws/us-east-1 |
| `host` | N/A | Optional | `MZ_HOST` | |
| `port` | N/A | Optional | `MZ_PORT` | 6875 |
| `username` | N/A | Optional | `MZ_USER` | materialize |
| `database` | Optional | Optional | `MZ_DATABASE` | materialize |
| `sslmode` | N/A | Optional | `MZ_SSLMODE` | require |

**Important:** Self-managed mode does not support Frontegg-dependent resources (app passwords, users, SSO, SCIM).

## Resource Catalog

### Compute

| Resource | Purpose | Docs |
|----------|---------|------|
| `materialize_cluster` | Managed compute cluster | `docs/resources/cluster.md` |
| `materialize_cluster_replica` | **DEPRECATED**: use `materialize_cluster` with `size` | `docs/resources/cluster_replica.md` |

**Cluster example:**
```hcl
resource "materialize_cluster" "analytics" {
  name               = "analytics"
  size               = "100cc"
  replication_factor = 1
}
```

Key cluster arguments: `name` (required), `size`, `replication_factor`, `availability_zones`, `identify_by_name` (useful for blue/green).

### Namespace

| Resource | Purpose | Docs |
|----------|---------|------|
| `materialize_database` | Top-level namespace | `docs/resources/database.md` |
| `materialize_schema` | Second-level namespace | `docs/resources/schema.md` |

Note: `materialize_database` does NOT auto-create a `public` schema.

### Connections

| Resource | Purpose | Docs |
|----------|---------|------|
| `materialize_connection_kafka` | Kafka cluster | `docs/resources/connection_kafka.md` |
| `materialize_connection_postgres` | PostgreSQL | `docs/resources/connection_postgres.md` |
| `materialize_connection_mysql` | MySQL | `docs/resources/connection_mysql.md` |
| `materialize_connection_sqlserver` | SQL Server | `docs/resources/connection_sqlserver.md` |
| `materialize_connection_aws` | AWS IAM (S3, DynamoDB, etc.) | `docs/resources/connection_aws.md` |
| `materialize_connection_ssh_tunnel` | SSH bastion | `docs/resources/connection_ssh_tunnel.md` |
| `materialize_connection_confluent_schema_registry` | Schema Registry | `docs/resources/connection_confluent_schema_registry.md` |
| `materialize_connection_aws_privatelink` | AWS PrivateLink | `docs/resources/connection_aws_privatelink.md` |
| `materialize_connection_iceberg_catalog` | Iceberg catalog | `docs/resources/connection_iceberg_catalog.md` |

**Connection example (Kafka with SASL):**
```hcl
resource "materialize_connection_kafka" "kafka" {
  name = "kafka_conn"

  kafka_broker {
    broker = "broker1.example.com:9092"
  }

  security_protocol = "SASL_SSL"
  sasl_mechanisms   = "SCRAM-SHA-256"

  sasl_username {
    text = "my_user"
  }

  sasl_password {
    secret {
      name = materialize_secret.kafka_password.name
    }
  }
}
```

All connections support: `database_name`, `schema_name`, `validate`, `comment`, `ownership_role`, `region`.

### Sources

| Resource | Purpose | Docs |
|----------|---------|------|
| `materialize_source_kafka` | Kafka source | `docs/resources/source_kafka.md` |
| `materialize_source_postgres` | PostgreSQL CDC | `docs/resources/source_postgres.md` |
| `materialize_source_mysql` | MySQL CDC | `docs/resources/source_mysql.md` |
| `materialize_source_sqlserver` | SQL Server CDC | `docs/resources/source_sqlserver.md` |
| `materialize_source_load_generator` | Test data | `docs/resources/source_load_generator.md` |

### Source Tables (Recommended)

Separate table definitions from source resources. This is the current recommended model.

| Resource | Purpose | Docs |
|----------|---------|------|
| `materialize_source_table_kafka` | Table from Kafka source | `docs/resources/source_table_kafka.md` |
| `materialize_source_table_postgres` | Table from PostgreSQL source | `docs/resources/source_table_postgres.md` |
| `materialize_source_table_mysql` | Table from MySQL source | `docs/resources/source_table_mysql.md` |
| `materialize_source_table_sqlserver` | Table from SQL Server source | `docs/resources/source_table_sqlserver.md` |
| `materialize_source_table_webhook` | Webhook table | `docs/resources/source_table_webhook.md` |

**Source + source table example (Kafka):**
```hcl
resource "materialize_source_kafka" "events" {
  name         = "events_source"
  cluster_name = materialize_cluster.analytics.name

  kafka_connection {
    name = materialize_connection_kafka.kafka.name
  }
}

resource "materialize_source_table_kafka" "events_table" {
  name  = "events"
  topic = "events-topic"

  source {
    name = materialize_source_kafka.events.name
  }

  format {
    avro {
      schema_registry_connection {
        name = materialize_connection_confluent_schema_registry.sr.name
      }
    }
  }

  envelope {
    upsert = true
  }
}
```

The migration guide at `docs/guides/materialize_source_table.md` explains how to move from inline table definitions to separate source table resources.

### Sinks

| Resource | Purpose | Docs |
|----------|---------|------|
| `materialize_sink_kafka` | Write to Kafka | `docs/resources/sink_kafka.md` |
| `materialize_sink_iceberg` | Write to Iceberg | `docs/resources/sink_iceberg.md` |

**Sink example (Kafka):**
```hcl
resource "materialize_sink_kafka" "output" {
  name         = "analytics_output"
  cluster_name = materialize_cluster.analytics.name
  topic        = "analytics-results"

  kafka_connection {
    name = materialize_connection_kafka.kafka.name
  }

  from {
    name = materialize_materialized_view.analytics.name
  }

  format {
    json = true
  }

  envelope {
    upsert = true
  }

  key = ["id"]
}
```

### Views and Indexes

| Resource | Purpose | Docs |
|----------|---------|------|
| `materialize_materialized_view` | Incrementally maintained view | `docs/resources/materialized_view.md` |
| `materialize_view` | Logical view (not materialized) | `docs/resources/view.md` |
| `materialize_index` | In-memory index on a view | `docs/resources/index.md` |
| `materialize_table` | Standard table | `docs/resources/table.md` |
| `materialize_type` | User-defined type | `docs/resources/type.md` |

**Materialized view example:**
```hcl
resource "materialize_materialized_view" "analytics" {
  name         = "order_totals"
  cluster_name = materialize_cluster.analytics.name

  statement = <<SQL
    SELECT customer_id, SUM(amount) AS total
    FROM orders
    GROUP BY customer_id
  SQL
}
```

### Security and Access Control

| Resource | Purpose | Docs |
|----------|---------|------|
| `materialize_role` | User/role management | `docs/resources/role.md` |
| `materialize_secret` | Credential storage | `docs/resources/secret.md` |
| `materialize_network_policy` | IP-based access rules | `docs/resources/network_policy.md` |
| `materialize_grant_system_privilege` | System-level grants | `docs/resources/grant_system_privilege.md` |

**Object-level grant resources** (all follow the same pattern):

`materialize_cluster_grant`, `materialize_database_grant`, `materialize_schema_grant`, `materialize_table_grant`, `materialize_view_grant`, `materialize_materialized_view_grant`, `materialize_source_grant`, `materialize_connection_grant`, `materialize_secret_grant`, `materialize_type_grant`, `materialize_role_grant`

Each grant resource requires `role_name`, `privilege`, and the object reference.

**Default privilege grants** (`materialize_*_grant_default_privilege`) set privileges on objects created in the future.

**RBAC example:**
```hcl
resource "materialize_role" "analyst" {
  name = "analyst"
}

resource "materialize_database_grant" "analyst_usage" {
  role_name     = materialize_role.analyst.name
  privilege     = "USAGE"
  database_name = "analytics"
}

resource "materialize_schema_grant" "analyst_usage" {
  role_name     = materialize_role.analyst.name
  privilege     = "USAGE"
  database_name = "analytics"
  schema_name   = "public"
}
```

**Network policy example:**
```hcl
resource "materialize_network_policy" "office" {
  name = "office_access"

  rule {
    name      = "office_cidr"
    action    = "allow"
    direction = "ingress"
    address   = "10.0.0.0/8"
  }
}
```

### Cloud-Only Resources (Frontegg-dependent)

These only work with Materialize Cloud, not self-managed:

| Resource | Purpose | Docs |
|----------|---------|------|
| `materialize_user` | Organization users | `docs/resources/user.md` |
| `materialize_app_password` | API access tokens | `docs/resources/app_password.md` |
| `materialize_sso_config` | SSO configuration | `docs/resources/sso_config.md` |
| `materialize_sso_domain` | SSO domain mapping | `docs/resources/sso_domain.md` |
| `materialize_sso_group_mapping` | SSO group to role mapping | `docs/resources/sso_group_mapping.md` |
| `materialize_sso_default_roles` | Default SSO roles | `docs/resources/sso_default_roles.md` |
| `materialize_scim_config` | SCIM integration | `docs/resources/scim_config.md` |
| `materialize_scim_group` | SCIM groups | `docs/resources/scim_group.md` |
| `materialize_region` | Region management | `docs/resources/region.md` |

### System Configuration

| Resource | Purpose | Docs |
|----------|---------|------|
| `materialize_system_parameter` | System-level config | `docs/resources/system_parameter.md` |
| `materialize_role_parameter` | Role-level config | `docs/resources/role_parameter.md` |

## Data Sources

Data sources return read-only lists filtered by optional parameters. All support `region` filtering.

| Data Source | Additional Filters | Docs |
|-------------|-------------------|------|
| `materialize_cluster` | | `docs/data-sources/materialize_cluster.md` |
| `materialize_connection` | `database_name`, `schema_name` | `docs/data-sources/materialize_connection.md` |
| `materialize_database` | `database_name` | `docs/data-sources/materialize_database.md` |
| `materialize_schema` | `database_name` | `docs/data-sources/materialize_schema.md` |
| `materialize_source` | `database_name`, `schema_name` | `docs/data-sources/materialize_source.md` |
| `materialize_sink` | `database_name`, `schema_name` | `docs/data-sources/materialize_sink.md` |
| `materialize_materialized_view` | `database_name`, `schema_name` | `docs/data-sources/materialize_materialized_view.md` |
| `materialize_view` | `database_name`, `schema_name` | `docs/data-sources/materialize_view.md` |
| `materialize_table` | `database_name`, `schema_name` | `docs/data-sources/materialize_table.md` |
| `materialize_index` | `database_name`, `schema_name` | `docs/data-sources/materialize_index.md` |
| `materialize_role` | | `docs/data-sources/materialize_role.md` |
| `materialize_secret` | `database_name`, `schema_name` | `docs/data-sources/materialize_secret.md` |
| `materialize_egress_ips` | | `docs/data-sources/materialize_egress_ips.md` |
| `materialize_region` | | `docs/data-sources/materialize_region.md` |
| `materialize_network_policy` | | `docs/data-sources/materialize_network_policy.md` |
| `materialize_system_parameter` | | `docs/data-sources/materialize_system_parameter.md` |
| `materialize_current_cluster` | | `docs/data-sources/materialize_current_cluster.md` |
| `materialize_current_database` | | `docs/data-sources/materialize_current_database.md` |

## Common Patterns

### Qualified SQL Names

Most resources return a read-only `qualified_sql_name` in the format `database.schema.object_name`. When referencing objects across resources, use `database_name` and `schema_name` arguments (both default to `"materialize"` and `"public"` respectively).

### Secret and Credential References

Many connection arguments accept either inline text or a secret reference:

```hcl
# Inline text
sasl_username {
  text = "my_user"
}

# Secret reference
sasl_username {
  secret {
    name          = materialize_secret.username.name
    database_name = "materialize"
    schema_name   = "public"
  }
}
```

### Connection References in Sources/Sinks

```hcl
kafka_connection {
  name          = materialize_connection_kafka.my_conn.name
  database_name = "materialize"
  schema_name   = "public"
}
```

### Identify by Name

`materialize_cluster` and `materialize_schema` support `identify_by_name = true`, which uses the object name as the Terraform state ID instead of the internal Materialize ID. This is useful for blue/green deployments where you want to swap clusters without changing Terraform state.

### Write-Only Arguments (Terraform 1.11+)

Some sensitive fields support write-only ephemeral values that never appear in state:
- `materialize_role`: `password_wo` / `password_wo_version`
- `materialize_secret`: `value_wo` / `value_wo_version`

## Importing Existing Resources

```bash
terraform import materialize_cluster.my_cluster <region>:<cluster_id>
# or with identify_by_name:
terraform import materialize_cluster.my_cluster <region>:name:<cluster_name>
```

Find resource IDs in `mz_catalog` system tables:

| Object | System Table |
|--------|-------------|
| Clusters | `mz_catalog.mz_clusters` |
| Databases | `mz_catalog.mz_databases` |
| Schemas | `mz_catalog.mz_schemas` |
| Sources | `mz_catalog.mz_sources` |
| Sinks | `mz_catalog.mz_sinks` |
| Views | `mz_catalog.mz_views` |
| Connections | `mz_catalog.mz_connections` |
| Secrets | `mz_catalog.mz_secrets` |
| Roles | `mz_catalog.mz_roles` |

## Common Gotchas

- **Cloud vs self-managed**: Frontegg-dependent resources (users, SSO, SCIM, app passwords) only work with Materialize Cloud. Using them against a self-managed instance will fail.
- **Database without public schema**: Creating a `materialize_database` does not auto-create a `public` schema. Create one explicitly if needed.
- **Source table migration**: The inline `table {}` block in source resources is deprecated. Use separate `materialize_source_table_*` resources instead. See `docs/guides/materialize_source_table.md`.
- **Webhook sources**: `materialize_source_webhook` is legacy. New webhooks should use `materialize_table` with webhook support, though automated migration is not yet available.
- **Cluster replicas deprecated**: `materialize_cluster_replica` is deprecated. Use `materialize_cluster` with `size` for managed clusters.
- **Sensitive values in state**: Passwords and secrets marked `Sensitive` won't show in plan output but are stored in Terraform state. Use write-only arguments (`*_wo`) on Terraform 1.11+ to keep them out of state entirely.
