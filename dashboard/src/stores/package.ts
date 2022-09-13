import { api, client } from "@/client";
import { useLayoutStore } from "@/stores/layout";
import { defineStore, acceptHMRUpdate } from "pinia";
import { ref } from "vue";

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
    getActionById: (state) => {
      return (
        actionId: number
      ): api.EnduroPackagePreservationAction | undefined => {
        const x = state.current_preservation_actions?.actions?.find(
          (action) => action.id === actionId
        );
        return x;
      };
    },
    getTaskById: (state) => {
      return (
        actionId: number,
        taskId: number
      ): api.EnduroPackagePreservationTask | undefined => {
        const action = state.current_preservation_actions?.actions?.find(
          (action) => action.id === actionId
        );
        if (!action) return;
        return action.tasks?.find((task) => task.id === taskId);
      };
    },
  },
  actions: {
    handleEvent(event: api.MonitorResponseBody) {
      let key: keyof api.MonitorResponseBody;
      for (key in event) {
        const payload: any = event[key];
        if (!payload) continue;
        const handler = handlers[key];
        handler(payload);
        break;
      }
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
        { routeName: "packages", text: "Packages" },
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
      const resp = await client.package.packageList();
      this.packages = resp.items;
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
  },
  debounce: {
    fetchPackagesDebounced: [500, { isImmediate: false }],
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(usePackageStore, import.meta.hot));
}

// Event handler function.
type Handler<T> = (event: T) => void;

// Utility that constructs a type with all properties of T set to required,
// mapping them to the event handlers.
type Partial<T> = {
  [P in keyof T]-?: Handler<NonNullable<T[P]>>;
};

// Pairs all known events with the event handlers.
const handlers: Partial<api.MonitorResponseBody> = {
  monitorPingEvent: handleMonitorPing,
  packageCreatedEvent: handlePackageCreated,
  packageLocationUpdatedEvent: handlePackageLocationUpdated,
  packageStatusUpdatedEvent: handlePackageStatusUpdated,
  packageUpdatedEvent: handlePackageUpdated,
  preservationActionCreatedEvent: handlePreservationActionCreated,
  preservationActionUpdatedEvent: handlePreservationActionUpdated,
  preservationTaskCreatedEvent: handlePreservationTaskCreated,
  preservationTaskUpdatedEvent: handlePreservationTaskUpdated,
};

function handleMonitorPing(event: api.EnduroMonitorPingEventResponseBody) {}

function handlePackageCreated(
  event: api.EnduroPackageCreatedEventResponseBody
) {
  const store = usePackageStore();
  store.fetchPackagesDebounced();
}

function handlePackageUpdated(
  event: api.EnduroPackageUpdatedEventResponseBody
) {
  const store = usePackageStore();
  store.fetchPackagesDebounced();
  if (store.$state.current?.id != event.id) return;
  Object.assign(store.$state.current, event.item);
}

function handlePackageStatusUpdated(
  event: api.EnduroPackageStatusUpdatedEventResponseBody
) {
  const store = usePackageStore();
  store.fetchPackagesDebounced();
  if (store.$state.current?.id != event.id) return;
  store.$state.current.status = event.status;
}

function handlePackageLocationUpdated(
  event: api.EnduroPackageLocationUpdatedEventResponseBody
) {
  const store = usePackageStore();
  store.fetchPackagesDebounced();
  store.$patch((state) => {
    if (state.current?.id != event.id) return;
    state.current.locationId = event.locationId;
    state.locationChanging = false;
  });
}

function handlePreservationActionCreated(
  event: api.EnduroPreservationActionCreatedEventResponseBody
) {
  const store = usePackageStore();

  // Ignore event if it does not relate to the current package.
  if (store.current?.id != event.item.packageId) return;

  // Append the action.
  store.current_preservation_actions?.actions?.unshift(event.item);
}

function handlePreservationActionUpdated(
  event: api.EnduroPreservationActionUpdatedEventResponseBody
) {
  const store = usePackageStore();

  // Ignore event if it does not relate to the current package.
  if (store.current?.id != event.item.packageId) return;

  // Find and update the action.
  const action = store.getActionById(event.id);
  if (!action) return;
  Object.assign(action, event.item);
}

function handlePreservationTaskCreated(
  event: api.EnduroPreservationTaskCreatedEventResponseBody
) {
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

function handlePreservationTaskUpdated(
  event: api.EnduroPreservationTaskUpdatedEventResponseBody
) {
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
