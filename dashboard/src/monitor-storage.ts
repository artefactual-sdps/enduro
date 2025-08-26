import { api } from "@/client";
import { transformKeys } from "@/helpers/transform";
import { StorageEventStorageValueTypeEnum } from "@/openapi-generator";
import { useAipStore } from "@/stores/aip";
import { useLocationStore } from "@/stores/location";

export function handleStorageEvent(event: api.StorageEventStorageValue) {
  handlers[event.type](transformKeys(event.value));
}

const handlers: {
  [key in api.StorageEventStorageValueTypeEnum]: (data: unknown) => void;
} = {
  [StorageEventStorageValueTypeEnum.StoragePingEvent]: () => {},
  [StorageEventStorageValueTypeEnum.LocationCreatedEvent]:
    handleLocationCreated,
  [StorageEventStorageValueTypeEnum.AipCreatedEvent]: handleAipCreated,
  [StorageEventStorageValueTypeEnum.AipStatusUpdatedEvent]:
    handleAipStatusUpdated,
  [StorageEventStorageValueTypeEnum.AipLocationUpdatedEvent]:
    handleAipLocationUpdated,
  [StorageEventStorageValueTypeEnum.AipWorkflowCreatedEvent]:
    handleAipWorkflowCreated,
  [StorageEventStorageValueTypeEnum.AipWorkflowUpdatedEvent]:
    handleAipWorkflowUpdated,
  [StorageEventStorageValueTypeEnum.AipTaskCreatedEvent]: handleAipTaskCreated,
  [StorageEventStorageValueTypeEnum.AipTaskUpdatedEvent]: handleAipTaskUpdated,
  [StorageEventStorageValueTypeEnum.AipDeletionRequestCreatedEvent]:
    handleAipDeletionRequestCreated,
  [StorageEventStorageValueTypeEnum.AipDeletionRequestUpdatedEvent]:
    handleAipDeletionRequestUpdated,
};

function handleLocationCreated() {
  const locationStore = useLocationStore();
  locationStore.fetchLocationsDebounced();
}

function handleAipCreated() {
  const aipStore = useAipStore();
  aipStore.fetchAipsDebounced(1);
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

function handleAipDeletionRequestCreated() {
  // We aren't directly showing deletion requests in the UI, so there's nothing
  // to update yet.
}

function handleAipDeletionRequestUpdated() {
  // We aren't directly showing deletion requests in the UI, so there's nothing
  // to update yet.
}
