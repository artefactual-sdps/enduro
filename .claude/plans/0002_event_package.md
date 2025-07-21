# Event Package Design: Hybrid Type-Safe Approach

## User Request
"Design and implement a new event package that combines type safety with simplicity, avoiding the drawbacks of the current event and event2 packages"

## Related Files
- Implementation: `.claude/implms/0002_event_package.md`

## Analysis of Three Approaches

### 1. Main Branch (Simple Approach)
- **Interface**: Single concrete `EventService` for `*goaingest.MonitorEvent` only
- **Type Safety**: High for ingest events only  
- **Complexity**: Low - minimal abstraction
- **Separation**: Limited - ingest events only
- **Usage**: Direct service method calls

### 2. Dev Branch (Type-Safe Generic Approach)  
- **Interface**: Generic `Service[T any]` with type aliases (`EventService`, `StorageEventService`)
- **Type Safety**: Very high - compile-time type checking
- **Complexity**: Medium - effective use of Go generics
- **Separation**: Excellent - clear service boundaries
- **Usage**: Two separate services with helper functions

### 3. Event2 (Unified Any Approach)
- **Interface**: Single service with `any` type + runtime type switching
- **Type Safety**: Low - runtime type assertions required
- **Complexity**: Medium - complex publish logic  
- **Separation**: Poor - mixed event types
- **Usage**: Single service with type assertion in consumers

## Recommended Solution: Enhanced Type-Safe Design

Create a hybrid approach combining the best aspects:

### Core Principles
1. **Simplicity** from main branch approach
2. **Type Safety** with separate event types like dev branch  
3. **Single Service Interface** concept from event2 but with better type safety
4. **Clean separation** between ingest and storage events

### Implementation Plan

1. **Create an event3 package** with:
   - Generic `Service[T]` interface for type safety
   - Convenience constructors for common services
   - Clean separation between ingest and storage events
   - Support for both in-memory and Redis implementations

2. **Key Features:**
   ```go
   // Core generic interface
   type Service[T any] interface {
       PublishEvent(ctx context.Context, event T)
       Subscribe(ctx context.Context) (Subscription[T], error)
   }
   
   // Type aliases for convenience
   type IngestEventService = Service[*goaingest.MonitorEvent]
   type StorageEventService = Service[*goastorage.StorageMonitorEvent]
   
   // Type-safe publish helpers
   func PublishIngestEvent(ctx context.Context, svc IngestEventService, event IngestEvent)
   func PublishStorageEvent(ctx context.Context, svc StorageEventService, event StorageEvent)
   ```

3. **Benefits:**
   - **Type Safety**: Compile-time checking prevents wrong event types
   - **Simplicity**: Clean API with helpful type aliases  
   - **Flexibility**: Handles current and future event types
   - **Performance**: No runtime type switching in hot paths
   - **Maintainability**: Clear separation of concerns
   - **Backward Compatibility**: Minimal changes needed in existing code

4. **Migration Path:**
   - Create event3 package alongside existing packages
   - Migrate services one by one
   - Remove event and event2 packages once migration is complete

## Detailed Design

### Interface Design
```go
package event3

import (
    "context"
    goaingest "github.com/artefactual-sdps/enduro/internal/api/gen/ingest"
    goastorage "github.com/artefactual-sdps/enduro/internal/api/gen/storage"
)

// Generic service interface
type Service[T any] interface {
    PublishEvent(ctx context.Context, event T)
    Subscribe(ctx context.Context) (Subscription[T], error)
}

// Generic subscription interface
type Subscription[T any] interface {
    C() <-chan T
    Close() error
}

// Type aliases for convenience
type IngestEventService = Service[*goaingest.MonitorEvent]
type StorageEventService = Service[*goastorage.StorageMonitorEvent]
type IngestSubscription = Subscription[*goaingest.MonitorEvent]
type StorageSubscription = Subscription[*goastorage.StorageMonitorEvent]
```

### Constructor Functions
```go
// Constructors for different backends
func NewIngestEventServiceInMem() IngestEventService
func NewStorageEventServiceInMem() StorageEventService
func NewIngestEventServiceRedis(logger, tp, cfg) (IngestEventService, error)
func NewStorageEventServiceRedis(logger, tp, cfg) (StorageEventService, error)

// Nop services for testing
func NopIngestEventService() IngestEventService
func NopStorageEventService() StorageEventService
```

### Type-Safe Publish Helpers
```go
// Type-safe publish functions that wrap individual events
func PublishIngestEvent(ctx context.Context, svc IngestEventService, event any) {
    // Type switch for ingest events only
    switch v := event.(type) {
    case *goaingest.MonitorPingEvent:
        svc.PublishEvent(ctx, &goaingest.MonitorEvent{Event: v})
    case *goaingest.SIPCreatedEvent:
        svc.PublishEvent(ctx, &goaingest.MonitorEvent{Event: v})
    // ... other ingest events
    default:
        panic("invalid ingest event type")
    }
}

func PublishStorageEvent(ctx context.Context, svc StorageEventService, event any) {
    // Type switch for storage events only
    switch v := event.(type) {
    case *goastorage.StorageMonitorPingEvent:
        svc.PublishEvent(ctx, &goastorage.StorageMonitorEvent{Event: v})
    case *goastorage.LocationCreatedEvent:
        svc.PublishEvent(ctx, &goastorage.StorageMonitorEvent{Event: v})
    // ... other storage events
    default:
        panic("invalid storage event type")
    }
}
```

This design provides the best of all worlds: type safety, simplicity, and clean separation of concerns while maintaining performance and backward compatibility.