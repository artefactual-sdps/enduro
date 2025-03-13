import { acceptHMRUpdate, defineStore } from "pinia";
import { ref } from "vue";

import { api, client } from "@/client";
import { IngestListSipsStatusEnum, ResponseError } from "@/openapi-generator";
import router from "@/router";
import { useLayoutStore } from "@/stores/layout";

const defaultPageSize = 20;

class UIRequest {
  inner = ref(0);
  request = () => this.inner.value++;
}

export interface Pager {
  // maxPages is the maximum number of page links to show in the pager.
  readonly maxPages: number;

  current: number;
  first: number;
  last: number;
  total: number;
  pages: Array<number>;
}

export const useSipStore = defineStore("sip", {
  state: () => ({
    // SIP currently displayed.
    current: null as api.EnduroIngestSip | null,

    // Preservation actions of the current SIP.
    currentPreservationActions: null as api.SIPPreservationActions | null,

    // A list of SIPs shown during searches.
    sips: [] as Array<api.EnduroIngestSip>,

    // The current SIP is being moved into a new location.
    // Set to true by this client when the SIP is moved.
    // Set to false by moveStatus or handleSipLocationUpdated.
    locationChanging: false,

    // Page is a subset of the total SIP list.
    page: { limit: defaultPageSize } as api.EnduroPage,

    // Pager contains a list of page numbers to show in the pager.
    pager: { maxPages: 7 } as Pager,

    // User-interface interactions between components.
    ui: {
      download: new UIRequest(),
    },

    filters: {
      name: "" as string | undefined,
      status: "" as IngestListSipsStatusEnum | undefined,
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
        const x = state.currentPreservationActions?.actions?.find(
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
        const action = state.currentPreservationActions?.actions?.find(
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
    async fetchCurrent(id: string) {
      const sipId = +id;
      if (Number.isNaN(sipId)) {
        throw Error("Unexpected parameter");
      }

      this.current = await client.ingest.ingestShowSip({ id: sipId });

      // Update breadcrumb. TODO: should this be done in the component?
      const layoutStore = useLayoutStore();

      layoutStore.updateBreadcrumb([
        { text: "Ingest" },
        { route: router.resolve("/ingest/sips/"), text: "SIPs" },
        { text: this.current.name },
      ]);

      await Promise.allSettled([
        client.ingest
          .ingestListSipPreservationActions({ id: sipId })
          .then((resp) => {
            this.currentPreservationActions = resp;
          }),
        client.ingest.ingestMoveSipStatus({ id: sipId }).then((resp) => {
          this.locationChanging = !resp.done;
        }),
      ]);
    },
    async fetchSips(page: number) {
      return client.ingest
        .ingestListSips({
          offset: page > 1 ? (page - 1) * this.page.limit : undefined,
          limit: this.page?.limit || undefined,
          name: this.filters.name,
          status: this.filters.status,
          earliestCreatedTime: this.filters.earliestCreatedTime,
          latestCreatedTime: this.filters.latestCreatedTime,
        })
        .then((resp) => {
          this.sips = resp.items;
          this.page = resp.page;
          this.updatePager();
        })
        .catch(async (err) => {
          this.sips = [];
          this.page = { limit: defaultPageSize, offset: 0, total: 0 };
          this.updatePager();

          if (err instanceof ResponseError) {
            // An invalid status or time range returns a ResponseError with the
            // error message in the response body (JSON).
            return err.response.text().then((body) => {
              const modelErr = api.ModelErrorFromJSON(JSON.parse(body));
              console.error(
                "API response",
                err.response.status,
                modelErr.message,
              );
              throw new Error(modelErr.message);
            });
          } else if (err instanceof RangeError) {
            // An invalid date parameter (e.g. earliestCreatedTime) returns a
            // RangeError with a message like "invalid date".
            console.error("Range error", err.message);
            throw new Error(err.message);
          } else {
            console.error("Unknown error", err.message);
            throw new Error(err.message);
          }
        });
    },
    async fetchSipsDebounced(page: number) {
      return this.fetchSips(page);
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
        this.fetchSips(this.pager.current + 1);
      }
    },
    prevPage() {
      if (this.hasPrevPage) {
        this.fetchSips(this.pager.current - 1);
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
    fetchSipsDebounced: [500, { isImmediate: false }],
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useSipStore, import.meta.hot));
}
