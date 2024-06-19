import { api, client } from "@/client";
import { MonitorEventEventTypeEnum } from "@/openapi-generator";
import { useLayoutStore } from "@/stores/layout";
import router from "@/router";
import { defineStore, acceptHMRUpdate } from "pinia";
import { ref } from "vue";
import { mapKeys, snakeCase } from "lodash-es";

export const usePackageStore = defineStore("package", {
  state: () => ({
    // Package currently displayed.
    current: null as api.EnduroStoredPackage | null,

    // Preservation actions of the current package.
    current_preservation_actions:
      null as api.EnduroPackagePreservationActions | null,

    // The current package is being moved into a new location.
    // Set to true by this client when the package is moved.
    // Set to false by moveStatus or handlePackageLocationUpdated.
    locationChanging: false,

    // A list of packages shown during searches.
    packages: [] as Array<api.EnduroStoredPackage>,

    // Cursor for this page of packages.
    cursor: 0,

    // Cursor for next page of packages.
    nextCursor: 0,

    // A list of previous page cursors.
    prevCursors: [] as Array<number>,

    // User-interface interactions between components.
    ui: {
      download: new UIRequest(),
    },
  }),
  getters: {
    isPending(): boolean {
      return this.current?.status == api.EnduroStoredPackageStatusEnum.Pending;
    },
    isDone(): boolean {
      return this.current?.status == api.EnduroStoredPackageStatusEnum.Done;
    },
    isMovable(): boolean {
      return this.isDone && !this.isMoving;
    },
    isMoving(): boolean {
      return this.locationChanging;
    },
    isRejected(): boolean {
      return this.isDone && this.current?.locationId === undefined;
    },
    hasNextPage(): boolean {
      return this.nextCursor != 0;
    },
    hasPrevPage(): boolean {
      return this.prevCursors.length > 0;
    },
    getActionById: (state) => {
      return (
        actionId: number,
      ): api.EnduroPackagePreservationAction | undefined => {
        const x = state.current_preservation_actions?.actions?.find(
          (action) => action.id === actionId,
        );
        return x;
      };
    },
    getTaskById: (state) => {
      return (
        actionId: number,
        taskId: number,
      ): api.EnduroPackagePreservationTask | undefined => {
        const action = state.current_preservation_actions?.actions?.find(
          (action) => action.id === actionId,
        );
        if (!action) return;
        return action.tasks?.find((task) => task.id === taskId);
      };
    },
  },
  actions: {
    handleEvent(event: api.MonitorEventEvent) {
      const json = JSON.parse(event.value);
      const value = mapKeys(json, (value, key) => snakeCase(key));
      handlers[event.type](value);
    },
    async fetchCurrent(id: string) {
      this.$reset();

      const packageId = +id;
      if (Number.isNaN(packageId)) {
        throw Error("Unexpected parameter");
      }

      this.current = await client.package.packageShow({ id: packageId });

      // Update breadcrumb. TODO: should this be done in the component?
      const layoutStore = useLayoutStore();

      layoutStore.updateBreadcrumb([
        { route: router.resolve("/packages/"), text: "Packages" },
        { text: this.current.name },
      ]);

      await Promise.allSettled([
        client.package
          .packagePreservationActions({ id: packageId })
          .then((resp) => {
            this.current_preservation_actions = resp;
          }),
        client.package.packageMoveStatus({ id: packageId }).then((resp) => {
          this.locationChanging = !resp.done;
        }),
      ]);
    },
    async fetchPackages() {
      const resp = await client.package.packageList({
        cursor: this.cursor > 0 ? this.cursor.toString() : undefined,
      });
      this.packages = resp.items;
      this.nextCursor = Number(resp.nextCursor);
    },
    async fetchPackagesDebounced() {
      return this.fetchPackages();
    },
    async move(locationId: string) {
      if (!this.current) return;
      try {
        await client.package.packageMove({
          id: this.current.id,
          confirmRequestBody: { locationId: locationId },
        });
      } catch (error) {
        return error;
      }
      this.$patch((state) => {
        if (!state.current) return;
        state.current.status = api.EnduroStoredPackageStatusEnum.InProgress;
        state.locationChanging = true;
      });
    },
    async moveStatus() {
      if (!this.current) return;
      let resp;
      try {
        resp = await client.package.packageMoveStatus({
          id: this.current?.id,
        });
      } catch (error) {
        return error;
      }
      this.locationChanging = !resp.done;
    },
    confirm(locationId: string) {
      if (!this.current) return;
      client.package
        .packageConfirm({
          id: this.current.id,
          confirmRequestBody: { locationId: locationId },
        })
        .then((payload) => {
          if (!this.current) return;
          this.current.status = api.EnduroStoredPackageStatusEnum.InProgress;
        });
    },
    reject() {
      if (!this.current) return;
      client.package.packageReject({ id: this.current.id }).then((payload) => {
        if (!this.current) return;
        this.current.status = api.EnduroStoredPackageStatusEnum.InProgress;
      });
    },
    nextPage() {
      if (this.nextCursor == 0) {
        return;
      }
      this.prevCursors.push(this.cursor);
      this.cursor = this.nextCursor;
      this.fetchPackages();
    },
    prevPage() {
      let prev = this.prevCursors.pop();
      if (prev !== undefined) {
        this.cursor = prev;
        this.fetchPackages();
      }
    },
  },
  debounce: {
    fetchPackagesDebounced: [500, { isImmediate: false }],
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(usePackageStore, import.meta.hot));
}

const handlers: {
  [key in api.MonitorEventEventTypeEnum]: (data: any) => void;
} = {
  [MonitorEventEventTypeEnum.MonitorPingEvent]: handleMonitorPing,
  [MonitorEventEventTypeEnum.PackageCreatedEvent]: handlePackageCreated,
  [MonitorEventEventTypeEnum.PackageUpdatedEvent]: handlePackageUpdated,
  [MonitorEventEventTypeEnum.PackageStatusUpdatedEvent]:
    handlePackageStatusUpdated,
  [MonitorEventEventTypeEnum.PackageLocationUpdatedEvent]:
    handlePackageLocationUpdated,
  [MonitorEventEventTypeEnum.PreservationActionCreatedEvent]:
    handlePreservationActionCreated,
  [MonitorEventEventTypeEnum.PreservationActionUpdatedEvent]:
    handlePreservationActionUpdated,
  [MonitorEventEventTypeEnum.PreservationTaskCreatedEvent]:
    handlePreservationTaskCreated,
  [MonitorEventEventTypeEnum.PreservationTaskUpdatedEvent]:
    handlePreservationTaskUpdated,
};

function handleMonitorPing(data: any) {
  const event = api.MonitorPingEventFromJSON(data);
}

function handlePackageCreated(data: any) {
  const event = api.PackageCreatedEventFromJSON(data);
  const store = usePackageStore();
  store.fetchPackagesDebounced();
}

function handlePackageUpdated(data: any) {
  const event = api.PackageUpdatedEventFromJSON(data);
  const store = usePackageStore();
  store.fetchPackagesDebounced();
  if (store.$state.current?.id != event.id) return;
  Object.assign(store.$state.current, event.item);
}

function handlePackageStatusUpdated(data: any) {
  const event = api.PackageStatusUpdatedEventFromJSON(data);
  const store = usePackageStore();
  store.fetchPackagesDebounced();
  if (store.$state.current?.id != event.id) return;
  store.$state.current.status = event.status;
}

function handlePackageLocationUpdated(data: any) {
  const event = api.PackageLocationUpdatedEventFromJSON(data);
  const store = usePackageStore();
  store.fetchPackagesDebounced();
  store.$patch((state) => {
    if (state.current?.id != event.id) return;
    state.current.locationId = event.locationId;
    state.locationChanging = false;
  });
}

function handlePreservationActionCreated(data: any) {
  const event = api.PreservationActionCreatedEventFromJSON(data);
  const store = usePackageStore();

  // Ignore event if it does not relate to the current package.
  if (store.current?.id != event.item.packageId) return;

  // Append the action.
  store.current_preservation_actions?.actions?.unshift(event.item);
}

function handlePreservationActionUpdated(data: any) {
  const event = api.PreservationActionUpdatedEventFromJSON(data);
  const store = usePackageStore();

  // Ignore event if it does not relate to the current package.
  if (store.current?.id != event.item.packageId) return;

  // Find and update the action.
  const action = store.getActionById(event.id);
  if (!action) return;
  Object.assign(action, event.item);
}

function handlePreservationTaskCreated(data: any) {
  const event = api.PreservationTaskCreatedEventFromJSON(data);
  const store = usePackageStore();

  // Find and update the action.
  if (!event.item.preservationActionId) return;
  const action = store.getActionById(event.item.preservationActionId);
  if (!action) return;
  if (action.id === event.item.preservationActionId) {
    if (!action.tasks) action.tasks = [];
    action.tasks.push(event.item);
  }
}

function handlePreservationTaskUpdated(data: any) {
  const event = api.PreservationTaskUpdatedEventFromJSON(data);
  const store = usePackageStore();

  if (!event.item.preservationActionId) return;
  const task = store.getTaskById(event.item.preservationActionId, event.id);
  if (!task) return;
  Object.assign(task, event.item);
}

class UIRequest {
  inner = ref(0);
  request = () => this.inner.value++;
}
