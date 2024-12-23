# Changelog

## 0.8.12 - 2024-12-20

## Features

* Support self-managed instances [#674](https://github.com/MaterializeInc/terraform-provider-materialize/pull/674).
    Now users can configure the provider like this:
    ```hcl
    # Self-managed configuration
    provider "materialize" {
      host     = "localhost"      # Required for self-managed deployments
      port     = 6875             # Optional, defaults to 6875
      database = "materialize"    # Optional, defaults to materialize
      username = "materialize"    # Optional, defaults to materialize
      password = ""               # Optional
      sslmode  = "disable"        # Optional, defaults to require
    }

    # SaaS configuration (unchanged)
    provider "materialize" {
      password       = "materialize_password"
      default_region = "aws/us-east-1"
    }
    ```

## Bug Fixes

* Fix intermittent test failures [#684](https://github.com/MaterializeInc/terraform-provider-materialize/pull/684)
* `materialize_sink_kafka` resource: sort topic config map keys for consistent SQL generation [#677](https://github.com/MaterializeInc/terraform-provider-materialize/pull/677)

## Misc

* Remove outdated feature lifecycle annotations for features that are now Generally Available (GA) [#679](https://github.com/MaterializeInc/terraform-provider-materialize/pull/679)
* Update Redpanda image reference: [#681](https://github.com/MaterializeInc/terraform-provider-materialize/pull/681)
* Routine dependency updates: [#672](https://github.com/MaterializeInc/terraform-provider-materialize/pull/672), [#673](https://github.com/MaterializeInc/terraform-provider-materialize/pull/673), [#678](https://github.com/MaterializeInc/terraform-provider-materialize/pull/678), [#680](https://github.com/MaterializeInc/terraform-provider-materialize/pull/680), [#683](https://github.com/MaterializeInc/terraform-provider-materialize/pull/683)

## 0.8.11 - 2024-11-13

## Features

* Adding a new `materialize_network_policy` resource and data source [#669](https://github.com/MaterializeInc/terraform-provider-materialize/pull/669).

  A network policy allows you to manage access to the system through IP-based rules.

  * Example `materialize_network_policy` resource:

    ```hcl
    resource "materialize_network_policy" "office_policy" {
      name = "office_access_policy"

      rule {
        name      = "new_york"
        action    = "allow"
        direction = "ingress"
        address   = "8.2.3.4/28"
      }

      rule {
        name      = "minnesota"
        action    = "allow"
        direction = "ingress"
        address   = "2.3.4.5/32"
      }

      comment = "Network policy for office locations"
    }
    ```

  * Example `materialize_network_policy` data source:

    ```hcl
    data "materialize_network_policy" "all" {}
    ```

  * Added support for the new `CREATENETWORKPOLICY` system privilege:

    ```hcl
    resource "materialize_role" "test" {
      name = "test_role"
    }

    resource "materialize_grant_system_privilege" "role_createnetworkpolicy" {
      role_name = materialize_role.test.name
      privilege = "CREATENETWORKPOLICY"
    }
    ```

  * An initial `default` network policy will be created.
    This policy allows open access to the environment and can be altered by a `superuser`.
    Use the `ALTER SYSTEM SET network_policy TO 'office_access_policy'` command
    or the `materialize_system_parameter` resource to update the default network policy.

    ```hcl
    resource "materialize_system_parameter" "system_parameter" {
      name  = "network_policy"
      value = "office_access_policy"
    }
    ```

## Bug Fixes

* Updated the cluster and cluster replica query builders to skip `DISK` property for `cc` and `C` clusters as this is enabled by default for those sizes [#671](https://github.com/MaterializeInc/terraform-provider-materialize/pull/671)

## Misc

* Upgrade from `pgx` v3 to v4 [#663](https://github.com/MaterializeInc/terraform-provider-materialize/pull/663)
* Routine dependency updates: [#668](https://github.com/MaterializeInc/terraform-provider-materialize/pull/668), [#667](https://github.com/MaterializeInc/terraform-provider-materialize/pull/667)
* Upgraded Go version from `1.22.0` to `1.22.7` for improved performance and security fixes [#669](https://github.com/MaterializeInc/terraform-provider-materialize/pull/669)
* Added `--bootstrap-builtin-analytics-cluster-replica-size` to the Docker compose file to fix failing tests [#671](https://github.com/MaterializeInc/terraform-provider-materialize/pull/671)

## 0.8.10 - 2024-10-7

## Features

* Add support for `partition_by` attribute in `materialize_sink_kafka` [#659](https://github.com/MaterializeInc/terraform-provider-materialize/pull/659)
  * The `partition_by` attribute accepts a SQL expression used to partition the data in the Kafka sink. Can only be used with `ENVELOPE UPSERT`.
  * Example usage:
  ```hcl
  resource "materialize_sink_kafka" "orders_kafka_sink" {
    name         = "orders_sink"
    kafka_connection {
      name = "kafka_connection"
    }
    topic        = "orders_topic"

    partition_by = "column_name"

    # Additional configuration...
  }
  ```

## Misc

* Set `transaction_isolation` as conneciton option instead of executing a `SET` command [#660](https://github.com/MaterializeInc/terraform-provider-materialize/pull/660)
* Routine dependency updates: [#661](https://github.com/MaterializeInc/terraform-provider-materialize/pull/661)

## 0.8.9 - 2024-09-30

### BugFixes

* Explicitly set `TRANSACTION_ISOLATION` to `STRICT SERIALIZABLE` [#657](https://github.com/MaterializeInc/terraform-provider-materialize/pull/657)
* Fix user not found state status in the `materialize_user` resource [#638](https://github.com/MaterializeInc/terraform-provider-materialize/pull/638)
* Fix Inconsistent Error Handling in `ReadUser` in the `materialize_user` resource [#642](https://github.com/MaterializeInc/terraform-provider-materialize/pull/642)

### Misc

* Update Go version to 1.22 [#650](https://github.com/MaterializeInc/terraform-provider-materialize/pull/650)
* Switched tests to use a stable version of the Rust Frontegg mock service [#653](https://github.com/MaterializeInc/terraform-provider-materialize/pull/653)
* Improve the Cloud Mock Service [#651](https://github.com/MaterializeInc/terraform-provider-materialize/pull/651)
* Disable telemetry in CI [#640](https://github.com/MaterializeInc/terraform-provider-materialize/pull/640)

## 0.8.8 - 2024-08-26

### Features

* Add `wait_until_ready` option to `cluster` resources, which allows graceful cluster reconfiguration (i.e., with no downtime) for clusters with no sources or sinks. [#632](https://github.com/MaterializeInc/terraform-provider-materialize/pull/632)
  * Example usage:
  ```hcl
  resource "materialize_cluster" "cluster" {
    name = var.mz_cluster
    size = "25cc"
    wait_until_ready {
      enabled = true
      timeout = "10m"
      on_timeout = "COMMIT"
    }
  }
  ```

### Misc

* Unify the cluster alter commands [#628](https://github.com/MaterializeInc/terraform-provider-materialize/pull/628)
* Switched tests to use the Rust Frontegg mock service [#634](https://github.com/MaterializeInc/terraform-provider-materialize/pull/634)


## 0.8.7 - 2024-08-15

### Features

* Add support for AWS IAM authentication in `materialize_connection_kafka` [#627](https://github.com/MaterializeInc/terraform-provider-materialize/pull/627)
  * Example usage:
  ```hcl
  # Create an AWS connection for IAM authentication
  resource "materialize_connection_aws" "msk_auth" {
    name                    = "aws_msk"
    assume_role_arn         = "arn:aws:iam::123456789012:role/MaterializeMSK"
  }

  # Create a Kafka connection using AWS IAM authentication
  resource "materialize_connection_kafka" "kafka_msk" {
    name              = "kafka_msk"
    kafka_broker {
      broker = "b-1.your-cluster-name.abcdef.c1.kafka.us-east-1.amazonaws.com:9098"
    }
    security_protocol = "SASL_SSL"
    aws_connection {
      name          = materialize_connection_aws.msk_auth.name
      database_name = materialize_connection_aws.msk_auth.database_name
      schema_name   = materialize_connection_aws.msk_auth.schema_name
    }
  }
  ```

### Bug Fixes

* Fix `materialize_connection_aws` read function issues caused by empty internal table [#630](https://github.com/MaterializeInc/terraform-provider-materialize/pull/630)
* Fix duplicate application name in the provider configuration [#626](https://github.com/MaterializeInc/terraform-provider-materialize/pull/626)

### Misc

* Add more information to import docs for all resources [#623](https://github.com/MaterializeInc/terraform-provider-materialize/pull/623)
* Routine dependency updates: [#631](https://github.com/MaterializeInc/terraform-provider-materialize/pull/631)

## 0.8.6 - 2024-07-31

### Features

* Add new `identify_by_name` option for `materialize_cluster` resource [#618](https://github.com/MaterializeInc/terraform-provider-materialize/pull/618)
  * When set to `true`, the cluster name is used as the Terraform resource ID instead of the internal cluster ID
  * This eliminates the need to update the Terraform state file if a cluster is recreated with the same name but a different ID outside of Terraform (e.g., via the Materialize UI or dbt)
  * The resource now uses the format `region:type:value` for IDs, where type is either "name" or "id"

  Example usage:
  ```hcl
  resource "materialize_cluster" "test_name_as_id" {
    name               = "test_name_as_id"
    size               = "25cc"
    replication_factor = "1"
    identify_by_name   = true # Set to true to use the cluster name as the resource ID
  }
  ```

  Existing `materialize_cluster` resources will be automatically migrated to the new ID format. To use this new feature:

  1. Update your Terraform configuration to `v0.8.6` or later of the Materialize provider.
  1. Run `terraform init` to download the new provider version.
  1. Run `terraform plan` and `terraform refresh` to verify the changes.

  No manual intervention is required for existing resources, but reviewing the plan is recommended to ensure the expected updates are made.

### Misc

* Add `materialize_source_mysql` tests for the `ignore_columns` attribute [#616](https://github.com/MaterializeInc/terraform-provider-materialize/pull/616)
* Extend integration tests to run create resources in two regions different [#614](https://github.com/MaterializeInc/terraform-provider-materialize/pull/614)

## 0.8.5 - 2024-07-22

### Features

* Change cluster option from `REHYDRATION TIME ESTIMATE` to `HYDRATION TIME ESTIMATE` [#603](https://github.com/MaterializeInc/terraform-provider-materialize/pull/603)
* Add support for upsert options error decoding alias [#612](https://github.com/MaterializeInc/terraform-provider-materialize/pull/612):
  ```hcl
  envelope {
    upsert = true
    upsert_options {
      value_decoding_errors {
        inline {
          enabled = true
          alias   = "my_error_col"
        }
      }
    }
  }
  ```

### Misc

* Update Terraform docs examples [#613](https://github.com/MaterializeInc/terraform-provider-materialize/pull/613)

## 0.8.4 - 2024-07-15

### Features

* Add support for adding and removing subsources for the `materialize_source_mysql` resource [#604](https://github.com/MaterializeInc/terraform-provider-materialize/pull/604)
* Allow the `roles` attribute for the `materialize_user` resource to be updated without `forceNew` [#610](https://github.com/MaterializeInc/terraform-provider-materialize/pull/610)
* Add `key_compatibility_level` and `value_compatibility_level` attributes to the
  `materialize_sink_kafka` resource [#600](https://github.com/MaterializeInc/terraform-provider-materialize/pull/600)
* Add `progress_topic_replication_factor` attribute to the `materialize_connection_kafka` resource [#598](https://github.com/MaterializeInc/terraform-provider-materialize/pull/598)
* Add topics options to the `materialize_sink_kafka` resource [#597](https://github.com/MaterializeInc/terraform-provider-materialize/pull/597).
  See the [Kafka documentation](https://kafka.apache.org/documentation/#topicconfigs) for available configs. For example:
  ```hcl
  topic_replication_factor = 1
  topic_partition_count    = 6
  topic_config = {
    "cleanup.policy" = "compact"
    "retention.ms"   = "86400000"
  }
  ```

### Bug Fixes

* Fix a bug in the `materialize_source_kafka` resource where the value format JSON was not processed correctly [#607](https://github.com/MaterializeInc/terraform-provider-materialize/pull/607)

### Misc

* Fix CI intermittent failing tests [#609](https://github.com/MaterializeInc/terraform-provider-materialize/pull/609)
* Fix connections data source tests [#594](https://github.com/MaterializeInc/terraform-provider-materialize/pull/594)
* Routine dependency updates: [#608](https://github.com/MaterializeInc/terraform-provider-materialize/pull/608)

## 0.8.3 - 2024-07-08

### Features

* New data source: `materialize_user` [#592](https://github.com/MaterializeInc/terraform-provider-materialize/pull/592).
  This data source allows retrieving information about existing users in an organization by email.
  It can be used together with the `materialize_user` resource for importing existing users:

  ```hcl
  # Retrieve an existing user by email
  data "materialize_user" "example" {
    email = "existing-user@example.com"
  }
  output "user_id" {
    value = data.materialize_user.example.id
  }
  ```

  Define the `materialize_user` resource and import the existing user:
  ```hcl
  resource "materialize_user" "example" {
    email = "existing-user@example.com"
  }
  ```

  Import command:
  ```sh
  # terraform import materialize_user.example ${data.materialize_user.example.id}
  ```

### Misc

* Refactor to only fetch the list of Frontegg roles once per Terraform provider invocation [#595](https://github.com/MaterializeInc/terraform-provider-materialize/pull/595).
* Improve Frontegg HTTP mock server to improve maintainability [#593](https://github.com/MaterializeInc/terraform-provider-materialize/pull/593).

## 0.8.2 - 2024-07-01

### Features

* Resource Kafka Sink: Add `headers` attribute [#569](https://github.com/MaterializeInc/terraform-provider-materialize/pull/569)
* Resource Kafka Source: Add upsert options `value_decoding_errors` attribute [#586](https://github.com/MaterializeInc/terraform-provider-materialize/pull/586)

### Misc

* Rename `BOOTSTRAP_BUILTIN_INTROSPECTION_CLUSTER_REPLICA_SIZE` [#584](https://github.com/MaterializeInc/terraform-provider-materialize/pull/584)

## 0.8.1 - 2024-06-20

### Features

* Allow creating service app passwords, which are app passwords that are
  associated with a user without an email address.

  For example, here's how you might provision an app password for a production
  dashboard application that should have the `USAGE` privilege on the
  `production_analytics` database:

  ```hcl
  resource "materialize_role" "production_dashboard" {
    name = "svc_production_dashboard"
  }

  resource "materialize_app_password" "production_dashboard_app_password" {
    name = "production_dashboard_app_password"
    type = "service"
    user = materialize_role.production_dashboard.name
    roles = ["Member"]
  }

  resource "materialize_database_grant" "database_grant_usage" {
    role_name     = materialize_role.production_dashboard.name
    privilege     = "USAGE"
    database_name = "production_analytics"
  }
  ```

* Allow skipping activation emails when creating users [#573](https://github.com/MaterializeInc/terraform-provider-materialize/pull/573)
* Allow `resource_sink_kafka` `FROM` attribute updates [#578](https://github.com/MaterializeInc/terraform-provider-materialize/pull/578)

### Misc

* Routine dependency updates: [#577](https://github.com/MaterializeInc/terraform-provider-materialize/pull/577), [#568](https://github.com/MaterializeInc/terraform-provider-materialize/pull/568), [#567](https://github.com/MaterializeInc/terraform-provider-materialize/pull/567)

## 0.8.0 - 2024-05-16

### Breaking Changes
* This release introduces a breaking change to the `materialize_source_postgres` resources configuration: [#487](https://github.com/MaterializeInc/terraform-provider-materialize/pull/487)
  * **The `schema` property is removed**: The `schema` property is removed from the `materialize_source_postgres` resource configuration. Users must now explicitly define the `table` block to specify the tables to include in the source. This change is designed to ensure consistency and predictability in the Terraform provider's behavior.
  * **The `table` block is now required**: Previously, the `table` block was optional, allowing users to specify specific tables to include in the source. Starting with version `v0.8.0`, the `table` block is now required. Users must explicitly define the tables to be included in the source. This change is designed to ensure consistency and predictability in the Terraform provider's behavior.
  * **Changes to the `table` block**: The `tables` property schema has been updated as follows:

  ```hcl
    table {
      upstream_name        = string # Required: The name of the table in the upstream database: Previously `name`
      upstream_schema_name = string # The schema of the table in the upstream database
      name                 = string # The name of the table in Materialize: Previously `alias`
      schema_name          = string # The schema of the table in Materialize
      datatabase_name      = string # The name of the database where the table will be created in Materialize
    }
  ```

  * **Migration Guide**: For a detailed guide on adapting to these changes, refer to the migration guide [here](https://github.com/MaterializeInc/terraform-provider-materialize/pull/487)

* The `subsource` read-only attribute is removed from all source resources as part of a change to align with Materialize's internal behavior.

## Misc
* Routine dependency updates: [#564](https://github.com/MaterializeInc/terraform-provider-materialize/pull/564)

## 0.7.1 - 2024-04-30

### Features
* Update `region` attribute for all resources to be `computed` [#559](https://github.com/MaterializeInc/terraform-provider-materialize/pull/559)
* Add `connection_id` attribute for `materialize_connection` data source [#553](https://github.com/MaterializeInc/terraform-provider-materialize/pull/553)

### Bug Fixes
* Check for `nil` values in `GetSliceValueString` [#552](https://github.com/MaterializeInc/terraform-provider-materialize/pull/552)
* Fix `materialize_connection_kafka` rename race condition [#561](https://github.com/MaterializeInc/terraform-provider-materialize/pull/561)

### Misc
* Routine dependency updates: [#557](https://github.com/MaterializeInc/terraform-provider-materialize/pull/557), [#558](https://github.com/MaterializeInc/terraform-provider-materialize/pull/558)

## 0.7.0 - 2024-04-24

### Breaking Changes
* **`public` schemas are no longer created by default**: In previous versions, the `materialize_database` resource automatically created a `public` schema in each new database, mimicking traditional SQL database behavior. Starting with version `v0.7.0`, this default behavior has been removed. Users must now explicitly define and manage `public` schemas within their Terraform configurations. This change is designed to align the Terraform provider's behavior more closely with its design principles, ensuring consistency and predictability.
    * **Action Required**: Explicitly define `public` schemas in your Terraform configurations if needed. Along with the required grant `USAGE` to the `PUBLIC` pseudo-role for the public schema
    * **Migration Guide**: This only affects newly created databases. Details on adapting to this change are available [here](https://github.com/MaterializeInc/terraform-provider-materialize/pull/546)

### Features
* Add scheduling attribute to the `materialize_cluster` resource [#545](https://github.com/MaterializeInc/terraform-provider-materialize/pull/545)

### Bug Fixes
* Fix an issue where resource imports were failing when using a non-default region [#550](https://github.com/MaterializeInc/terraform-provider-materialize/pull/550)

### Misc.
* Routine dependency updates [#549](https://github.com/MaterializeInc/terraform-provider-materialize/pull/549)

## 0.6.10 - 2024-04-19

### Features
* Allow `ALTER CONNECTION` updates for the following connection resources:
  * `materialize_connection_mysql` resource (#541)
  * `materialize_connection_confluent_schema_registry` resource (#540)
  * `materialize_connection_kafka` resource (#538)
  * `materialize_connection_aws_privatelink` resource (#533)
  * `materialize_connection_aws` resource (#529)

* Add support for Frontegg SCIM groups which includes the following new resources (#525):
  * `materialize_scim_group`
  * `materialize_scim_group_roles`
  * `materialize_scim_group_users`

* Add support for the key value load generator source (#537)
* New `materialize_region` resource (#535)
* Add `validate` parameter to `materialize_aws_privatelink` connection (#539)
* Remove `idle_arrangement_merge_effort` option from `materialize_cluster` (#532)

### Misc
* Add additional `materialize_connection_postgres` unit tests (#542)
* Define builtin probe cluster size (#536)
* Remove unnecessary SSH connections in Postgres tests (#531)
* Refactor the Frontegg package (#530)
* Use unique SSH conn name to fix CI flakes (#527)

## 0.6.9 - 2024-03-29

### Features
* Add `PUBLIC` pseudo-role to resource grants [#524](https://github.com/MaterializeInc/terraform-provider-materialize/pull/524)
* Add `materialize_role_parameter` resource [#522](https://github.com/MaterializeInc/terraform-provider-materialize/pull/522)
* Allow `ALTER CONNECTION` updates for `resource_connection_postgres` [#511](https://github.com/MaterializeInc/terraform-provider-materialize/pull/511)
* Allow `ALTER CONNECTION` updates for SSH tunnels [#523](https://github.com/MaterializeInc/terraform-provider-materialize/pull/523)

### BugFixes
* Fix failing environmentd bootstrap [#512](https://github.com/MaterializeInc/terraform-provider-materialize/pull/512)

### Misc

* Added acceptance tests for:
  * Cloud region data source [#521](https://github.com/MaterializeInc/terraform-provider-materialize/pull/521)
  * SCIM config resource [#520](https://github.com/MaterializeInc/terraform-provider-materialize/pull/520)
  * SCIM Groups data source [#509](https://github.com/MaterializeInc/terraform-provider-materialize/pull/509)
  * SSO config data source [#507](https://github.com/MaterializeInc/terraform-provider-materialize/pull/507)

## 0.6.8 - 2024-03-20

### Features
* Allow `region` option for data sources [#506](https://github.com/MaterializeInc/terraform-provider-materialize/pull/506)
* Remove scale factor for auction/counter load generator sources [#502](https://github.com/MaterializeInc/terraform-provider-materialize/pull/502)
* Add cluster availability zone attribute [#498](https://github.com/MaterializeInc/terraform-provider-materialize/pull/498)

### Misc
* Added acceptance tests for SSO group mapping resource [#505](https://github.com/MaterializeInc/terraform-provider-materialize/pull/505)
* Added acceptance tests for SSO default roles resource [#503](https://github.com/MaterializeInc/terraform-provider-materialize/pull/503)
* Added acceptance tests for SSO domain resource [#497](https://github.com/MaterializeInc/terraform-provider-materialize/pull/497)

## 0.6.7 - 2024-03-11

### Features
* New resource `materialize_connection_mysql` [#480](https://github.com/MaterializeInc/terraform-provider-materialize/pull/480)
* New resource `materialize_source_mysql` [#486](https://github.com/MaterializeInc/terraform-provider-materialize/pull/486)
* New resource `materialize_connection_aws` [#492](https://github.com/MaterializeInc/terraform-provider-materialize/pull/492)

### BugFixes
* Add region prefix hint message [#488](https://github.com/MaterializeInc/terraform-provider-materialize/pull/488)

### Misc
* Acceptance tests for SCIM config resource [#483](https://github.com/MaterializeInc/terraform-provider-materialize/pull/483)
* Acceptance tests for SCIM config data source [#485](https://github.com/MaterializeInc/terraform-provider-materialize/pull/485)
* Dependency updates: [#484](https://github.com/MaterializeInc/terraform-provider-materialize/pull/484), [#494](https://github.com/MaterializeInc/terraform-provider-materialize/pull/494), [#495](https://github.com/MaterializeInc/terraform-provider-materialize/pull/495)

## 0.6.6 - 2024-02-29

### Features
* Allow single system parameter data sources [#474](https://github.com/MaterializeInc/terraform-provider-materialize/pull/474)
* Allow schema rename [#450](https://github.com/MaterializeInc/terraform-provider-materialize/pull/450)

### Misc
* Acceptance tests for user resource [#477](https://github.com/MaterializeInc/terraform-provider-materialize/pull/477)
* Dependency updates: [#476](https://github.com/MaterializeInc/terraform-provider-materialize/pull/476)

## 0.6.5 - 2024-02-19

### Features
* Add support for top level PrivateLink connections to Kafka [#471](https://github.com/MaterializeInc/terraform-provider-materialize/pull/471)

## 0.6.4 - 2024-02-14

### Features
* New resource: `materialize_system_parameter` [#464](https://github.com/MaterializeInc/terraform-provider-materialize/pull/464)
* New data source: `materialize_system_parameter` [#464](https://github.com/MaterializeInc/terraform-provider-materialize/pull/464)
* Add the new `cc` cluster sizes [#467](https://github.com/MaterializeInc/terraform-provider-materialize/pull/467)

### Misc
* Mark the disk option as deprecated [#460](https://github.com/MaterializeInc/terraform-provider-materialize/pull/460)
* Dependabot updates: [#459](https://github.com/MaterializeInc/terraform-provider-materialize/pull/459)

## 0.6.3 - 2024-02-02

### BugFixes
* Fix the SSO configuration SP Entity ID definition [#456](https://github.com/MaterializeInc/terraform-provider-materialize/pull/456)

## 0.6.2 - 2024-02-02

### BugFixes
* Fix the SSO configuration SP Entity ID definition [#456](https://github.com/MaterializeInc/terraform-provider-materialize/pull/456)

## 0.6.1 - 2024-02-01

### Features
* New resource: `materialize_scim_config` [#449](https://github.com/MaterializeInc/terraform-provider-materialize/pull/449)

### BugFixes
* Allow imports for SSO resources [#453](https://github.com/MaterializeInc/terraform-provider-materialize/pull/453)
* Fix failing weekly tests [#448](https://github.com/MaterializeInc/terraform-provider-materialize/pull/448)

### Misc
* Dependabot updates: [#447](https://github.com/MaterializeInc/terraform-provider-materialize/pull/447)
* Documentation updates: [#446](https://github.com/MaterializeInc/terraform-provider-materialize/pull/446), [#453](https://github.com/MaterializeInc/terraform-provider-materialize/pull/453)

## 0.6.0 - 2024-01-23

### Breaking Changes
* Drop `SIZE` support for sources and sinks [#438](https://github.com/MaterializeInc/terraform-provider-materialize/pull/438)

### Features
* Make `cluster_name` parameter required for `materialized_view` and `index` resources [#435](https://github.com/MaterializeInc/terraform-provider-materialize/pull/435)
* Include `create_sql` for `view` and `materialized_view` [#436](https://github.com/MaterializeInc/terraform-provider-materialize/pull/436)
* New resources: [#442](https://github.com/MaterializeInc/terraform-provider-materialize/pull/442)
  * `materialize_sso_config`: Manages [SSO configuration](https://materialize.com/docs/manage/access-control/sso/) details
  * `materialize_sso_default_roles`: Manages SSO default roles
  * `materialize_sso_domain`: Manages SSO domains
  * `materialize_sso_group_mappings`: Manages SSO group mappings
* New data sources:
  * `materialize_scim_configs`: Fetches SCIM configuration details
  * `materialize_scim_groups`: Fetches SCIM group details
  * `materialize_sso_config`: Fetches SSO configuration details

### Misc
* Add Tests for `INCLUDE KEY AS` for kafka sources [#439](https://github.com/MaterializeInc/terraform-provider-materialize/pull/439)
* Mark comments as public preview [#440](https://github.com/MaterializeInc/terraform-provider-materialize/pull/440)
* Dependabot updates: [#441](https://github.com/MaterializeInc/terraform-provider-materialize/pull/441), [#443](https://github.com/MaterializeInc/terraform-provider-materialize/pull/443)

## 0.5.0 - 2024-01-10

### Features
* Introduced a unified interface for managing both global and regional resources.
* Implemented single authentication using an app password for all operations.
* Added dynamic client allocation for managing different resource types.
* Enhanced provider configuration with parameters for default settings and optional endpoint overrides.
* New resources:
  * App passwords: `materialize_app_password`.
  * User management `materialize_user`.
* Added data sources for fetching region details (`materialize_region`).
* Implemented support for establishing SQL connections across multiple regions.
* Introduced a new `region` parameter in all resource and data source configurations. This allows users to specify the region for resource creation and data retrieval.

### Breaking Changes
* **Provider Configuration Changes**:
  * Deprecated the `host`, `port`, and `user` parameters in the provider configuration. These details are now derived from the app password.
  * Retained only the `password` definition in the provider configuration. This password is used to fetch all necessary connection information.
* **New `region` Configuration**:
  * Introduced a new `default_region` parameter in the provider configuration. This allows users to specify the default region for resource creation.
  * The `default_region` parameter can be overridden in specific resource configurations if a particular resource needs to be created in a non-default region.

  ```hcl
  provider "materialize" {
    password       = var.materialize_app_password
    default_region = "aws/us-east-1"
  }

  resource "materialize_cluster" "cluster" {
    name   = "cluster"
    region = "aws/us-west-2"
  }
  ```

### Misc
* Mock Services for Testing:
  * Added a new mocks directory, which includes mock services for the Cloud API and the FrontEgg API.
  * These mocks are intended for local testing and CI, facilitating development and testing without the need for a live backend.

### Migration Guide
* Before upgrading to `v0.5.0`, users should ensure that they have upgraded to `v0.4.x` which introduced the Terraform state migration necessary for `v0.5.0`. After upgrading to `v0.4.x`, users should run `terraform plan` to ensure that the state migration has completed successfully.
* Users upgrading to `v0.5.0` should update their provider configurations to remove the `host`, `port`, and `user` parameters and ensure that the `password` parameter is set with the app password.
* For managing resources across multiple regions, users should specify the `default_region` parameter in their provider configuration or override it in specific resource blocks as needed using the `region` parameter.

## 0.4.3 - 2024-01-08

### Breaking Changes
* Rename "default" cluster to "quickstart" as part of a change on the Materialize side [#423](https://github.com/MaterializeInc/terraform-provider-materialize/pull/423)

### Misc
* Dependabot updates [#427](https://github.com/MaterializeInc/terraform-provider-materialize/pull/427)
* Change the user of the scheduled tests [#428](https://github.com/MaterializeInc/terraform-provider-materialize/pull/428)

## 0.4.2 - 2024-01-02

### Features
* Add `COMPRESSION TYPE`` option to `materialize_sink_kafka` resource [#414](https://github.com/MaterializeInc/terraform-provider-materialize/pull/414)

### BugFixes
* Fix Kafka offset acceptance tests [#418](https://github.com/MaterializeInc/terraform-provider-materialize/pull/418)

### Misc
* Include additional tests for subsource [#413](https://github.com/MaterializeInc/terraform-provider-materialize/pull/413)
* Dependabot updates [#416](https://github.com/MaterializeInc/terraform-provider-materialize/pull/416), [#415](https://github.com/MaterializeInc/terraform-provider-materialize/pull/415)

## 0.4.1 - 2023-12-12

### Features
* Allow Avro comments (`avro_doc_type` and `avro_doc_column`) for resource `materialize_sink_kafka` [#373](https://github.com/MaterializeInc/terraform-provider-materialize/pull/373)

### Misc
* Include additional acceptance tests for datasources [#410](https://github.com/MaterializeInc/terraform-provider-materialize/issues/410)

## 0.4.0 - 2023-12-12

### Features
* Improved ID structuring in Terraform state file with region-prefixed IDs, enhancing state management to allow supporting new features like managing cloud resources ([#400](https://github.com/MaterializeInc/terraform-provider-materialize/issues/400), [#401](https://github.com/MaterializeInc/terraform-provider-materialize/issues/401), [#402](https://github.com/MaterializeInc/terraform-provider-materialize/issues/402), and [#406](https://github.com/MaterializeInc/terraform-provider-materialize/issues/406))

## 0.3.4 - 2023-12-06

### Features
* Add `ssh_tunnel` as a broker level attribute for `materialize_connection_kafka`. `ssh_tunnel` can be applied as a top level attribute (the default for all brokers) or both the individual broker level [#366](https://github.com/MaterializeInc/terraform-provider-materialize/pull/366)

### BugFixes
* Allow `PUBLIC` as `grantee` for default grant resources [#397](https://github.com/MaterializeInc/terraform-provider-materialize/issues/397)

## 0.3.3 - 2023-11-30

### Features
* Add `default` to columns when defining a `materialize_table` [#374](https://github.com/MaterializeInc/terraform-provider-materialize/pull/374)
* Add `expose_progress` to `materialize_source_load_generator` [#374](https://github.com/MaterializeInc/terraform-provider-materialize/pull/374)
* Support [row type](https://materialize.com/docs/sql/create-type/#row-properties) in `materialize_type` [#374](https://github.com/MaterializeInc/terraform-provider-materialize/pull/374)

### BugFixes
* Fix `expose_progress` in `materialize_source_postgres` and `materialize_source_kafka` [#374](https://github.com/MaterializeInc/terraform-provider-materialize/pull/374)
* Fix `start_offset` in `materialize_source_kafka` [#374](https://github.com/MaterializeInc/terraform-provider-materialize/pull/374)
* Allow `replication_factor` of 0 for `materialize_cluster` [#390](https://github.com/MaterializeInc/terraform-provider-materialize/pull/390)

### Misc
* Set `replication_factor` as computed in `materialize_cluster` [#374](https://github.com/MaterializeInc/terraform-provider-materialize/pull/374)

### Breaking Changes
* Remove `session_variables` from `materialize_role` [#374](https://github.com/MaterializeInc/terraform-provider-materialize/pull/374)

## 0.3.2 - 2023-11-24

### Features

### BugFixes
* Fix default grant read [#381](https://github.com/MaterializeInc/terraform-provider-materialize/pull/381)

### Misc

## 0.3.1 - 2023-11-21

### Features
* Add `security_protocol` to `materialize_connection_kafka` [#365](https://github.com/MaterializeInc/terraform-provider-materialize/pull/365)

### BugFixes

* Handle `user` values that contain special characters, without requiring manual
  URL escaping (e.g., escaping `you@corp.com` as `you%40corp.com`) [#372](https://github.com/MaterializeInc/terraform-provider-materialize/pull/372)
* Load generator source `TPCH` requires `ALL TABLES` [#377](https://github.com/MaterializeInc/terraform-provider-materialize/pull/377)
* Improve grant reads [#378](https://github.com/MaterializeInc/terraform-provider-materialize/pull/378)

### Misc
* `materialize_cluster_replica` is deprecated [#370](https://github.com/MaterializeInc/terraform-provider-materialize/pull/370)
* Raise `max_clusters` for testing [#371](https://github.com/MaterializeInc/terraform-provider-materialize/pull/371)

## 0.3.0 - 2023-11-16

### Features
* Add `key_not_enforced` to `materialize_sink_kafka` [#361](https://github.com/MaterializeInc/terraform-provider-materialize/pull/361)

### BugFixes
* Fix a bug where topics were defined after keys in `materialize_sink_kafka` create statements [#358](https://github.com/MaterializeInc/terraform-provider-materialize/pull/358)
* Correct `ForceNew` for column attributes in `materialize_table` [#363](https://github.com/MaterializeInc/terraform-provider-materialize/pull/363)

### Misc
* Update go.mod version to `1.20` [#369](https://github.com/MaterializeInc/terraform-provider-materialize/pull/369)

### Breaking Changes
* Previously, blocks within resources that included optional `schema_name` and `database_name` attributes would inherit the top level attributes of the resource if set. So in the following example:
  ```
  resource "materialize_source_postgres" "example_source_postgres" {
    name          = "source_postgres"
    schema_name   = "my_schema"
    database_name = "my_database"

    postgres_connection {
        name          = "postgres_connection"
    }
  }
  ```
  The Postgres connection would have the schema name of `my_schema` and database name `my_database`. Now, if `schema_name` or `database_name` are not set, they will use the same defaults as top level attributes (`public` for schema and `materialize` for database) [#353](https://github.com/MaterializeInc/terraform-provider-materialize/pull/353)

## 0.2.2 - 2023-11-10

### Features
* Include detail and hint messages for SQL errors [#354](https://github.com/MaterializeInc/terraform-provider-materialize/pull/354)

### BugFixes

### Misc

## 0.2.1 - 2023-11-09

### Features
* Support `ASSERT NOT NULL` for materialized view resource [#341](https://github.com/MaterializeInc/terraform-provider-materialize/pull/341)

### BugFixes

### Misc
* Update testing plugin [#345](https://github.com/MaterializeInc/terraform-provider-materialize/pull/345)

### Breaking Changes
* Update header attributes for `materialize_source_webhook`. Adds `include_header` and is now a complex type `include_headers` and no longer boolean [#346](https://github.com/MaterializeInc/terraform-provider-materialize/pull/346)

## 0.2.0 - 2023-10-30

### Features

### BugFixes

### Misc

### Breaking Changes
* Provider configuration parameters so that they are consistent across all components [#339](https://github.com/MaterializeInc/terraform-provider-materialize/pull/339):
    * The configuration variable `username` is changed to `user`
    * The environment variable `MZ_PW` is changed to `MZ_PASSWORD`

## 0.1.14 - 2023-10-25

### Features

### BugFixes
* Fix `grantRead` failures if the underlying object that the grant is on has been dropped [#338](https://github.com/MaterializeInc/terraform-provider-materialize/pull/338)

### Misc
* Prevent force new for comments on cluster replicas, indexes and roles [#333](https://github.com/MaterializeInc/terraform-provider-materialize/pull/333)
* Mask the local sizes for cluster replicas used by Docker [#355](https://github.com/MaterializeInc/terraform-provider-materialize/pull/335)

## 0.1.13 - 2023-10-12

### Features
* Support for `COMMENTS` on resources [#324](https://github.com/MaterializeInc/terraform-provider-materialize/pull/324)

### BugFixes

### Misc

## 0.1.12 - 2023-09-14

### Features

### BugFixes
* Fix Postgres source `schema` Attribute [#314](https://github.com/MaterializeInc/terraform-provider-materialize/pull/314)
* Require `FOR ALL TABLES` Multi Output Sources [#310](https://github.com/MaterializeInc/terraform-provider-materialize/pull/310)

### Misc

## 0.1.11 - 2023-09-07

### Features

### BugFixes
* Add support for format `JSON` in Kafka source [#305](https://github.com/MaterializeInc/terraform-provider-materialize/pull/305)

### Misc

### Breaking Changes
* Remove `Table`` attributes for load gen source [#303](https://github.com/MaterializeInc/terraform-provider-materialize/pull/303)

## 0.1.10 - 2023-09-01

### Features

### BugFixes
* Fix `ALL ROLES` for default grants [#300](https://github.com/MaterializeInc/terraform-provider-materialize/pull/300)

### Misc

## 0.1.9 - 2023-08-31

### Features
* Support FOR SCHEMAS for postgres source [#262](https://github.com/MaterializeInc/terraform-provider-materialize/pull/262)

### BugFixes
* Remove unnecessary default privilege attributes [#294](https://github.com/MaterializeInc/terraform-provider-materialize/pull/294)

### Misc

## 0.1.8 - 2023-08-24

### Features
* New resource `materialize_source_webhook` for [Webhook Sources](https://materialize.com/docs/sql/create-source/webhook/) [#271](https://github.com/MaterializeInc/terraform-provider-materialize/pull/271)
* Support `disk` attribute for clusters and replicas [#279](https://github.com/MaterializeInc/terraform-provider-materialize/pull/279)

### BugFixes

### Misc
* Include missing attributes for managed cluster data sources [#282](https://github.com/MaterializeInc/terraform-provider-materialize/pull/282)
* Support key rotation for SSH tunnels [#278](https://github.com/MaterializeInc/terraform-provider-materialize/pull/278)

## 0.1.7 - 2023-08-17

### Features
* Support for `ADD|DROP` tables with postgres sources [#265](https://github.com/MaterializeInc/terraform-provider-materialize/pull/265)
* Additional attributes for managed clusters [#275](https://github.com/MaterializeInc/terraform-provider-materialize/pull/275)

### BugFixes

### Misc
* Consistent documentation for common attributes [#276](https://github.com/MaterializeInc/terraform-provider-materialize/pull/276)

## 0.1.6 - 2023-08-09

### Features

### BugFixes
* Correct replica sizes >xlarge [#268](https://github.com/MaterializeInc/terraform-provider-materialize/pull/268)

### Misc

## 0.1.5 - 2023-08-07

### Features
* Include `subsource` as computed attribute for sources [#263](https://github.com/MaterializeInc/terraform-provider-materialize/pull/263)

### BugFixes
* Remove managed clusters testing flag [#261](https://github.com/MaterializeInc/terraform-provider-materialize/pull/261)

### Misc

### Breaking Changes
* Remove `ownership` for cluster replica resource [#259](https://github.com/MaterializeInc/terraform-provider-materialize/pull/259)
* Require `target_role_name` for all default privilege resources [#260](https://github.com/MaterializeInc/terraform-provider-materialize/pull/260)
* Require `col_expr` for index resources [#220](https://github.com/MaterializeInc/terraform-provider-materialize/pull/220)

## 0.1.4 - 2023-07-27

### Features

### BugFixes
* Fix removing grants outside of Terraform state [#245](https://github.com/MaterializeInc/terraform-provider-materialize/pull/245)

### Misc

## 0.1.3 - 2023-07-27

### Features
* Support `INCLUDE KEY AS <name>` for Kafka sources [#250](https://github.com/MaterializeInc/terraform-provider-materialize/pull/250)

### BugFixes
* Fix Default type grants case sensitivity [#247](https://github.com/MaterializeInc/terraform-provider-materialize/pull/247)
* Remove Kafka Source Primary Key [#243](https://github.com/MaterializeInc/terraform-provider-materialize/pull/243)

### Misc
* RBAC Refactor [#234](https://github.com/MaterializeInc/terraform-provider-materialize/pull/234)

## 0.1.2 - 2023-07-17

### Features
* Include `WITH (VALIDATE = false)` for testing [#236](https://github.com/MaterializeInc/terraform-provider-materialize/pull/236)

### BugFixes
* Fix identifier quoting [#239](https://github.com/MaterializeInc/terraform-provider-materialize/pull/239)

### Misc

## 0.1.1 - 2023-07-14

### Features

### BugFixes
* Qualify role name in grant resources [#235](https://github.com/MaterializeInc/terraform-provider-materialize/pull/235)

### Misc

## 0.1.0 - 2023-07-11

### Features
* Revised RBAC resources [#218](https://github.com/MaterializeInc/terraform-provider-materialize/pull/218). A full overview of the Terraform RBAC resources can be found in the `rbac.md`
* Support Managed Clusters [#216](https://github.com/MaterializeInc/terraform-provider-materialize/pull/216)
* Support `FORMAT JSON` for sources [#227](https://github.com/MaterializeInc/terraform-provider-materialize/pull/227)
* Support `EXPOSE PROGRESS` for kafka and postgres sources [#213](https://github.com/MaterializeInc/terraform-provider-materialize/pull/213)

### BugFixes
* Rollback resource creation if ownership query fails [#221](https://github.com/MaterializeInc/terraform-provider-materialize/pull/221)

### Misc
* Table context read includes column attributes [#215](https://github.com/MaterializeInc/terraform-provider-materialize/pull/215)

### Breaking Changes
* As part of the [#218](https://github.com/MaterializeInc/terraform-provider-materialize/pull/218) grant resources introduced in `0.0.9` have been renamed from `materialize_grant_{object}` to `materialize_{object}_grant`

## 0.0.9 - 2023-06-23

### Features
* Resource type `grants` ([#191](https://github.com/MaterializeInc/terraform-provider-materialize/pull/191), [#205](https://github.com/MaterializeInc/terraform-provider-materialize/pull/205), [#209](https://github.com/MaterializeInc/terraform-provider-materialize/pull/209))
* Enable resource and data source `roles` [#206](https://github.com/MaterializeInc/terraform-provider-materialize/pull/206)
* Add attribute `ownership_role` to existing resources ([#208](https://github.com/MaterializeInc/terraform-provider-materialize/pull/208), [#211](https://github.com/MaterializeInc/terraform-provider-materialize/pull/211))

### BugFixes

### Misc

## 0.0.8 - 2023-06-15

### Features
* Include acceptance tests ([#177](https://github.com/MaterializeInc/terraform-provider-materialize/pull/177), [#198](https://github.com/MaterializeInc/terraform-provider-materialize/pull/198), [#200](https://github.com/MaterializeInc/terraform-provider-materialize/pull/200), [#201](https://github.com/MaterializeInc/terraform-provider-materialize/pull/201))

### BugFixes
* Fixes for resource updates (included as part of acceptance test coverage)
* Correct schema index read [#202](https://github.com/MaterializeInc/terraform-provider-materialize/pull/202)
* Attributes missing force new ([#188](https://github.com/MaterializeInc/terraform-provider-materialize/pull/188), [#189](https://github.com/MaterializeInc/terraform-provider-materialize/pull/189))

### Misc
* Include `application_name` in connection string [#184](https://github.com/MaterializeInc/terraform-provider-materialize/pull/184)

## 0.0.7 - 2023-06-07

### Features

### BugFixes
* Handle missing resources on refresh [#176](https://github.com/MaterializeInc/terraform-provider-materialize/pull/176)
* Typo in Privatelink Connection [#182](https://github.com/MaterializeInc/terraform-provider-materialize/pull/182)

### Misc

## 0.0.6 - 2023-05-31

### Features
* Resource and data source [Type](https://materialize.com/docs/sql/create-type/)
* Support for load generator type [Marketing](https://materialize.com/docs/sql/create-source/load-generator/#marketing)

### BugFixes

### Misc
* Refactor of `materialize` package ([#164](https://github.com/MaterializeInc/terraform-provider-materialize/pull/164), [#161](https://github.com/MaterializeInc/terraform-provider-materialize/pull/161), [#158](https://github.com/MaterializeInc/terraform-provider-materialize/pull/158))
* Improvements to documentation

## 0.0.5 - 2023-05-18

### Features
* Include datasource `materialize_egress_ips`

### BugFixes
* Remove improper validation for cluster replica availability zones
* Include `3xsmall` as a valid size

### Misc
* Update index queries to use `mz_objects`

## 0.0.4 - 2023-05-01

### Features
* Include `cluster_name` as a read parameter for the Materialized view query
* Include SSH keys in SSH connection resource

### BugFixes
* Cleanup `resources` Functions
* Fix Slice Params

### Misc

## 0.0.3 - 2023-04-20

### Features
* Adds `principal` property to the AWS PrivateLink connection resource

### BugFixes

### Misc
* Remove unnecessary type property
* Dependabot updates

## 0.0.2 - 2023-04-18

### Features

### BugFixes
* Fixes to datasources and added coverage to integration tests
* Fixes to `UpdateContext` to resources and added coverage to unit tests

### Misc
* Change the Go import path

## 0.0.1 - 2023-04-06

Initial release.
