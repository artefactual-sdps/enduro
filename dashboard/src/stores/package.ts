import { api, client } from "@/client";
import { defineStore, acceptHMRUpdate } from "pinia";
import { ref } from "vue";

export const usePackageStore = defineStore("package", {
  state: () => ({
    // Package currently displayed.
    current: null as api.PackageShowResponseBody | null,

    // Preservation actions of the current package.
    current_preservation_actions:
      null as api.PackagePreservationActionsResponseBody | null,

    // The current package is being moved into a new location.
    // Set to true by this client when the package is moved.
    // Set to false by moveStatus or handlePackageLocationUpdated.
    locationChanging: false,

    // A list of packages shown during searches.
    packages: [] as Array<api.EnduroStoredPackageResponseBody>,

    // User-interface interactions between components.
    ui: {
      download: new UIRequest(),
    },
  }),
  getters: {
    isPending(): boolean {
      return (
        this.current?.status ==
        api.EnduroStoredPackageResponseBodyStatusEnum.Pending
      );
    },
    isDone(): boolean {
      return this.current?.status == api.PackageShowResponseBodyStatusEnum.Done;
    },
    isMovable(): boolean {
      return this.isDone && !this.isMoving;
    },
    isMoving(): boolean {
      return this.locationChanging;
    },
    isRejected(): boolean {
      return this.isDone && this.current?.location === undefined;
    },
  },
  actions: {
    handleEvent(event: PackageMonitorResponseBody) {
      if (event.monitor_ping_event) {
        handleMonitorPing(event.monitor_ping_event);
      } else if (event.package_created_event) {
        handlePackageCreated(event.package_created_event);
      } else if (event.package_deleted_event) {
        handlePackageDeleted(event.package_deleted_event);
      } else if (event.package_location_updated_event) {
        handlePackageLocationUpdated(event.package_location_updated_event);
      } else if (event.package_status_updated_event) {
        handlePackageStatusUpdated(event.package_status_updated_event);
      } else if (event.package_updated_event) {
        handlePackageUpdated(event.package_updated_event);
      }
    },
    async fetchCurrent(id: string) {
      this.$reset();

      const packageId = +id;
      if (Number.isNaN(packageId)) {
        throw Error("Unexpected parameter");
      }

      this.current = await client.package.packageShow({ id: packageId });

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
    async move(locationName: string) {
      if (!this.current) return;
      try {
        await client.package.packageMove({
          id: this.current.id,
          moveRequestBody: { location: locationName },
        });
      } catch (error) {
        return error;
      }
      this.$patch((state) => {
        if (!state.current) return;
        state.current.status =
          api.EnduroStoredPackageResponseBodyStatusEnum.InProgress;
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
    confirm(locationName: string) {
      if (!this.current) return;
      client.package
        .packageConfirm({
          id: this.current.id,
          confirmRequestBody: { location: locationName },
        })
        .then((payload) => {
          if (!this.current) return;
          this.current.status =
            api.EnduroStoredPackageResponseBodyStatusEnum.InProgress;
        });
    },
    reject() {
      if (!this.current) return;
      client.package.packageReject({ id: this.current.id }).then((payload) => {
        if (!this.current) return;
        this.current.status =
          api.EnduroStoredPackageResponseBodyStatusEnum.InProgress;
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

// TODO: use api.PackageMonitorResponseBody.
export interface PackageMonitorResponseBody {
  monitor_ping_event?: api.EnduroMonitorPingEventResponseBody;
  package_created_event?: api.EnduroPackageCreatedEventResponseBody;
  package_deleted_event?: api.EnduroPackageDeletedEventResponseBody;
  package_location_updated_event?: api.EnduroPackageLocationUpdatedEventResponseBody;
  package_status_updated_event?: api.EnduroPackageStatusUpdatedEventResponseBody;
  package_updated_event?: api.EnduroPackageUpdatedEventResponseBody;
}

function handleMonitorPing(event: api.EnduroMonitorPingEventResponseBody) {}

function handlePackageCreated(
  event: api.EnduroPackageCreatedEventResponseBody
) {
  const store = usePackageStore();
  store.fetchPackagesDebounced();
}

function handlePackageDeleted(
  event: api.EnduroPackageDeletedEventResponseBody
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
    state.current.location = event.location;
    state.locationChanging = false;
  });
}

class UIRequest {
  inner = ref(0);
  request = () => this.inner.value++;
}
