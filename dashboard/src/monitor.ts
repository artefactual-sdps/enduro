import { api } from "@/client";
import { transformKeys } from "@/helpers/transform";
import { MonitorEventEventTypeEnum } from "@/openapi-generator";
import { useSipStore } from "@/stores/sip";

export function handleEvent(event: api.MonitorEventEvent) {
  handlers[event.type](transformKeys(event.value));
}

const handlers: {
  [key in api.MonitorEventEventTypeEnum]: (data: unknown) => void;
} = {
  [MonitorEventEventTypeEnum.MonitorPingEvent]: handleMonitorPing,
  [MonitorEventEventTypeEnum.SipCreatedEvent]: handleSipCreated,
  [MonitorEventEventTypeEnum.SipUpdatedEvent]: handleSipUpdated,
  [MonitorEventEventTypeEnum.SipStatusUpdatedEvent]: handleSipStatusUpdated,
  [MonitorEventEventTypeEnum.SipWorkflowCreatedEvent]: handleWorkflowCreated,
  [MonitorEventEventTypeEnum.SipWorkflowUpdatedEvent]: handleWorkflowUpdated,
  [MonitorEventEventTypeEnum.SipTaskCreatedEvent]: handleTaskCreated,
  [MonitorEventEventTypeEnum.SipTaskUpdatedEvent]: handleTaskUpdated,
};

function handleMonitorPing(data: unknown) {
  api.MonitorPingEventFromJSON(data);
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

function handleWorkflowCreated(data: unknown) {
  const event = api.SIPWorkflowCreatedEventFromJSON(data);
  const store = useSipStore();

  // Ignore event if it does not relate to the current SIP.
  if (store.current?.uuid != event.item.sipUuid) return;

  // Append the workflow.
  if (!store.currentWorkflows) store.currentWorkflows = {};
  if (!store.currentWorkflows.workflows) store.currentWorkflows.workflows = [];
  store.currentWorkflows?.workflows?.unshift(event.item);
}

function handleWorkflowUpdated(data: unknown) {
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

function handleTaskCreated(data: unknown) {
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

function handleTaskUpdated(data: unknown) {
  const event = api.SIPTaskUpdatedEventFromJSON(data);
  const store = useSipStore();

  if (!event.item.workflowUuid) return;
  const task = store.getTaskById(event.item.workflowUuid, event.uuid);
  if (!task) return;
  Object.assign(task, event.item);
}
