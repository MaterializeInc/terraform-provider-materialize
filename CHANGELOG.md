# Changelog

## 0.0.5 - 2022-05-18

### Features

* Include datasource `materialize_egress_ips`

### BugFixes

* Remove improper validation for cluster replica availability zones
* Include `3xsmall` as a valid size

### Misc
* Update index queries to use `mz_objects`

## 0.0.4 - 2022-05-01

### Features

* Include `cluster_name` as a read parameter for the Materialized view query
* Include SSH keys in SSH connection resource

### BugFixes

* Cleanup `resources` Functions
* Fix Slice Params

### Misc

## 0.0.3 - 2022-04-20

### Features

* Adds `principal` property to the AWS PrivateLink connection resource

### BugFixes

### Misc

* Remove unnecessary type property
* Dependabot updates

## 0.0.2 - 2022-04-18

### Features

### BugFixes
* Fixes to datasources and added coverage to integration tests
* Fixes to `UpdateContext` to resources and added coverage to unit tests

### Misc
* Change the Go import path

## 0.0.1 - 2022-04-06

Initial release.
