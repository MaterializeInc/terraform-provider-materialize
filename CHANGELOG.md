# Changelog

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
