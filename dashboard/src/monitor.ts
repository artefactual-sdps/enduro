import { mapKeys, snakeCase } from "lodash-es";

import { api } from "@/client";
import { MonitorEventEventTypeEnum } from "@/openapi-generator";
import { useSipStore } from "@/stores/sip";

export function handleEvent(event: api.MonitorEventEvent) {
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
  [key in api.MonitorEventEventTypeEnum]: (data: unknown) => void;
} = {
  [MonitorEventEventTypeEnum.MonitorPingEvent]: handleMonitorPing,
  [MonitorEventEventTypeEnum.SipCreatedEvent]: handleSipCreated,
  [MonitorEventEventTypeEnum.SipUpdatedEvent]: handleSipUpdated,
  [MonitorEventEventTypeEnum.SipStatusUpdatedEvent]: handleSipStatusUpdated,
  [MonitorEventEventTypeEnum.SipLocationUpdatedEvent]: handleSipLocationUpdated,
  [MonitorEventEventTypeEnum.SipPreservationActionCreatedEvent]:
    handlePreservationActionCreated,
  [MonitorEventEventTypeEnum.SipPreservationActionUpdatedEvent]:
    handlePreservationActionUpdated,
  [MonitorEventEventTypeEnum.SipPreservationTaskCreatedEvent]:
    handlePreservationTaskCreated,
  [MonitorEventEventTypeEnum.SipPreservationTaskUpdatedEvent]:
    handlePreservationTaskUpdated,
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
  if (store.$state.current?.id != event.id) return;
  Object.assign(store.$state.current, event.item);
}

function handleSipStatusUpdated(data: unknown) {
  const event = api.SIPStatusUpdatedEventFromJSON(data);
  const store = useSipStore();
  store.fetchSipsDebounced(1);
  if (store.$state.current?.id != event.id) return;
  store.$state.current.status = event.status;
}

function handleSipLocationUpdated(data: unknown) {
  const event = api.SIPLocationUpdatedEventFromJSON(data);
  const store = useSipStore();
  store.fetchSipsDebounced(1);
  store.$patch((state) => {
    if (state.current?.id != event.id) return;
    state.current.locationId = event.locationId;
  });
}

function handlePreservationActionCreated(data: unknown) {
  const event = api.SIPPreservationActionCreatedEventFromJSON(data);
  const store = useSipStore();

  // Ignore event if it does not relate to the current SIP.
  if (store.current?.id != event.item.sipId) return;

  // Append the action.
  store.currentPreservationActions?.actions?.unshift(event.item);
}

function handlePreservationActionUpdated(data: unknown) {
  const event = api.SIPPreservationActionUpdatedEventFromJSON(data);
  const store = useSipStore();

  // Ignore event if it does not relate to the current SIP.
  if (store.current?.id != event.item.sipId) return;

  // Find and update the action.
  const action = store.getActionById(event.id);
  if (!action) return;

  // Keep existing tasks, this event doesn't include them.
  const tasks = action.tasks;
  Object.assign(action, event.item);
  action.tasks = tasks;
}

function handlePreservationTaskCreated(data: unknown) {
  const event = api.SIPPreservationTaskCreatedEventFromJSON(data);
  const store = useSipStore();

  // Find and update the action.
  if (!event.item.preservationActionId) return;
  const action = store.getActionById(event.item.preservationActionId);
  if (!action) return;
  if (action.id === event.item.preservationActionId) {
    if (!action.tasks) action.tasks = [];
    action.tasks.push(event.item);
  }
}

function handlePreservationTaskUpdated(data: unknown) {
  const event = api.SIPPreservationTaskUpdatedEventFromJSON(data);
  const store = useSipStore();

  if (!event.item.preservationActionId) return;
  const task = store.getTaskById(event.item.preservationActionId, event.id);
  if (!task) return;
  Object.assign(task, event.item);
}
