import { api } from "@/client";
import { StorageEventValueTypeEnum } from "@/openapi-generator";
import { useAipStore } from "@/stores/aip";
import { useLocationStore } from "@/stores/location";

// Local websocket event boundary used by the storage monitor code.
//
// The generated OpenAPI client gives us the outer event envelope (`type` and
// `value`), but it does not preserve the monitor payload `anyOf` as a usable
// discriminated union. Instead, `value` is typed as a single generated model
// that does not correctly represent every storage monitor event payload.
//
// We therefore treat `value` as `unknown` here and decode it in the per-event
// handlers with the appropriate generated `FromJSON` helper.
type StorageMonitorEvent = {
  type: api.StorageEventValueTypeEnum;
  value: unknown;
};

// TODO: Replace this unknown-based event boundary with a typed monitor-event
// decoder once the generated client preserves the websocket payload `anyOf`,
// or after adding a small local typed decoder facade.
export function handleStorageEvent(event: StorageMonitorEvent) {
  const handler = handlers[event.type];
  if (!handler) return;
  handler(event.value);
}

const handlers: {
  [key in api.StorageEventValueTypeEnum]: (data: unknown) => void;
} = {
  [StorageEventValueTypeEnum.StoragePingEvent]: () => {},
  [StorageEventValueTypeEnum.LocationCreatedEvent]: handleLocationCreated,
  [StorageEventValueTypeEnum.AipCreatedEvent]: handleAipCreated,
  [StorageEventValueTypeEnum.AipUpdatedEvent]: handleAipUpdated,
  [StorageEventValueTypeEnum.AipStatusUpdatedEvent]: handleAipStatusUpdated,
  [StorageEventValueTypeEnum.AipLocationUpdatedEvent]: handleAipLocationUpdated,
  [StorageEventValueTypeEnum.AipWorkflowCreatedEvent]: handleAipWorkflowCreated,
  [StorageEventValueTypeEnum.AipWorkflowUpdatedEvent]: handleAipWorkflowUpdated,
  [StorageEventValueTypeEnum.AipTaskCreatedEvent]: handleAipTaskCreated,
  [StorageEventValueTypeEnum.AipTaskUpdatedEvent]: handleAipTaskUpdated,
};

function handleLocationCreated() {
  const locationStore = useLocationStore();
  locationStore.fetchLocationsDebounced();
}

function handleAipCreated() {
  const aipStore = useAipStore();
  aipStore.fetchAipsDebounced(1);
}

function handleAipUpdated(data: unknown) {
  const event = api.AIPUpdatedEventFromJSON(data);
  const aipStore = useAipStore();
  aipStore.fetchAipsDebounced(1);
  if (aipStore.current?.uuid != event.uuid) return;
  Object.assign(aipStore.current, event.item);
}

function handleAipStatusUpdated(data: unknown) {
  const event = api.AIPStatusUpdatedEventFromJSON(data);
  const aipStore = useAipStore();
  aipStore.fetchAipsDebounced(1);
  if (aipStore.current?.uuid != event.uuid) return;
  aipStore.current.status = event.status;
}

function handleAipLocationUpdated(data: unknown) {
  const event = api.AIPLocationUpdatedEventFromJSON(data);
  const aipStore = useAipStore();
  aipStore.fetchAipsDebounced(1);
  if (aipStore.current?.uuid != event.uuid) return;
  aipStore.current.locationUuid = event.locationUuid;
}

function handleAipWorkflowCreated(data: unknown) {
  const event = api.AIPWorkflowCreatedEventFromJSON(data);
  const aipStore = useAipStore();

  // Ignore event if it does not relate to the current AIP.
  if (aipStore.current?.uuid != event.item.aipUuid) return;

  // Append the workflow.
  if (!aipStore.currentWorkflows) aipStore.currentWorkflows = {};
  if (!aipStore.currentWorkflows.workflows)
    aipStore.currentWorkflows.workflows = [];
  aipStore.currentWorkflows?.workflows?.unshift(event.item);
}

function handleAipWorkflowUpdated(data: unknown) {
  const event = api.AIPWorkflowUpdatedEventFromJSON(data);
  const aipStore = useAipStore();

  // Ignore event if it does not relate to the current AIP.
  if (aipStore.current?.uuid != event.item.aipUuid) return;

  // Find and update the workflow.
  const workflow = aipStore.getWorkflowById(event.uuid);
  if (!workflow) return;

  // Keep existing tasks, this event doesn't include them.
  const tasks = workflow.tasks;
  Object.assign(workflow, event.item);
  workflow.tasks = tasks;
}

function handleAipTaskCreated(data: unknown) {
  const event = api.AIPTaskCreatedEventFromJSON(data);
  const aipStore = useAipStore();

  // Find and update the workflow tasks.
  if (!event.item.workflowUuid) return;
  const workflow = aipStore.getWorkflowById(event.item.workflowUuid);
  if (!workflow) return;
  if (workflow.uuid === event.item.workflowUuid) {
    if (!workflow.tasks) workflow.tasks = [];
    workflow.tasks.push(event.item);
  }
}

function handleAipTaskUpdated(data: unknown) {
  const event = api.AIPTaskUpdatedEventFromJSON(data);
  const aipStore = useAipStore();

  // Find and update the task.
  if (!event.item.workflowUuid) return;
  const task = aipStore.getTaskById(event.item.workflowUuid, event.uuid);
  if (!task) return;
  Object.assign(task, event.item);
}
