# Storage WebSocket Frontend Implementation Plan

## User Request

We have added Websocket support to the storage service in the backend/API through the /monitor endpoints. We need to implement the frontend side in the dashboard now. Follow these steps:

- Confirm the API client code in the dashboard is up-to-date
- Add connectStorageMonitor to client.ts
- Connect to the websocket when the user is valid in App.vue
- Handle the storage events, two options for these, choose the one you find best:
  - Mix both ingest and storage event types in the existing handler
  - Create an specific handler for the storage events in monitor_storage.ts
- The handlers need to update the location and the aip stores state

## Related Files

Implementation details will be documented in `.claude/implms/0004_storage_websocket_frontend.md`

## Analysis & Current State

### Backend Status ✅
- Storage monitor WebSocket endpoints are fully implemented
- API endpoints: `POST /storage/monitor` (request ticket) and `GET /storage/monitor` (WebSocket)
- Event types defined: StoragePingEvent, LocationCreated/Updated, AIPCreated/Updated, AIPWorkflowCreated/Updated, AIPTaskCreated/Updated
- Generated TypeScript models exist for individual event types

### Frontend Gaps ❌
- No WebSocket connection to storage monitor
- Missing main container event models (StorageMonitorEvent, StorageMonitorEventEvent)
- No event handlers for storage events
- No real-time updates for location and AIP stores

### Existing Patterns
- Ingest monitor already implemented with similar WebSocket pattern
- Client connection in `client.ts` with `connectIngestMonitor()`
- Event handling in `monitor.ts` with type-safe handlers
- Store integration with `useSipStore()` for real-time updates

## Architecture Decision: Separate Storage Event Handler

After analyzing the codebase, I recommend creating a **separate storage event handler** (`monitor_storage.ts`) for these reasons:

1. **Type Safety**: Storage events have different types than ingest events
2. **Separation of Concerns**: Storage and ingest have different data models and update patterns
3. **Maintainability**: Easier to maintain separate handlers than a mixed handler
4. **Existing Pattern**: Current monitor.ts is specifically for ingest events with `IngestMonitorEventEventTypeEnum`

## Implementation Plan

### Step 1: Update Generated API Models
**Estimated Time**: 5 minutes
- Regenerate OpenAPI client to ensure StorageMonitorEvent and StorageMonitorEventEvent models exist
- Verify all storage event models are up-to-date

### Step 2: Add Storage Monitor Connection to Client
**File**: `dashboard/src/client.ts`
**Estimated Time**: 10 minutes

1. Add `connectStorageMonitor: () => void` to Client interface
2. Implement `connectStorageMonitor()` function following the ingest pattern:
   - Get WebSocket URL + "/storage/monitor"
   - Create WebSocket connection
   - Parse JSON messages using `StorageMonitorEventFromJSON`
   - Call storage event handler

3. Add to client creation function

### Step 3: Create Storage Event Handler
**File**: `dashboard/src/monitor_storage.ts` (new file)
**Estimated Time**: 30 minutes

1. Import necessary dependencies:
   - `api` from client
   - Storage event enums and types
   - `useLocationStore` and `useAipStore`
   - lodash utilities for key transformation

2. Main handler function:
   ```typescript
   export function handleStorageEvent(event: api.StorageMonitorEventEvent)
   ```

3. Type-safe handler mapping:
   ```typescript
   const handlers: {
     [key in api.StorageMonitorEventEventTypeEnum]: (data: unknown) => void;
   } = { ... }
   ```

4. Individual event handlers:
   - `handleStoragePing`: No-op for heartbeat
   - `handleLocationCreated`: Update location store, trigger refresh
   - `handleLocationUpdated`: Update specific location in store
   - `handleAIPCreated`: Update AIP store, trigger refresh
   - `handleAIPUpdated`: Update specific AIP in store if currently displayed
   - `handleAIPWorkflowCreated`: Add workflow to current AIP
   - `handleAIPWorkflowUpdated`: Update workflow in current AIP
   - `handleAIPTaskCreated`: Add task to workflow
   - `handleAIPTaskUpdated`: Update specific task

### Step 4: Connect WebSocket in App.vue
**File**: `dashboard/src/App.vue`
**Estimated Time**: 5 minutes

1. Update the watch function to also connect storage monitor:
   ```typescript
   watch(
     () => authStore.isUserValid,
     (valid) => {
       if (valid) {
         // Existing ingest monitor
         client.ingest.ingestMonitorRequest().then(() => {
           client.connectIngestMonitor();
         });
         
         // New storage monitor
         client.storage.storageMonitorRequest().then(() => {
           client.connectStorageMonitor();
         });
       }
     },
     { immediate: true },
   );
   ```

### Step 5: Store Integration & Real-time Updates
**Files**: `dashboard/src/stores/location.ts`, `dashboard/src/stores/aip.ts`
**Estimated Time**: 15 minutes

**Location Store Enhancements**:
- Add debounced refresh methods similar to sip store
- Update current location when location events received
- Refresh location list when new locations created

**AIP Store Enhancements**:
- Add debounced refresh methods
- Update current AIP when AIP events received  
- Update workflows and tasks in real-time
- Handle AIP status changes and location moves

### Step 6: Error Handling & Connection Management
**Estimated Time**: 10 minutes

1. Add error handling for WebSocket connections
2. Handle connection drops and reconnection
3. Ensure proper cleanup on auth state changes

### Step 7: Testing & Validation
**Estimated Time**: 15 minutes

1. Test WebSocket connection establishment
2. Verify event parsing and handling
3. Confirm store updates work correctly
4. Test error scenarios (connection drops, invalid events)

## Technical Considerations

### Event Data Transformation
- Follow existing pattern with `mapKeys` and `snakeCase` for JSON key transformation
- Handle nested objects (item property) similar to ingest events

### Store Update Strategies
- **Immediate Updates**: For current displayed items (current AIP, current location)
- **Debounced Refresh**: For lists (prevent excessive API calls)
- **Conditional Updates**: Only update if item is currently displayed

### Memory Management
- Ensure WebSocket connections are properly closed
- Clean up event listeners on component unmount
- Handle authentication state changes

### Browser Compatibility
- WebSocket is well supported in modern browsers
- Consider connection retry logic for unreliable networks

## Success Criteria

1. ✅ WebSocket connects successfully to storage monitor endpoint
2. ✅ Storage events are received and parsed correctly
3. ✅ Location store updates in real-time when locations change
4. ✅ AIP store updates in real-time when AIPs/workflows/tasks change
5. ✅ No performance degradation from additional WebSocket connection
6. ✅ Proper error handling and connection management
7. ✅ Existing ingest monitor functionality remains unaffected

## Risk Mitigation

- **API Model Issues**: Regenerate OpenAPI client before implementation
- **Event Format Changes**: Use generated TypeScript models for type safety
- **Performance Impact**: Implement debounced updates and conditional rendering
- **Connection Issues**: Add retry logic and graceful degradation
- **State Consistency**: Ensure events don't conflict with manual refresh operations