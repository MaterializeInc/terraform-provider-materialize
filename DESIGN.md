## Patterns
Below are some of the patterns and standards we have decided on for the provider.

### Builder
All the resources in the provider follow the [builder pattern](https://en.wikipedia.org/wiki/Builder_pattern). This makes it easy to accommodate the large number of parameters that can be set when defining resources. For consistency, even simple resources with few parameters follow the same pattern.

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
Complex Materailize resources are separated out into more specific provider resources. For example sources are divided across `materialize_source_kafka`, `materialize_source_loadgen`, `materialize_source_postgres`. Resources that have a large number of possibly contradictory parameters should be given their own resource. This offers more guidance by allowing more accurate required parameters and not confusing users with details for unnecessary fields.