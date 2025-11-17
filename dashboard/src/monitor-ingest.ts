import { api } from "@/client";
import { transformKeys } from "@/helpers/transform";
import { IngestEvent2ValueTypeEnum } from "@/openapi-generator";
import { useSipStore } from "@/stores/sip";

export function handleIngestEvent(event: api.IngestEvent2Value) {
  handlers[event.type](transformKeys(event.value));
}

const handlers: {
  [key in api.IngestEvent2ValueTypeEnum]: (data: unknown) => void;
} = {
  [IngestEvent2ValueTypeEnum.IngestPingEvent]: () => {},
  [IngestEvent2ValueTypeEnum.SipCreatedEvent]: handleSipCreated,
  [IngestEvent2ValueTypeEnum.SipUpdatedEvent]: handleSipUpdated,
  [IngestEvent2ValueTypeEnum.SipStatusUpdatedEvent]: handleSipStatusUpdated,
  [IngestEvent2ValueTypeEnum.SipWorkflowCreatedEvent]: handleSipWorkflowCreated,
  [IngestEvent2ValueTypeEnum.SipWorkflowUpdatedEvent]: handleSipWorkflowUpdated,
  [IngestEvent2ValueTypeEnum.SipTaskCreatedEvent]: handleSipTaskCreated,
  [IngestEvent2ValueTypeEnum.SipTaskUpdatedEvent]: handleSipTaskUpdated,
  [IngestEvent2ValueTypeEnum.BatchCreatedEvent]: handleBatchCreated,
  [IngestEvent2ValueTypeEnum.BatchUpdatedEvent]: handleBatchUpdated,
};

function handleSipCreated() {
  const store = useSipStore();
  store.fetchSipsDebounced(1);
}

function handleSipUpdated(data: unknown) {
  const event = api.SIPUpdatedEventFromJSON(data);
  const store = useSipStore();
  store.fetchSipsDebounced(1);
  if (store.current?.uuid != event.uuid) return;
  Object.assign(store.current, event.item);
}

function handleSipStatusUpdated(data: unknown) {
  const event = api.SIPStatusUpdatedEventFromJSON(data);
  const store = useSipStore();
  store.fetchSipsDebounced(1);
  if (store.current?.uuid != event.uuid) return;
  store.current.status = event.status;
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

function handleBatchCreated(data: unknown) {
  console.log("Batch created event received:", data);
  // TODO: add batch store and update it here.
  // const store = useBatchStore();
  // store.fetchBatchesDebounced(1);
}

function handleBatchUpdated(data: unknown) {
  console.log("Batch updated event received:", data);
  // TODO: add batch store and update it here.
  // const event = api.BatchUpdatedEventFromJSON(data);
  // const store = useBatchStore();
  // store.fetchBatchesDebounced(1);
  // if (store.current?.uuid != event.uuid) return;
  // Object.assign(store.current, event.item);
}
