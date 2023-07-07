# RBAC

Managing role-based access control can be difficult within Terraform. To handle [privileges within Materialize](https://materialize.com/docs/manage/access-control/) there are a number of different resources that can be used to manage the different permissions and relationships of privileges.

Also developing for these grant resources present their own challenges compared to resources for other database objects. Below are the considerations for the different grant resources


## Object Grant

These resources handle [granting privileges](https://materialize.com/docs/sql/grant-privilege/) on specific database objects. Each database object has a unique Terraform grant resource to manage privileges. Currently there is only a 1:1 relationship between an object, role and privilege to the Terraform resource. A database object and role can have multiple grant resources to manage multiple privileges.

### Example
```hcl
resource "materialize_grant_table" "table_grant_select" {
  role_name     = "qa_role"
  privilege     = "SELECT"
  database_name = "example_database"
  schema_name   = "example_schema"
  table_name    = "simple_table"
}

resource "materialize_grant_table" "table_grant_insert" {
  role_name     = "qa_role"
  privilege     = "INSERT"
  database_name = "example_database"
  schema_name   = "example_schema"
  table_name    = "simple_table"
}
```

### Metadata
The metadata for resource grants is stored in `privileges` within the specific mz system catalog for the object associated with the grant. For example, when using the `materialize_grant_table` resource, the privileges will be stored within `mz_tables`.

```sql
> SELECT privileges FROM mz_tables WHERE name = 'simple_table';

                privileges
------------------------------------------
 {s1=arwd/s1,u6=ar/s1,u7=wd/s1}
```

Querying the table will give information for all roles and privileges granted on that given object. The provider will parse the privileges (in our case we are only interested in those associated with the `role` which has an id of `u6` so only `u6=ar/s1`) and ensure the each combination of role, object and privilege is present as part of the `ReadContext`.

### Id
The id for the grant objects is a combination of:
* `GRANT`
* Object Type
* Object Id
* Role Id
* Privilege

The id for `materialize_grant_table.table_grant_select` would be:
```
GRANT|TABLE|u48|u6|SELECT
```


## Role Grant

This resource assigns [one role to another](https://materialize.com/docs/sql/grant-role/). There is a 1:1 relationship between a role and user to the Terraform resource. Though a role can be assigned to multiple users and users can have multiple roles assigned.

### Example
```hcl
resource "materialize_grant_role" "qa_role_grant_joe" {
  role_name   = "qa_role"
  member_name = "joe"
}

resource "materialize_grant_role" "qa_role_grant_emily" {
  role_name   = "qa_role"
  member_name = "emily"
}
```

### Metadata
The metadata for role grants is in the mz system catalog `mz_role_members`.

```sql
> SELECT role_id, member, grantor FROM mz_role_members WHERE role_id = 'u1';

 role_id | member | grantor
---------+--------+---------
 u1      | u2     | s1
 u1      | u3     | s1
```

Querying the table for the specific role will show all of its members. The `ReadContext` for the resource will lookup based on the role and member ids.

### Id
The id for the role grant is a combination of:
* `ROLE MEMBER`
* Role Id
* Member Id

The id for `materialize_grant_role.qa_role_grant_joe` would be:
```
ROLE MEMBER|u1|u2
```

## System Grant

This resource assigns [system level privileges](https://materialize.com/docs/sql/grant-privilege/). This is very similar to the grant objects except there is no specific object being used. This resource also has the unique privileges `CREATEROLE`, `CREATEDB`, `CREATECLUSTER`. There is a 1:1 relationship between the role and privilege to the Terraform resource.

### Example
```hcl
resource "materialize_grant_system" "qa_role_system_createdb" {
  role_name = "qa_role"
  privilege = "CREATEDB"
}
```

### Metadata
The metadata for system grants is in the mz system catalog `mz_system_privileges`.

```sql
> SELECT privileges FROM mz_system_privileges;
 privileges
------------
 s1=RBN/s1
 u2=B/s1
```

Querying the system privileges for the specific role will show all of its privileges. The `ReadContext` for the resource will query all privileges, there is no filtering.

### Id
The id for the system grant is a combination of:
* `GRANT SYSTEM`
* Role Id
* Privilege

The id for `materialize_grant_system.qa_role_system_createdb` would be:
```
GRANT SYSTEM|u2|CREATEDB
```

## Default Privilege Grant

This resource assigns [privileges to objects created in the future](https://materialize.com/docs/sql/alter-default-privileges/). This is very similar to the grant objects except there is no specific object being used. This resource also has many more optional fields since objects can be filtered to only apply to a particular database or schema.

### Example
```hcl
resource "materialize_grant_default_privilege" "test_insert" {
  grantee_name     = "qa_role"
  object_type      = "TABLE"
  privilege        = "INSERT"
  target_role_name = "dev_role"
}

resource "materialize_grant_default_privilege" "test_schema_database" {
  grantee_name     = "qa_role"
  object_type      = "TABLE"
  privilege        = "UPDATE"
  target_role_name = "qa_role"
  schema_name      = "example_schema"
  database_name    = "example_database"
}
```

### Metadata
The metadata for system grants is in the mz system catalog `mz_default_privileges`.

```sql
> SELECT * FROM mz_default_privileges;
 role_id | database_id | schema_id | object_type | grantee | privileges
---------+-------------+-----------+-------------+---------+------------
 u7      |             |           | TABLE       | u2      | a
 u7      | u3          | u9        | TABLE       | u2      | w
```

There will be multiple rows for the `grantee` for those default privileges that do and do not specify database and schemas for the same object type.

Querying the default privileges for the specific role will show all of its privileges. The `ReadContext` for the resource will lookup based on the role name, object type, grantee and optionally database and schema.

### Id
The id for the system grant is a combination of:
* `GRANT DEFAULT`
* Object Type
* Grantee Id
* Target Role Id
* Database Id - Optional
* Schema Id - Optional
* Privilege

The id for `materialize_grant_default_privilege.test_insert` would be:
```
GRANT DEFAULT|TABLE|u2|u7|||INSERT
```

The id for `materialize_grant_default_privilege.test_schema_database` would be:
```
GRANT DEFAULT|TABLE|u2|u7|u9|u3|UPDATE
```