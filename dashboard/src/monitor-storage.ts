import { mapKeys, snakeCase } from "lodash-es";

import { api } from "@/client";
import { StorageMonitorEventStorageEventTypeEnum } from "@/openapi-generator";
import { useAipStore } from "@/stores/aip";
import { useLocationStore } from "@/stores/location";

export function handleStorageEvent(event: api.StorageMonitorEventStorageEvent) {
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
  [key in api.StorageMonitorEventStorageEventTypeEnum]: (data: unknown) => void;
} = {
  [StorageMonitorEventStorageEventTypeEnum.StoragePingEvent]: handleStoragePing,
  [StorageMonitorEventStorageEventTypeEnum.LocationCreatedEvent]:
    handleLocationCreated,
  [StorageMonitorEventStorageEventTypeEnum.LocationUpdatedEvent]:
    handleLocationUpdated,
  [StorageMonitorEventStorageEventTypeEnum.AipCreatedEvent]: handleAIPCreated,
  [StorageMonitorEventStorageEventTypeEnum.AipUpdatedEvent]: handleAIPUpdated,
  [StorageMonitorEventStorageEventTypeEnum.AipWorkflowCreatedEvent]:
    handleAIPWorkflowCreated,
  [StorageMonitorEventStorageEventTypeEnum.AipWorkflowUpdatedEvent]:
    handleAIPWorkflowUpdated,
  [StorageMonitorEventStorageEventTypeEnum.AipTaskCreatedEvent]:
    handleAIPTaskCreated,
  [StorageMonitorEventStorageEventTypeEnum.AipTaskUpdatedEvent]:
    handleAIPTaskUpdated,
};

function handleStoragePing(data: unknown) {
  api.StoragePingEventFromJSON(data);
}

function handleLocationCreated(data: unknown) {
  api.LocationCreatedEventFromJSON(data);
  const locationStore = useLocationStore();
  locationStore.fetchLocationsDebounced();
}

function handleLocationUpdated(data: unknown) {
  const event = api.LocationUpdatedEventFromJSON(data);
  const locationStore = useLocationStore();
  locationStore.fetchLocationsDebounced();
  if (locationStore.$state.current?.uuid != event.uuid) return;
  if (event.item) {
    Object.assign(locationStore.$state.current, event.item);
  }
}

function handleAIPCreated(data: unknown) {
  const event = api.AIPCreatedEventFromJSON(data);
  const aipStore = useAipStore();
  const locationStore = useLocationStore();

  aipStore.fetchAipsDebounced(1);

  if (
    event.item?.locationId &&
    locationStore.$state.current?.uuid === event.item.locationId
  ) {
    locationStore.fetchCurrentDebounced(locationStore.$state.current.uuid);
  }
}

function handleAIPUpdated(data: unknown) {
  const event = api.AIPUpdatedEventFromJSON(data);
  const aipStore = useAipStore();

  aipStore.fetchAipsDebounced(
    Math.floor(
      (aipStore.$state.page.offset || 0) / aipStore.$state.page.limit,
    ) + 1,
  );
  if (aipStore.$state.current?.uuid != event.uuid) return;
  if (event.item) {
    Object.assign(aipStore.$state.current, event.item);
  }
}

function handleAIPWorkflowCreated(data: unknown) {
  api.AIPWorkflowCreatedEventFromJSON(data);
  const aipStore = useAipStore();

  if (aipStore.current?.uuid) {
    aipStore.fetchWorkflows(aipStore.current.uuid);
  }
}

function handleAIPWorkflowUpdated(data: unknown) {
  api.AIPWorkflowUpdatedEventFromJSON(data);
  const aipStore = useAipStore();

  if (aipStore.current?.uuid) {
    aipStore.fetchWorkflows(aipStore.current.uuid);
  }
}

function handleAIPTaskCreated(data: unknown) {
  api.AIPTaskCreatedEventFromJSON(data);
  const aipStore = useAipStore();

  if (aipStore.current?.uuid) {
    aipStore.fetchWorkflows(aipStore.current.uuid);
  }
}

function handleAIPTaskUpdated(data: unknown) {
  api.AIPTaskUpdatedEventFromJSON(data);
  const aipStore = useAipStore();

  if (aipStore.current?.uuid) {
    aipStore.fetchWorkflows(aipStore.current.uuid);
  }
}
