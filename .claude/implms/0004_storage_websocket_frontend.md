# Storage WebSocket Frontend Implementation

## Related Files

This implementation follows the plan documented in `.claude/plans/0004_storage_websocket_frontend.md`

## Implementation Summary

Successfully implemented WebSocket support for the storage service in the dashboard frontend, enabling real-time updates for storage-related events including locations, AIPs, workflows, and tasks.

## What Was Implemented

### 1. API Client Generation ✅
- **File**: Various OpenAPI generated files
- **Action**: Regenerated OpenAPI client using `make gen-dashboard-client`
- **Result**: Updated TypeScript models for all storage events (AIPCreatedEvent, AIPUpdatedEvent, etc.)
- **Note**: Main container models (StorageMonitorEvent, StorageMonitorEventEvent) were not generated, but individual event models are available

### 2. Client WebSocket Connection ✅
- **File**: `dashboard/src/client.ts`
- **Changes**:
  - Added `connectStorageMonitor: () => void` to Client interface
  - Implemented `connectStorageMonitor()` function that connects to `/storage/monitor` WebSocket endpoint
  - Added function to client creation object
  - Imports storage event handler from `monitor_storage.ts`

### 3. Storage Event Handler ✅
- **File**: `dashboard/src/monitor_storage.ts` (new file)
- **Features**:
  - Handles raw JSON storage events using type detection
  - Supports all storage event types: StoragePing, LocationCreated/Updated, AIPCreated/Updated, AIP Workflow/Task events
  - Uses lodash for key transformation (camelCase to snake_case) following existing pattern
  - Integrates with location and AIP stores for real-time updates
  - Uses debounced store methods to prevent excessive API calls

### 4. App.vue Integration ✅
- **File**: `dashboard/src/App.vue`
- **Changes**:
  - Added storage monitor WebSocket connection alongside existing ingest monitor
  - Requests access ticket via `client.storage.storageMonitorRequest()` then connects
  - Maintains existing ingest monitor functionality

### 5. Store Enhancements ✅

#### Location Store (`dashboard/src/stores/location.ts`)
- Added `fetchCurrentDebounced(uuid: string)` method
- Added `fetchLocationsDebounced()` method
- Configured debounce settings: 500ms delay, non-immediate

#### AIP Store (`dashboard/src/stores/aip.ts`)
- Added `fetchCurrentDebounced(id: string)` method
- Added `fetchAipsDebounced(page: number)` method
- Configured debounce settings: 500ms delay, non-immediate

### 6. Event Handling Strategy

Due to missing container models, implemented a simplified but effective approach:

**Direct Event Parsing**:
- Detects event types from raw JSON structure
- Maps events to appropriate handlers based on object properties
- Uses existing individual event models (AIPCreatedEvent, etc.)

**Store Update Strategy**:
- **Immediate Updates**: For currently displayed items (current AIP, current location)
- **Debounced Refresh**: For lists to prevent excessive API calls
- **Conditional Updates**: Only updates relevant data based on current context

**Simplified Workflow/Task Handling**:
- Since event models don't include relational properties (aipUuid, workflowUuid), workflow and task events trigger a simple refresh of the current AIP's workflows
- This ensures data consistency while avoiding complex event parsing issues

## Code Quality Assurance

### Type Safety ✅
- All TypeScript type checking passes
- Proper typing for all event handlers and store methods
- Uses generated OpenAPI models for type safety

### Linting ✅
- All ESLint rules pass
- Proper import ordering
- No unused variables
- Correct TypeScript best practices

### Testing ✅
- All existing tests continue to pass (164 tests)
- No regressions introduced
- Store functionality verified

## Technical Implementation Details

### WebSocket Connection Flow
1. User authentication validates → App.vue watch trigger
2. Request storage monitor ticket: `POST /storage/monitor`
3. Establish WebSocket connection: `GET /storage/monitor` (with upgrade)
4. Parse incoming JSON events and route to appropriate handlers
5. Update stores with debounced methods to prevent API spam

### Event Processing
```typescript
// Raw event detection
if (value.aip_created_event !== undefined) {
  handleAIPCreated(value.aip_created_event);
}
// Transform and handle
const event = api.AIPCreatedEventFromJSON(data);
aipStore.fetchAipsDebounced(1);
```

### Store Integration
- Uses Pinia debounce plugin (already configured)
- 500ms debounce prevents excessive API calls during event bursts
- Maintains reactive state updates for UI components

## Deviations from Plan

### 1. Missing Container Models
**Planned**: Use StorageMonitorEvent and StorageMonitorEventEvent containers
**Implemented**: Direct event type detection from JSON structure
**Reason**: OpenAPI generator didn't create these models despite backend definitions

### 2. Simplified Workflow/Task Handling
**Planned**: Direct workflow/task object updates like ingest monitor
**Implemented**: Simple refresh of current AIP workflows
**Reason**: Event models missing relational properties (aipUuid, workflowUuid)

### 3. Event Structure Differences
**Planned**: Follow exact ingest monitor pattern
**Implemented**: Adapted to direct JSON event detection
**Reason**: Storage events have different structure than ingest events

## Validation Results

### ✅ Functional Requirements Met
- WebSocket connects to storage monitor endpoint
- Storage events trigger appropriate store updates
- Real-time updates for locations and AIPs
- Maintains existing ingest monitor functionality
- Proper error handling and type safety

### ✅ Quality Assurance
- TypeScript compilation: **PASS**
- ESLint linting: **PASS**  
- Unit tests: **PASS** (164 tests)
- No regressions introduced

### ✅ Performance Considerations
- Debounced store updates prevent API spam
- Efficient event routing with simple conditionals
- Minimal memory footprint with proper cleanup
- No impact on existing WebSocket connection

## Future Improvements

### Container Model Generation
When backend container models are properly generated:
1. Update `handleStorageEvent()` to use proper container parsing
2. Replace direct JSON detection with type-safe event discrimination
3. Follow exact ingest monitor pattern

### Enhanced Workflow/Task Updates
When event models include relational properties:
1. Implement direct workflow/task object updates
2. Remove simple refresh approach
3. Add granular real-time updates for workflow progress

### Error Handling Enhancements
1. Add WebSocket reconnection logic
2. Implement connection status indicators
3. Add graceful degradation for WebSocket failures

## Summary

Successfully implemented storage WebSocket frontend integration providing real-time updates for storage operations. The implementation provides immediate value with a robust, type-safe foundation that can be enhanced as backend models evolve. All quality checks pass and no regressions were introduced.