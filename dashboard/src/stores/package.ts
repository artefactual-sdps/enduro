import { api, client } from "@/client";
import { MonitorEventEventTypeEnum } from "@/openapi-generator";
import { useLayoutStore } from "@/stores/layout";
import router from "@/router";
import { defineStore, acceptHMRUpdate } from "pinia";
import { ref } from "vue";
import { mapKeys, snakeCase } from "lodash-es";

export interface Pager {
  // maxPages is the maximum number of page links to show in the pager.
  readonly maxPages: number;

  current: number;
  first: number;
  last: number;
  total: number;
  pages: Array<number>;
}

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

    // Page is a subset of the total package list.
    page: { limit: 20 } as api.EnduroPage,

    // Pager contains a list of pages numbers to show in the pager.
    pager: { maxPages: 7 } as Pager,

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
      return this.page.offset + this.page.limit < this.page.total;
    },
    hasPrevPage(): boolean {
      return this.page.offset > 0;
    },
    lastResultOnPage(): number {
      let i = this.page.offset + this.page.limit;
      if (i > this.page.total) {
        i = this.page.total;
      }
      return i;
    },
    updatePager(): void {
      let pgr = this.pager;
      pgr.total = Math.ceil(this.page.total / this.page.limit);
      pgr.current = Math.floor(this.page.offset / this.page.limit) + 1;

      let first = 1;
      let count = pgr.total < pgr.maxPages ? pgr.total : pgr.maxPages;
      let half = Math.floor(pgr.maxPages / 2);
      if (pgr.current > half + 1) {
        if (pgr.total - pgr.current < half) {
          first = pgr.total - count + 1;
        } else {
          first = pgr.current - half;
        }
      }
      pgr.first = first;
      pgr.last = first + count - 1;

      pgr.pages = new Array(count);
      for (var i = 0; i < count; i++) {
        pgr.pages[i] = i + first;
      }
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
      // TODO: avoid key transformation in the backend or make
      // this fully recursive, considering objects and slices.
      let value = mapKeys(json, (_, key) => snakeCase(key));
      if (value.item) {
        value.item = mapKeys(value.item, (_, key) => snakeCase(key));
      }
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
    async fetchPackages(page: number) {
      const resp = await client.package.packageList({
        offset: page > 1 ? (page - 1) * this.page.limit : undefined,
        limit: this.page?.limit || undefined,
      });
      this.packages = resp.items;
      this.page = resp.page;
      this.updatePager;
    },
    async fetchPackagesDebounced(page: number) {
      return this.fetchPackages(page);
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
      if (this.hasNextPage) {
        this.fetchPackages(this.pager.current + 1);
      }
    },
    prevPage() {
      if (this.hasPrevPage) {
        this.fetchPackages(this.pager.current - 1);
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
  store.fetchPackagesDebounced(1);
}

function handlePackageUpdated(data: any) {
  const event = api.PackageUpdatedEventFromJSON(data);
  const store = usePackageStore();
  store.fetchPackagesDebounced(1);
  if (store.$state.current?.id != event.id) return;
  Object.assign(store.$state.current, event.item);
}

function handlePackageStatusUpdated(data: any) {
  const event = api.PackageStatusUpdatedEventFromJSON(data);
  const store = usePackageStore();
  store.fetchPackagesDebounced(1);
  if (store.$state.current?.id != event.id) return;
  store.$state.current.status = event.status;
}

function handlePackageLocationUpdated(data: any) {
  const event = api.PackageLocationUpdatedEventFromJSON(data);
  const store = usePackageStore();
  store.fetchPackagesDebounced(1);
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

  // Keep existing tasks, this event doesn't include them.
  const tasks = action.tasks;
  Object.assign(action, event.item);
  action.tasks = tasks;
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
