import { api } from "@/client";
import { IngestMonitorEventIngestEventTypeEnum } from "@/openapi-generator";
import { useSipStore } from "@/stores/sip";

export function handleIngestEvent(event: api.IngestMonitorEventIngestEvent) {
  const json = JSON.parse(event.value);
  // TODO: avoid key transformation in the backend or make
  // this fully recursive, considering objects and slices.
  const value = mapKeys(json, (_, key) => snakeCase(key));
  if (value.item) {
    value.item = mapKeys(value.item, (_, key) => snakeCase(key));
  }
  handlers[event.type](value);
}

const handlers: {
  [key in api.IngestMonitorEventIngestEventTypeEnum]: (data: unknown) => void;
} = {
  [IngestMonitorEventIngestEventTypeEnum.IngestPingEvent]: handleIngestPing,
  [IngestMonitorEventIngestEventTypeEnum.SipCreatedEvent]: handleSipCreated,
  [IngestMonitorEventIngestEventTypeEnum.SipUpdatedEvent]: handleSipUpdated,
  [IngestMonitorEventIngestEventTypeEnum.SipStatusUpdatedEvent]:
    handleSipStatusUpdated,
  [IngestMonitorEventIngestEventTypeEnum.SipWorkflowCreatedEvent]:
    handleSipWorkflowCreated,
  [IngestMonitorEventIngestEventTypeEnum.SipWorkflowUpdatedEvent]:
    handleSipWorkflowUpdated,
  [IngestMonitorEventIngestEventTypeEnum.SipTaskCreatedEvent]:
    handleSipTaskCreated,
  [IngestMonitorEventIngestEventTypeEnum.SipTaskUpdatedEvent]:
    handleSipTaskUpdated,
};

function handleIngestPing(data: unknown) {
  api.IngestPingEventFromJSON(data);
}

function handleSipCreated(data: unknown) {
  api.SIPCreatedEventFromJSON(data);
  const store = useSipStore();
  store.fetchSipsDebounced(1);
}

function handleSipUpdated(data: unknown) {
  const event = api.SIPUpdatedEventFromJSON(data);
  const store = useSipStore();
  store.fetchSipsDebounced(1);
  if (store.$state.current?.uuid != event.uuid) return;
  Object.assign(store.$state.current, event.item);
}

function handleSipStatusUpdated(data: unknown) {
  const event = api.SIPStatusUpdatedEventFromJSON(data);
  const store = useSipStore();
  store.fetchSipsDebounced(1);
  if (store.$state.current?.uuid != event.uuid) return;
  store.$state.current.status = event.status;
}

function handleSipWorkflowCreated(data: unknown) {
  const event = api.SIPWorkflowCreatedEventFromJSON(data);
  const store = useSipStore();

  // Ignore event if it does not relate to the current SIP.
  if (store.current?.uuid != event.item.sipUuid) return;

  // Append the workflow.
  if (!store.currentWorkflows) store.currentWorkflows = {};
  if (!store.currentWorkflows.workflows) store.currentWorkflows.workflows = [];
  store.currentWorkflows?.workflows?.unshift(event.item);
}

function handleSipWorkflowUpdated(data: unknown) {
  const event = api.SIPWorkflowUpdatedEventFromJSON(data);
  const store = useSipStore();

  // Ignore event if it does not relate to the current SIP.
  if (store.current?.uuid != event.item.sipUuid) return;

  // Find and update the workflow.
  const workflow = store.getWorkflowById(event.uuid);
  if (!workflow) return;

  // Keep existing tasks, this event doesn't include them.
  const tasks = workflow.tasks;
  Object.assign(workflow, event.item);
  workflow.tasks = tasks;
}

function handleSipTaskCreated(data: unknown) {
  const event = api.SIPTaskCreatedEventFromJSON(data);
  const store = useSipStore();

  // Find and update the workflow.
  if (!event.item.workflowUuid) return;
  const workflow = store.getWorkflowById(event.item.workflowUuid);
  if (!workflow) return;
  if (workflow.uuid === event.item.workflowUuid) {
    if (!workflow.tasks) workflow.tasks = [];
    workflow.tasks.push(event.item);
  }
}

function handleSipTaskUpdated(data: unknown) {
  const event = api.SIPTaskUpdatedEventFromJSON(data);
  const store = useSipStore();

  if (!event.item.workflowUuid) return;
  const task = store.getTaskById(event.item.workflowUuid, event.uuid);
  if (!task) return;
  Object.assign(task, event.item);
}
