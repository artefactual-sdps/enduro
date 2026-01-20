import { acceptHMRUpdate, defineStore } from "pinia";

import { api, client } from "@/client";
import {
  IngestListBatchesStatusEnum,
  ResponseError,
} from "@/openapi-generator";
import router from "@/router";
import { useLayoutStore } from "@/stores/layout";

const defaultPageSize = 20;

export const useBatchStore = defineStore("batch", {
  state: () => ({
    // Batch currently displayed.
    current: null as api.EnduroIngestBatch | null,

    // SIPs of the current Batch.
    currentSips: [] as Array<api.EnduroIngestSip>,

    // A list of Batches shown during searches.
    batches: [] as Array<api.EnduroIngestBatch>,

    // Page is a subset of the total Batches list.
    page: { limit: defaultPageSize } as api.EnduroPage,

    filters: {
      identifier: "" as string | undefined,
      status: "" as IngestListBatchesStatusEnum | undefined,
      earliestCreatedTime: undefined as Date | undefined,
      latestCreatedTime: undefined as Date | undefined,
      uploaderId: undefined as string | undefined,
    },
  }),
  getters: {
    isPending(): boolean {
      return this.current?.status === api.EnduroIngestBatchStatusEnum.Pending;
    },
  },
  actions: {
    async fetchCurrent(uuid: string) {
      this.current = await client.ingest.ingestShowBatch({ uuid: uuid });

      // Update breadcrumb. TODO: should this be done in the component?
      const layoutStore = useLayoutStore();
      layoutStore.updateBreadcrumb([
        { text: "Ingest" },
        {
          route: router.resolve({ name: "/ingest/batches/" }),
          text: "Batches",
        },
        { text: this.current.identifier },
      ]);

      // TODO: add filtering and pagination for SIPs in batch view.
      await client.ingest
        .ingestListSips({ batchUuid: uuid, limit: -1 })
        .then((resp) => (this.currentSips = resp.items));
    },
    async fetchBatches(page: number) {
      return client.ingest
        .ingestListBatches({
          offset: page > 1 ? (page - 1) * this.page.limit : undefined,
          limit: this.page?.limit || undefined,
          identifier: this.filters.identifier,
          status: this.filters.status,
          earliestCreatedTime: this.filters.earliestCreatedTime,
          latestCreatedTime: this.filters.latestCreatedTime,
          uploaderUuid:
            this.filters.uploaderId !== ""
              ? this.filters.uploaderId
              : undefined,
        })
        .then((resp) => {
          this.batches = resp.items;
          this.page = resp.page;
        })
        .catch(async (err) => {
          this.batches = [];
          this.page = { limit: defaultPageSize, offset: 0, total: 0 };

          if (err instanceof ResponseError) {
            // An invalid status or time range returns a ResponseError with the
            // error message in the response body (JSON).
            return err.response.json().then((body) => {
              const modelErr = api.ModelErrorFromJSON(body);
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
    async fetchBatchesDebounced(page: number) {
      return this.fetchBatches(page);
    },
    async reviewBatch(_continue: boolean) {
      if (!this.current || !this.isPending) return;
      return client.ingest
        .ingestReviewBatch({
          uuid: this.current.uuid,
          reviewBatchRequestBody: { _continue },
        })
        .then(() => {
          // Set batch status to processing to remove
          // the review alert as soon as possible.
          if (this.current)
            this.current.status = api.EnduroIngestBatchStatusEnum.Processing;
        })
        .catch((e) => {
          console.error("Error reviewing batch", e.message);
          throw new Error("Couldn't review batch");
        });
    },
  },
  debounce: {
    fetchBatchesDebounced: [500, { isImmediate: false }],
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useBatchStore, import.meta.hot));
}
