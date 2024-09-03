# Design Document: Implementing Source Versioning / Tables from Sources in Terraform Provider for Materialize

## 1. Introduction

This design document outlines the implementation plan for supporting the new "Source Versioning / Tables from Sources" feature in the Terraform provider for Materialize.

This feature aims to simplify the user experience, create a more unified model for managing data ingested from upstream systems, and provide more flexibility in handling upstream schema changes.

## 2. Background

Materialize is introducing changes to its source and table model:
- The concept of a source will be unified around a single relation representing the progress of that source's ingestion pipeline.
- Subsources will be replaced with tables created from sources.
- Users will use a `CREATE TABLE .. FROM SOURCE ..` statement to ingest data from upstream systems.

These changes require corresponding updates to the Terraform provider to maintain alignment with Materialize's data model and provide a smooth migration path for existing users.

## 3. Objectives

- Update the Terraform provider to support the new source and table model.
- Provide a migration path for existing Terraform configurations.
- Maintain backwards compatibility where possible.
- Ensure the provider can work with both old and new versions of Materialize during the transition period.

## 4. Design

### 4.1 Schema Updates

#### 4.1.1 Existing Source Resources

We will maintain the existing source resources (e.g., `materialize_source_postgres`) but deprecate fields related to subsources:

```go
var sourcePostgresSchema = map[string]*schema.Schema{
    // ... existing fields ...
    "table": {
        Description: "Tables to be ingested from the source. This field is deprecated and will be removed in a future version.",
        Type:        schema.TypeSet,
        Optional:    true,
        Deprecated:  "Use the new `materialize_source_table` resource instead.",
        Elem: &schema.Resource{
            Schema: map[string]*schema.Schema{
                // ... existing table schema ...
            },
        },
    },
}
```

#### 4.1.2 New Table From Source Resource

Introduce a new `materialize_source_table` resource:

```go
var sourceTableSchema = map[string]*schema.Schema{
    "name":               ObjectNameSchema("table", true, false),
    "schema_name":        SchemaNameSchema("table", false),
    "database_name":      DatabaseNameSchema("table", false),
    "source": IdentifierSchema(IdentifierSchemaParams{
        Elem:        "source",
        Description: "The source this table is created from.",
        Required:    true,
        ForceNew:    true,
    }),
    "upstream_name": {
        Type:        schema.TypeString,
        Required:    true,
        ForceNew:    true,
        Description: "The name of the table in the upstream database.",
    },
    "upstream_schema_name": {
        Type:        schema.TypeString,
        Optional:    true,
        ForceNew:    true,
        Description: "The schema of the table in the upstream database.",
    },
    "text_columns": {
        Description: "Columns to be decoded as text.",
        Type:        schema.TypeList,
        Elem:        &schema.Schema{Type: schema.TypeString},
        Optional:    true,
        ForceNew:    true,
    },
    // ... other fields as needed ...
}
```

### 4.2 Resource Implementation

#### 4.2.1 Table Source Table Resource

Implement CRUD operations for the new `materialize_source_table` resource:

```go
func SourceTable() *schema.Resource {
    return &schema.Resource{
        CreateContext: sourceTableCreate,
        ReadContext:   sourceTableRead,
        UpdateContext: sourceTableUpdate,
        DeleteContext: sourceTableDelete,
        Importer: &schema.ResourceImporter{
            StateContext: schema.ImportStatePassthroughContext,
        },
        Schema: sourceTableSchema,
    }
}
```

The `CreateContext` function will use the new SQL syntax:

```sql
CREATE TABLE <database_name>.<schema_name>.<name> FROM SOURCE <source_name> (REFERENCE = <upstream name>) WITH (TEXT COLUMNS = (..), ..)
```

#### 4.2.2 Update Existing Source Resources

Modify the CRUD operations for existing source resources to handle the deprecation of subsource-related fields:

- In `Create` and `Update` operations, if the deprecated `table` field is used, log a warning message advising users to migrate to the new `materialize_source_table` resource.
- In `Read` operations, continue to populate the `table` field if it exists in the state, but also log a deprecation warning.

### 4.3 Migration Strategy

We will not create separate resources with a v2 suffix for sources. Instead, we'll use a gradual migration approach:

1. Deprecate the `table` field in existing source resources.
2. Introduce the new `materialize_source_table` resource.
3. Allow both old and new configurations to coexist during a transition period.

This approach allows users to migrate their configurations gradually:

- Existing sources can still be created and managed.
- New tables (formerly subsources) will be created as separate `materialize_source_table` resources.
- Users can migrate their configurations at their own pace by replacing `table` blocks with `materialize_source_table` resources.

### 4.4 Import Logic

The import logic should work out of the box for the new `materialize_source_table` resource as long we have the required information (e.g., source name, upstream table name) stored in a system catalog table.

The read operation for the `materialize_source_table` resource should be able to fetch the necessary details from the Materialize system catalog to populate the state. If not all information is available, some fields may need to be ignored or set to defaults during import.

### 4.5 Versioning and Compatibility

- These changes will be introduced in a new minor version of the provider (e.g., v0.9.0), not a major version bump.
- The provider will support both old and new Materialize versions during the transition period while this is supported by Materialize itself.
- Deprecation warnings will be logged when users interact with the deprecated `table` field.

### 4.6 Testing

- Update existing tests for source resources to cover the deprecation warnings and backwards compatibility.
- Add all required tests for the new `materialize_source_table` resource.
- Implement integration tests to ensure compatibility with both old and new Materialize versions.

## 5. Migration Guide for Users

Provide a migration guide for users to update their Terraform configurations:

1. Update the provider version to v0.9.0 or later.
2. For each source with subsources:
   a. Keep the existing source resource as-is.
   b. Create a new `materialize_source_table` resource for each former subsource.
   c. Set the `source` in the new resource to the fully qualified name of the source.
3. Run `terraform import` to import the state of the new resources.
4. Gradually remove the deprecated `table` blocks from source resources as you migrate to the new structure.

## 6. Backwards Compatibility

- The `table` field in source resources will be marked as deprecated but still functional during the transition period as long as Materialize supports it.
- Existing Terraform configurations will continue to work without immediate changes.
- Deprecation warnings will be logged when users interact with the deprecated fields.

## 7. Documentation Updates

- Update provider documentation to reflect the new resource and changes to existing resources.
- Create a dedicated migration guide with step-by-step instructions and examples.
- Update examples in the documentation to use the new structure.
- Add a section in the documentation explaining the rationale behind these changes and the benefits of the new model.

## 8. Open Questions

- How long should we maintain support for the deprecated `table` field in source resources?
- What is the expected timeline for Materialize to fully transition to the new source and table model?
- Consider webhook sources and their impact on the migration process.
