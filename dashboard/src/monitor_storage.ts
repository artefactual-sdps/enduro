import { mapKeys, snakeCase } from "lodash-es";

import { api } from "@/client";
import { useAipStore } from "@/stores/aip";
import { useLocationStore } from "@/stores/location";

export function handleStorageEvent(eventData: unknown) {
  // Parse the event data to determine the event type
  // The storage monitor events come as raw JSON with event type information
  if (!eventData || typeof eventData !== 'object') {
    return;
  }

  // Transform keys to snake_case (following the pattern from monitor.ts)
  const value = mapKeys(eventData, (_, key) => snakeCase(key));
  if (value.item) {
    value.item = mapKeys(value.item, (_, key) => snakeCase(key));
  }

  // Determine event type from the data structure
  // Storage events are direct objects, not wrapped like ingest events
  if (value.storage_ping_event !== undefined) {
    handleStoragePing(value.storage_ping_event);
  } else if (value.location_created_event !== undefined) {
    handleLocationCreated(value.location_created_event);
  } else if (value.location_updated_event !== undefined) {
    handleLocationUpdated(value.location_updated_event);
  } else if (value.aip_created_event !== undefined) {
    handleAIPCreated(value.aip_created_event);
  } else if (value.aip_updated_event !== undefined) {
    handleAIPUpdated(value.aip_updated_event);
  } else if (value.aip_workflow_created_event !== undefined) {
    handleAIPWorkflowCreated(value.aip_workflow_created_event);
  } else if (value.aip_workflow_updated_event !== undefined) {
    handleAIPWorkflowUpdated(value.aip_workflow_updated_event);
  } else if (value.aip_task_created_event !== undefined) {
    handleAIPTaskCreated(value.aip_task_created_event);
  } else if (value.aip_task_updated_event !== undefined) {
    handleAIPTaskUpdated(value.aip_task_updated_event);
  }
}

function handleStoragePing(data: unknown) {
  // Parse ping event for heartbeat - no action needed
  api.StoragePingEventFromJSON(data);
}

function handleLocationCreated(data: unknown) {
  // Parse and handle location creation
  api.LocationCreatedEventFromJSON(data);
  const locationStore = useLocationStore();
  
  // Refresh locations list to include the new location
  locationStore.fetchLocationsDebounced();
}

function handleLocationUpdated(data: unknown) {
  // Parse and handle location updates
  const event = api.LocationUpdatedEventFromJSON(data);
  const locationStore = useLocationStore();
  
  // If this location is currently displayed, update it
  if (locationStore.$state.current?.uuid === event.uuid) {
    if (event.item && locationStore.$state.current) {
      Object.assign(locationStore.$state.current, event.item);
    }
  }
  
  // Refresh locations list to reflect changes
  locationStore.fetchLocationsDebounced();
}

function handleAIPCreated(data: unknown) {
  // Parse and handle AIP creation
  const event = api.AIPCreatedEventFromJSON(data);
  const aipStore = useAipStore();
  const locationStore = useLocationStore();
  
  // Refresh AIP list to include the new AIP
  aipStore.fetchAipsDebounced(1);
  
  // If showing a location that now has this AIP, refresh location's AIPs
  if (event.item?.locationId && locationStore.$state.current?.uuid === event.item.locationId) {
    locationStore.fetchCurrentDebounced(locationStore.$state.current.uuid);
  }
}

function handleAIPUpdated(data: unknown) {
  // Parse and handle AIP updates
  const event = api.AIPUpdatedEventFromJSON(data);
  const aipStore = useAipStore();
  const locationStore = useLocationStore();
  
  // If this AIP is currently displayed, update it
  if (aipStore.$state.current?.uuid === event.uuid) {
    if (event.item && aipStore.$state.current) {
      Object.assign(aipStore.$state.current, event.item);
    }
  }
  
  // Refresh AIP list to reflect changes
  aipStore.fetchAipsDebounced(Math.floor((aipStore.$state.page.offset || 0) / aipStore.$state.page.limit) + 1);
  
  // If showing a location that contains this AIP, refresh location's AIPs
  if (event.item?.locationId && locationStore.$state.current?.uuid === event.item.locationId) {
    locationStore.fetchCurrentDebounced(locationStore.$state.current.uuid);
  }
}

function handleAIPWorkflowCreated(data: unknown) {
  // Parse and handle AIP workflow creation
  api.AIPWorkflowCreatedEventFromJSON(data);
  const aipStore = useAipStore();
  
  // For now, just refresh workflows for the current AIP
  if (aipStore.current?.uuid) {
    aipStore.fetchWorkflows(aipStore.current.uuid);
  }
}

function handleAIPWorkflowUpdated(data: unknown) {
  // Parse and handle AIP workflow updates
  api.AIPWorkflowUpdatedEventFromJSON(data);
  const aipStore = useAipStore();
  
  // For now, just refresh workflows for the current AIP
  if (aipStore.current?.uuid) {
    aipStore.fetchWorkflows(aipStore.current.uuid);
  }
}

function handleAIPTaskCreated(data: unknown) {
  // Parse and handle AIP task creation
  api.AIPTaskCreatedEventFromJSON(data);
  const aipStore = useAipStore();
  
  // For now, just refresh workflows for the current AIP
  if (aipStore.current?.uuid) {
    aipStore.fetchWorkflows(aipStore.current.uuid);
  }
}

function handleAIPTaskUpdated(data: unknown) {
  // Parse and handle AIP task updates
  api.AIPTaskUpdatedEventFromJSON(data);
  const aipStore = useAipStore();
  
  // For now, just refresh workflows for the current AIP
  if (aipStore.current?.uuid) {
    aipStore.fetchWorkflows(aipStore.current.uuid);
  }
}