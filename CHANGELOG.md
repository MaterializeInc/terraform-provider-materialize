# Changelog

## Unreleased

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
