# Frontegg Mock Service Refactor and Consolidation

- Associated tracking issue: [#461](https://github.com/MaterializeInc/terraform-provider-materialize/issues/461)
- Associated PRs:
    - Transfer SCIM config endpoints [#28614](https://github.com/MaterializeInc/materialize/pull/28614)
    - Transfer SSO config endpoints [#28177](https://github.com/MaterializeInc/materialize/pull/28177)
    - Transfer additional user endpoints [#27358](https://github.com/MaterializeInc/materialize/pull/27358)


## The Problem

Materialize currently maintains two separate Frontegg mock services: one in [Go](https://github.com/MaterializeInc/terraform-provider-materialize/blob/main/mocks/frontegg/mock_server.go) and another in [Rust](https://github.com/bobbyiliev/materialize/blob/cee14e4212e0dc3757960544928308cbf02b94dd/src/frontegg-mock/src/lib.rs).

This duplication leads to increased maintenance overhead and potential inconsistencies. We need to consolidate these services into a single, well-structured Rust implementation to improve maintainability and ensure consistency.

GitHub Issue: [#461](https://github.com/MaterializeInc/terraform-provider-materialize/issues/461)

## Success Criteria

1. All endpoints from the Go mock service are successfully transferred to the Rust implementation.
2. The Rust implementation is refactored to improve code organization, maintainability, and testability.
3. The refactored service includes comprehensive logging for improved debugging capabilities.
4. Tests are added to the consolidated service to ensure the correctness of the implementation and facilitate future changes.
5. Development and maintenance time for the Frontegg mock service is reduced.
6. The refactored codebase is easier to understand and extend for new team members.

## Out of Scope

- Changing the overall behavior or API of the mock service.
- Rewriting or refactoring any dependent systems that use the mock service.

## Solution Proposal

We propose to complete the transfer of all endpoints from the Go implementation to the Rust implementation, followed by a refactor of the Rust codebase. The refactored structure will be modular, making it easier to maintain and extend in the future.

Key aspects of the refactor:

1. At the moment all the endpoints are in a single file. We will want to split the codebase into separate modules for better organization:

```
src/
├── main.rs
├── lib.rs
├── config.rs
├── server.rs
├── models/
│   ├── mod.rs
│   ├── user.rs
│   ├── token.rs
│   ├── sso.rs
│   └── scim.rs
├── handlers/
│   ├── mod.rs
│   ├── auth.rs
│   ├── user.rs
│   ├── sso.rs
│   ├── group.rs
│   └── scim.rs
├── middleware/
│   ├── mod.rs
│   ├── latency.rs
│   └── role_update.rs
└── utils.rs
```

This structure separates concerns, making the code more manageable and easier to navigate. It also facilitates easier testing of individual components.

1. **Improved Error Handling**: Implement a custom error type for more detailed error messages and consistent error handling throughout the application.

1. **Logging**: Add structured logging throughout the application for better observability and debugging. At the moment, the Rust implementation lacks any logging capabilities, which makes it harder to diagnose issues.

1. **Testing**: Implement unit and integration tests for all endpoints. This includes tests for individual handlers, models, and utilities.

## Minimal Viable Prototype

The minimal viable prototype for this refactor will consist of:

1. A fully functional Rust implementation with all endpoints from both the Go and Rust versions.
1. The modular structure as outlined in the solution proposal, with at least one module (e.g., `handlers/auth.rs`) fully implemented.
1. Implementation of structured logging for the authentication flow.
1. A sample unit test for a handler and an integration test for the authentication flow.

This prototype will be validated by the team to ensure it meets the requirements and improves upon the existing implementations.

## Alternatives

The main alternative considered was to continue maintaining both Go and Rust implementations. This was rejected due to the increased maintenance overhead and potential for inconsistencies. Consolidating into a single, well-structured Rust implementation is the most efficient path forward.

Optionally, we could consider keeping the Rust implementation as a single file but improving the code organization within that file. However, this approach may not be as scalable or maintainable in the long run as new endpoints are added.

## Open questions

1. How will we handle versioning of the mock service?
1. What is the best approach for using the mock service in other projects or repositories like the Terraform provider and the Pulumi provider?
1. What is the best strategy for maintaining consistency between the mock service and the actual Frontegg API?
