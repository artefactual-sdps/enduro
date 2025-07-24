import { api } from "@/client";
import { transformKeys } from "@/helpers/transform";
import { StorageMonitorEventStorageEventTypeEnum } from "@/openapi-generator";
import { useAipStore } from "@/stores/aip";
import { useLocationStore } from "@/stores/location";

export function handleStorageEvent(event: api.StorageMonitorEventStorageEvent) {
  handlers[event.type](transformKeys(event.value));
}

const handlers: {
  [key in api.StorageMonitorEventStorageEventTypeEnum]: (data: unknown) => void;
} = {
  [StorageMonitorEventStorageEventTypeEnum.StoragePingEvent]: () => {},
  [StorageMonitorEventStorageEventTypeEnum.LocationCreatedEvent]:
    handleLocationCreated,
  [StorageMonitorEventStorageEventTypeEnum.LocationUpdatedEvent]:
    handleLocationUpdated,
  [StorageMonitorEventStorageEventTypeEnum.AipCreatedEvent]: handleAipCreated,
  [StorageMonitorEventStorageEventTypeEnum.AipUpdatedEvent]: handleAipUpdated,
  [StorageMonitorEventStorageEventTypeEnum.AipWorkflowCreatedEvent]:
    handleAipWorkflowCreated,
  [StorageMonitorEventStorageEventTypeEnum.AipWorkflowUpdatedEvent]:
    handleAipWorkflowUpdated,
  [StorageMonitorEventStorageEventTypeEnum.AipTaskCreatedEvent]:
    handleAipTaskCreated,
  [StorageMonitorEventStorageEventTypeEnum.AipTaskUpdatedEvent]:
    handleAipTaskUpdated,
};

function handleLocationCreated() {
  const locationStore = useLocationStore();
  locationStore.fetchLocationsDebounced();
}

function handleLocationUpdated(data: unknown) {
  const event = api.LocationUpdatedEventFromJSON(data);
  const locationStore = useLocationStore();
  locationStore.fetchLocationsDebounced();
  if (locationStore.current?.uuid != event.uuid) return;
  Object.assign(locationStore.current, event.item);
}

function handleAipCreated() {
  const aipStore = useAipStore();
  aipStore.fetchAipsDebounced(1);
  // TODO: Consider AIPs shown in the location page.
}

function handleAipUpdated(data: unknown) {
  const event = api.AIPUpdatedEventFromJSON(data);
  const aipStore = useAipStore();
  aipStore.fetchAipsDebounced(1);
  if (aipStore.current?.uuid != event.uuid) return;
  Object.assign(aipStore.current, event.item);
  // TODO: Consider AIPs shown in the location page.
}

function handleAipWorkflowCreated(data: unknown) {
  const event = api.AIPWorkflowCreatedEventFromJSON(data);
  const aipStore = useAipStore();

  // Ignore event if it does not relate to the current AIP.
  // TODO: Include AIP UUID in the workflows.
  // if (aipStore.current?.uuid != event.item.aipUuid) return;

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
  // TODO: Include AIP UUID in the workflows.
  // if (aipStore.current?.uuid != event.item.aipUuid) return;

  // Find and update the workflow.
  const workflow = aipStore.getWorkflowById(event.uuid);
  if (!workflow) return;

  // Keep existing tasks, this event doesn't include them.
  const tasks = workflow.tasks;
  Object.assign(workflow, event.item);
  workflow.tasks = tasks;
}

function handleAipTaskCreated() {
  // const event = api.AIPTaskCreatedEventFromJSON(data);
  // const aipStore = useAipStore();
  // Find and update the workflow.
  // TODO: Include Workflow UUID in the tasks.
  // if (!event.item.workflowUuid) return;
  // const workflow = aipStore.getWorkflowById(event.item.workflowUuid);
  // if (!workflow) return;
  // if (workflow.uuid === event.item.workflowUuid) {
  //   if (!workflow.tasks) workflow.tasks = [];
  //   workflow.tasks.push(event.item);
  // }
}

function handleAipTaskUpdated() {
  // const event = api.AIPTaskUpdatedEventFromJSON(data);
  // const aipStore = useAipStore();
  // TODO: Include Workflow UUID in the tasks.
  // if (!event.item.workflowUuid) return;
  // const task = aipStore.getTaskById(event.item.workflowUuid, event.uuid);
  // if (!task) return;
  // Object.assign(task, event.item);
}
