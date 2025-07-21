# Storage Service WebSocket Implementation Plan

## Overview
This plan outlines the implementation of WebSocket support for real-time updates in the storage service, replicating the pattern established in the ingest service.

## Current Architecture Analysis

### 1. Storage Service Structure
- API design in `/internal/api/design/storage.go`
- Service implementation in `/internal/storage/service.go`
- Entity types: Location, AIP, Workflow, Task, DeletionRequest
- Persistence layer using Ent ORM with MySQL

### 2. Existing Event System
- Event service in `/internal/event/` with Redis and in-memory implementations
- Currently used only by ingest service for SIP monitoring
- WebSocket pattern established in ingest service with ticket-based authentication

### 3. Authentication
- JWT-based auth with attribute-based permissions
- Storage attributes: `storage:aips:list`, `storage:aips:read`, `storage:aips:workflows:list`, etc.
- Ticket system for WebSocket connections (cookie-based)

## Implementation Plan

### Phase 1: API Design

#### 1.1 Create Storage Monitor API Design
Create new file `/internal/api/design/storage_monitor.go` with:

**Event Types:**
- `LocationCreatedEvent` - New location created
- `LocationUpdatedEvent` - Location updated
- `AIPCreatedEvent` - New AIP created
- `AIPUpdatedEvent` - AIP updated (including status changes)
- `WorkflowCreatedEvent` - New workflow created
- `WorkflowUpdatedEvent` - Workflow updated
- `TaskCreatedEvent` - New task created
- `TaskUpdatedEvent` - Task updated
- `MonitorPingEvent` - Keep-alive messages

**Endpoints:**
- `POST /storage/monitor` - Request monitoring ticket (requires JWT auth)
- `GET /storage/monitor` - Establish WebSocket connection (uses ticket from cookie)

### Phase 2: Backend Implementation

#### 2.1 Extend Storage Service
Modify `/internal/storage/service.go`:
- Add EventService dependency
- Implement `MonitorRequest` method for ticket generation
- Implement `Monitor` method for WebSocket handling

#### 2.2 Create WebSocket Handler
Create `/internal/storage/monitor.go` with:
- WebSocket connection management
- Event subscription and filtering based on user permissions
- Ping/pong mechanism for connection health
- Graceful disconnect handling

#### 2.3 Add Event Publishing
Update the following methods in storage service to publish events:

**Location Events:**
- `CreateLocation()` → publish `LocationCreatedEvent`
- `UpdateLocation()` → publish `LocationUpdatedEvent`

**AIP Events:**
- `CreateAip()` → publish `AIPCreatedEvent`
- `UpdateAipStatus()` → publish `AIPUpdatedEvent`
- `MoveAip()` → publish `AIPUpdatedEvent`
- `RejectAip()` → publish `AIPUpdatedEvent`

**Workflow Events:**
- `CreateWorkflow()` → publish `WorkflowCreatedEvent`
- `UpdateWorkflow()` → publish `WorkflowUpdatedEvent`

**Task Events:**
- `CreateTask()` → publish `TaskCreatedEvent`
- `UpdateTask()` → publish `TaskUpdatedEvent`

### Phase 3: Event Service Integration

#### 3.1 Update Event Types
Modify `/internal/event/publish.go`:
- Add storage event type constants
- Ensure proper JSON marshaling for all event types

#### 3.2 Wire Dependencies
Update `/cmd/enduro/main.go`:
- Ensure EventService is wired to storage service (if not already)
- Verify Redis connection is available for both services

### Phase 4: Authorization and Security

#### 4.1 Permission Mapping
Define which attributes allow access to which events:
- `storage:locations:list` → Location events
- `storage:aips:list` → AIP created events
- `storage:aips:read` → AIP updated events
- `storage:aips:workflows:list` → Workflow/Task events

#### 4.2 Ticket Management
- Reuse existing ticket store implementation
- 5-second TTL for security
- Cookie-based authentication for WebSocket upgrade

### Phase 5: Testing

#### 5.1 Unit Tests
- Event publishing verification
- Permission filtering logic
- WebSocket message formatting

#### 5.2 Integration Tests
- End-to-end WebSocket connection flow
- Event propagation from database changes to client
- Authentication and authorization scenarios

#### 5.3 Load Testing
- Multiple concurrent WebSocket connections
- High-frequency event publishing
- Connection stability over time

## Implementation Order

1. Create API design file with event types
2. Implement basic WebSocket handler
3. Add event publishing to one entity type (e.g., AIP)
4. Test end-to-end flow with single entity
5. Extend to all entity types
6. Add comprehensive permission checking
7. Write tests
8. Update API documentation

## Estimated Effort

- API Design: 2 hours
- WebSocket Handler: 4 hours
- Event Publishing: 3 hours
- Testing: 4 hours
- Documentation: 1 hour

**Total: ~14 hours**

## Risks and Considerations

1. **Performance Impact**: Publishing events on every database operation may impact performance. Consider batching or rate limiting if needed.

2. **Event Ordering**: Ensure events are published in the correct order, especially for rapid state changes.

3. **Connection Management**: Plan for handling large numbers of concurrent WebSocket connections in production.

4. **Backwards Compatibility**: Ensure existing API clients continue to work without WebSocket support.

5. **Error Handling**: Define behavior when event publishing fails (should not block the main operation).

## Success Criteria

1. All storage entity changes trigger real-time events
2. Events are properly filtered based on user permissions
3. WebSocket connections remain stable under normal load
4. Frontend can consume events without modification to existing patterns
5. No performance degradation in storage operations