# Changelog

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
