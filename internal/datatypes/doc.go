// Package datatypes defines the core domain entities used throughout Enduro.
//
// These types represent the fundamental business objects (SIP, Batch, Workflow,
// Task, User) and serve as the canonical data structures shared across
// application layers. They are intentionally decoupled from persistence
// concerns (e.g., ent schemas) and API representations (e.g., Goa types),
// acting as an intermediary that both layers can convert to and from.
//
// By maintaining domain types separate from ORM-generated or API-generated
// types, the codebase achieves loose coupling (business logic depends on stable
// domain types rather than auto-generated code), clear boundaries (the
// persistence and API layers convert to/from domain types), and testability
// (domain types can be instantiated directly without infrastructure).
package datatypes
