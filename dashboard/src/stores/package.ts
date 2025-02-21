import { mapKeys, snakeCase } from "lodash-es";
import { acceptHMRUpdate, defineStore } from "pinia";
import { ref } from "vue";

import { api, client } from "@/client";
import {
  MonitorEventEventTypeEnum,
  IngestListSipsStatusEnum,
} from "@/openapi-generator";
import router from "@/router";
import { useLayoutStore } from "@/stores/layout";

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
    current: null as api.EnduroIngestSip | null,

    // Preservation actions of the current package.
    current_preservation_actions: null as api.SIPPreservationActions | null,

    // The current package is being moved into a new location.
    // Set to true by this client when the package is moved.
    // Set to false by moveStatus or handlePackageLocationUpdated.
    locationChanging: false,

    // A list of packages shown during searches.
    packages: [] as Array<api.EnduroIngestSip>,

    // Page is a subset of the total package list.
    page: { limit: 20 } as api.EnduroPage,

    // Pager contains a list of pages numbers to show in the pager.
    pager: { maxPages: 7 } as Pager,

    // User-interface interactions between components.
    ui: {
      download: new UIRequest(),
    },

    filters: {
      status: "" as IngestListSipsStatusEnum,
      name: "",
      earliestCreatedTime: undefined as Date | undefined,
      latestCreatedTime: undefined as Date | undefined,
    },
  }),
  getters: {
    isPending(): boolean {
      return this.current?.status == api.EnduroIngestSipStatusEnum.Pending;
    },
    isDone(): boolean {
      return this.current?.status == api.EnduroIngestSipStatusEnum.Done;
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
    getActionById: (state) => {
      return (
        actionId: number,
      ): api.EnduroIngestSipPreservationAction | undefined => {
        const x = state.current_preservation_actions?.actions?.find(
          (action: api.EnduroIngestSipPreservationAction) =>
            action.id === actionId,
        );
        return x;
      };
    },
    getTaskById: (state) => {
      return (
        actionId: number,
        taskId: number,
      ): api.EnduroIngestSipPreservationTask | undefined => {
        const action = state.current_preservation_actions?.actions?.find(
          (action: api.EnduroIngestSipPreservationAction) =>
            action.id === actionId,
        );
        if (!action) return;
        return action.tasks?.find(
          (task: api.EnduroIngestSipPreservationTask) => task.id === taskId,
        );
      };
    },
  },
  actions: {
    handleEvent(event: api.MonitorEventEvent) {
      const json = JSON.parse(event.value);
      // TODO: avoid key transformation in the backend or make
      // this fully recursive, considering objects and slices.
      const value = mapKeys(json, (_, key) => snakeCase(key));
      if (value.item) {
        value.item = mapKeys(value.item, (_, key) => snakeCase(key));
      }
      handlers[event.type](value);
    },
    async fetchCurrent(id: string) {
      const packageId = +id;
      if (Number.isNaN(packageId)) {
        throw Error("Unexpected parameter");
      }

      this.current = await client.ingest.ingestShowSip({ id: packageId });

      // Update breadcrumb. TODO: should this be done in the component?
      const layoutStore = useLayoutStore();

      layoutStore.updateBreadcrumb([
        { route: router.resolve("/packages/"), text: "Packages" },
        { text: this.current.name },
      ]);

      await Promise.allSettled([
        client.ingest
          .ingestListSipPreservationActions({ id: packageId })
          .then((resp) => {
            this.current_preservation_actions = resp;
          }),
        client.ingest.ingestMoveSipStatus({ id: packageId }).then((resp) => {
          this.locationChanging = !resp.done;
        }),
      ]);
    },
    async fetchPackages(page: number) {
      const resp = await client.ingest.ingestListSips({
        offset: page > 1 ? (page - 1) * this.page.limit : undefined,
        limit: this.page?.limit || undefined,
        status: this.filters.status ?? undefined,
        name: this.filters.name ?? undefined,
        earliestCreatedTime: this.filters.earliestCreatedTime,
        latestCreatedTime: this.filters.latestCreatedTime,
      });
      this.packages = resp.items;
      this.page = resp.page;
      this.updatePager();
    },
    async fetchPackagesDebounced(page: number) {
      return this.fetchPackages(page);
    },
    async move(locationId: string) {
      if (!this.current) return;
      try {
        await client.ingest.ingestMoveSip({
          id: this.current.id,
          confirmSipRequestBody: { locationId: locationId },
        });
      } catch (error) {
        return error;
      }
      this.$patch((state) => {
        if (!state.current) return;
        state.current.status = api.EnduroIngestSipStatusEnum.InProgress;
        state.locationChanging = true;
      });
    },
    async moveStatus() {
      if (!this.current) return;
      let resp;
      try {
        resp = await client.ingest.ingestMoveSipStatus({
          id: this.current?.id,
        });
      } catch (error) {
        return error;
      }
      this.locationChanging = !resp.done;
    },
    confirm(locationId: string) {
      if (!this.current) return;
      client.ingest
        .ingestConfirmSip({
          id: this.current.id,
          confirmSipRequestBody: { locationId: locationId },
        })
        .then(() => {
          if (!this.current) return;
          this.current.status = api.EnduroIngestSipStatusEnum.InProgress;
        });
    },
    reject() {
      if (!this.current) return;
      client.ingest.ingestRejectSip({ id: this.current.id }).then(() => {
        if (!this.current) return;
        this.current.status = api.EnduroIngestSipStatusEnum.InProgress;
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
    updatePager(): void {
      const pgr = this.pager;
      pgr.total = Math.ceil(this.page.total / this.page.limit);
      pgr.current = Math.floor(this.page.offset / this.page.limit) + 1;

      let first = 1;
      const count = pgr.total < pgr.maxPages ? pgr.total : pgr.maxPages;
      const half = Math.floor(pgr.maxPages / 2);
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
      for (let i = 0; i < count; i++) {
        pgr.pages[i] = i + first;
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
  [key in api.MonitorEventEventTypeEnum]: (data: unknown) => void;
} = {
  [MonitorEventEventTypeEnum.MonitorPingEvent]: handleMonitorPing,
  [MonitorEventEventTypeEnum.SipCreatedEvent]: handlePackageCreated,
  [MonitorEventEventTypeEnum.SipUpdatedEvent]: handlePackageUpdated,
  [MonitorEventEventTypeEnum.SipStatusUpdatedEvent]: handlePackageStatusUpdated,
  [MonitorEventEventTypeEnum.SipLocationUpdatedEvent]:
    handlePackageLocationUpdated,
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

function handlePackageCreated(data: unknown) {
  api.SIPCreatedEventFromJSON(data);
  const store = usePackageStore();
  store.fetchPackagesDebounced(1);
}

function handlePackageUpdated(data: unknown) {
  const event = api.SIPUpdatedEventFromJSON(data);
  const store = usePackageStore();
  store.fetchPackagesDebounced(1);
  if (store.$state.current?.id != event.id) return;
  Object.assign(store.$state.current, event.item);
}

function handlePackageStatusUpdated(data: unknown) {
  const event = api.SIPStatusUpdatedEventFromJSON(data);
  const store = usePackageStore();
  store.fetchPackagesDebounced(1);
  if (store.$state.current?.id != event.id) return;
  store.$state.current.status = event.status;
}

function handlePackageLocationUpdated(data: unknown) {
  const event = api.SIPLocationUpdatedEventFromJSON(data);
  const store = usePackageStore();
  store.fetchPackagesDebounced(1);
  store.$patch((state) => {
    if (state.current?.id != event.id) return;
    state.current.locationId = event.locationId;
    state.locationChanging = false;
  });
}

function handlePreservationActionCreated(data: unknown) {
  const event = api.SIPPreservationActionCreatedEventFromJSON(data);
  const store = usePackageStore();

  // Ignore event if it does not relate to the current package.
  if (store.current?.id != event.item.sipId) return;

  // Append the action.
  store.current_preservation_actions?.actions?.unshift(event.item);
}

function handlePreservationActionUpdated(data: unknown) {
  const event = api.SIPPreservationActionUpdatedEventFromJSON(data);
  const store = usePackageStore();

  // Ignore event if it does not relate to the current package.
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

function handlePreservationTaskUpdated(data: unknown) {
  const event = api.SIPPreservationTaskUpdatedEventFromJSON(data);
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
