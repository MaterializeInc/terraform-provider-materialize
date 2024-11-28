## Patterns
Below are some of the patterns and standards we have decided on for the provider.

### Builder
All the SQL objects in the `materialize` package follow the [builder pattern](https://en.wikipedia.org/wiki/Builder_pattern). This makes it easy to accommodate the large number of parameters that can be set when defining resources. For consistency, even simple resources with few parameters follow the same pattern.

After defining the struct for the resource:
```
type ResourceBuilder struct {
    name      string
    parameter string
}
```
We add functions for `Create` and `Drop` and if applicable we also include functions for allowed updates like `Rename`. These funcs generate the SQL to perform the associated action.

### Parameters
The parameters for the provider resources should match as closely as possible to the parameters outlined in the documentation. The same is true if a parameter is optional.

### Nested Types
When defining the struct for the builder, use nested types when possible for clarity. The following would be preferred:
```
type SubParam struct {
    name      string
    parameter string
}

type ResourceBuilder struct {
    name      string
    subparam  []SubParam
}
```
To the more ambiguous:
```
type ResourceBuilder struct {
    name      string
    parameter []map[string]interface{}
}
```

### Ids
The id for all provider resources is the corresponding id in [mz_catalog](https://materialize.com/docs/sql/system-catalog/mz_catalog/). For example the id for sources can be found in `mz_sources`. This is used both for creating new resources and [importing existing resources into state](https://developer.hashicorp.com/terraform/cli/import).

When initially creating a resource via SQL, the id is not returned as part of the command. That is why after we create a resource the provider will query the mz_catalog using the name (and if applicable schema and database names) to lookup the id which will then be set with the `ReadContext`.

### Dividing Resources
Complex Materialize resources are separated out into more specific provider resources. For example sources are divided across `materialize_source_kafka`, `materialize_source_load_generator`, `materialize_source_postgres`. Resources that have a large number of possibly contradictory parameters should be given their own resource. This offers more guidance by allowing more accurate required parameters and not confusing users with details for unnecessary fields.

### Naming Resources
The names of resources should exactly match Materialize. For example the load generator source should be named `materialize_source_load_generator` to match the [SQL statement](https://materialize.com/docs/sql/create-source/load-generator/).

### Testing
**Unit tests** are spread across the packages:
* `datasources` - Should use the `TestResourceDataRaw` to ensure the parameters are properly executed by the mock database for data sources.
* `materialize` - Should ensure the builder properly executes SQL for all valid permutations of the object.
* `resources` - Should use the `TestResourceDataRaw` to ensure the parameters are properly executed by the mock database for resources.

Being the most lightweight the unit tests should cover the wide of SQL variations that exist with each resource.

**Acceptance tests** use the Terraform `acctest` package to execute Terraform commands in a certain order. These tests are used to ensure the applys, updates and destroys work as expected for each resource. These tests should not cover every SQL permutation but ensure that high level Terraform commands execute against the Materialize image. These tests rely on the docker compose.

**Integration tests** consist of an entire Terraform package in the `integration` directory. This will spin up a docker compose using the `materialized` and surrounding kafka and database dependencies. All resources are applied and destroyed as part of the same terraform project. Any new resources or permutations should be added to the integration tests.
