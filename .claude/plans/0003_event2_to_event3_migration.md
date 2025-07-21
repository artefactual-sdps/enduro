# Event2 to Event3 Migration Plan

## User Request
"Migrate the codebase from the event2 package to the new event3 package, replacing the unified event service with separate type-safe services for ingest and storage events"

## Related Files
- Implementation: `.claude/implms/0003_event2_to_event3_migration.md`

## Overview
Migrate from the unified `event2` package to the type-safe `event3` package. The key architectural change is moving from a single `EventService` that handles all event types to separate, strongly-typed services for ingest and storage events.

## Key Differences

### Event2 (Current)
- Single `EventService` interface handles all event types (`any`)
- Single unified service instance manages both ingest and storage events
- Events are wrapped dynamically based on type using type switches
- Single Redis channel for all events

### Event3 (Target)
- Generic `Service[T]` interface with type parameters
- Separate `IngestEventService` and `StorageEventService` instances
- Type-safe event publishing with dedicated functions
- Separate Redis channels/serializers for each event type

## Migration Strategy

### Phase 1: Service Instantiation Changes
1. **Main binary (`cmd/enduro/main.go`)**
   - Replace single `event2.NewEventServiceRedis()` with two separate services:
     - `event3.NewIngestEventServiceRedis()` for ingest events
     - `event3.NewStorageEventServiceRedis()` for storage events
   - Update service dependencies to use appropriate event service

2. **Worker binaries**
   - `cmd/enduro-am-worker/main.go`: Use ingest event service
   - `cmd/enduro-a3m-worker/main.go`: Use ingest event service

3. **Service constructors**
   - Update `ingest.NewService()` to accept `event3.IngestEventService`
   - Update `storage.NewService()` to accept `event3.StorageEventService`

### Phase 2: Service Interface Updates
1. **Ingest service (`internal/ingest/ingest.go`)**
   - Change field type from `event.EventService` to `event3.IngestEventService`
   - Update all event publishing calls to use `event3.PublishIngestEvent()`

2. **Storage service (`internal/storage/service.go`)**
   - Change field type from `event.EventService` to `event3.StorageEventService`
   - Update all event publishing calls to use `event3.PublishStorageEvent()`

### Phase 3: Monitor Updates
1. **Ingest monitor (`internal/ingest/monitor.go`)**
   - Update subscription to use `IngestEventService.Subscribe()`
   - Update type assertions for `*goaingest.MonitorEvent`

2. **Storage monitor (`internal/storage/monitor.go`)**
   - Update subscription to use `StorageEventService.Subscribe()`
   - Update type assertions for `*goastorage.StorageMonitorEvent`

### Phase 4: Test Updates
1. Update all test files that use event services:
   - `internal/ingest/monitor_test.go`
   - `internal/ingest/ingest_test.go`
   - `internal/storage/service_test.go`
2. Replace `event.NewEventServiceInMemImpl()` with appropriate type-safe versions

### Phase 5: Import Updates
Update all import statements:
- Change `event "github.com/artefactual-sdps/enduro/internal/event2"` 
- To `event3 "github.com/artefactual-sdps/enduro/internal/event3"`

## Implementation Details

### Configuration Impact
- No changes to `Config` struct (same Redis address/channel fields)
- Event3 uses separate Redis channels internally for type safety

### Service Dependencies
**Current dependency injection:**
```go
NewService(..., evsvc event.EventService, ...)
```

**New dependency injection:**
```go
// Ingest service
NewService(..., evsvc event3.IngestEventService, ...)

// Storage service  
NewService(..., evsvc event3.StorageEventService, ...)
```

### Event Publishing Changes
**Current publishing:**
```go
event.PublishEvent(ctx, svc.evsvc, &goaingest.SIPCreatedEvent{...})
```

**New publishing:**
```go
event3.PublishIngestEvent(ctx, svc.evsvc, &goaingest.SIPCreatedEvent{...})
```

### Monitor Subscription Changes
**Current subscription:**
```go
sub, err := w.evsvc.Subscribe(ctx)
// Type assertion: event, ok := eventAny.(*goaingest.MonitorEvent)
```

**New subscription:**
```go
sub, err := w.ingestEvSvc.Subscribe(ctx) 
// Direct typed access: event := <-sub.C()
```

## Risk Assessment

### Low Risk
- Event3 package already exists and is tested
- Configuration remains the same
- Event payloads unchanged

### Medium Risk
- Multiple service instances vs single unified service
- Type safety may catch hidden bugs
- Redis channel separation may affect cross-service event consumption

### Mitigation
- Thorough testing of monitor functionality
- Verify event delivery across ingest/storage boundaries
- Test Redis connectivity for both service types

## Benefits of Migration

1. **Type Safety**: Compile-time validation of event types
2. **Separation of Concerns**: Clear boundary between ingest and storage events
3. **Better Testing**: Easier to mock and test specific event types
4. **Future Extensibility**: Generic design supports additional event types

## Rollback Plan
- Event2 package remains available for quick rollback
- Configuration compatibility maintained
- All changes are code-level only, no data migration required

## Testing Strategy
1. Unit tests for all modified services
2. Integration tests for monitor functionality  
3. End-to-end tests for event flow across services
4. Redis connectivity tests for both service types