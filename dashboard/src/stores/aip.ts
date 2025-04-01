import { acceptHMRUpdate, defineStore } from "pinia";
import { ref } from "vue";

import { api, client } from "@/client";
import { ResponseError, StorageListAipsStatusEnum } from "@/openapi-generator";
import router from "@/router";
import { useLayoutStore } from "@/stores/layout";

class UIRequest {
  inner = ref(0);
  request = () => this.inner.value++;
}

const defaultPageSize = 20;

export const useAipStore = defineStore("aip", {
  state: () => ({
    // AIP currently displayed.
    current: null as api.EnduroStorageAip | null,

    // Workflows of the current AIP.
    currentWorkflows: null as api.AIPWorkflows | null,

    // The current AIP is being moved into a new location.
    // Set to true by this client when the AIP is moved.
    // Set to false by moveStatus or handleSipLocationUpdated.
    locationChanging: false,

    // A list of AIPs shown during searches.
    aips: [] as Array<api.EnduroStorageAip>,

    // Page is a subset of the total AIP list.
    page: { limit: defaultPageSize } as api.EnduroPage,

    filters: {
      name: "" as string | undefined,
      status: "" as StorageListAipsStatusEnum | undefined,
      earliestCreatedTime: undefined as Date | undefined,
      latestCreatedTime: undefined as Date | undefined,
    },

    // User-interface interactions between components.
    ui: {
      download: new UIRequest(),
    },
  }),
  getters: {
    isDeleted(): boolean {
      return this.current?.status == api.EnduroStorageAipStatusEnum.Deleted;
    },
    isMovable(): boolean {
      return this.isStored && !this.isMoving;
    },
    isMoving(): boolean {
      return this.locationChanging;
    },
    isPending(): boolean {
      return this.current?.status == api.EnduroStorageAipStatusEnum.Pending;
    },
    isRejected(): boolean {
      return this.current?.status == api.EnduroStorageAipStatusEnum.Rejected;
    },
    isStored(): boolean {
      return this.current?.status == api.EnduroStorageAipStatusEnum.Stored;
    },
  },
  actions: {
    async fetchCurrent(id: string) {
      this.current = await client.storage.storageShowAip({ uuid: id });
      this.currentWorkflows = await client.storage
        .storageListAipWorkflows({
          uuid: id,
        })
        .then((resp) => {
          resp.workflows?.reverse();
          return resp;
        });

      this.locationChanging =
        this.current?.status == api.EnduroStorageAipStatusEnum.Moving;

      // Update breadcrumb. TODO: should this be done in the component?
      const layoutStore = useLayoutStore();
      layoutStore.updateBreadcrumb([
        { text: "Storage" },
        { route: router.resolve("/storage/aips/"), text: "AIPs" },
        { text: this.current.name },
      ]);
    },
    async fetchAips(page: number) {
      return client.storage
        .storageListAips({
          offset: page > 1 ? (page - 1) * this.page.limit : undefined,
          limit: this.page?.limit || undefined,
          name: this.filters.name,
          status: this.filters.status,
          earliestCreatedTime: this.filters.earliestCreatedTime,
          latestCreatedTime: this.filters.latestCreatedTime,
        })
        .then((resp) => {
          this.aips = resp.items;
          this.page = resp.page;
        })
        .catch(async (err) => {
          this.aips = [];
          this.page = { limit: defaultPageSize, offset: 0, total: 0 };

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
    async move(locationId: string) {
      if (!this.current) return;
      try {
        await client.storage.storageMoveAip({
          uuid: this.current.uuid,
          confirmSipRequestBody: { locationId: locationId },
        });
      } catch (error) {
        return error;
      }
      this.$patch((state) => {
        if (!state.current) return;
        state.current.status = api.EnduroStorageAipStatusEnum.Moving;
        state.locationChanging = true;
      });
    },
    async moveStatus() {
      if (!this.current) return;
      let resp;
      try {
        resp = await client.storage.storageMoveAipStatus({
          uuid: this.current?.uuid,
        });
      } catch (error) {
        return error;
      }
      this.locationChanging = !resp.done;
    },
    async requestDeletion(reason: string) {
      if (!this.current) return;
      try {
        await client.storage.storageRequestAipDeletion({
          uuid: this.current.uuid,
          requestAipDeletionRequestBody: { reason: reason },
        });
      } catch (error) {
        return error;
      }

      // TODO: Remove this once we have websocket in the storage service.
      while (true) {
        if (!this.current) return;
        this.fetchCurrent(this.current.uuid);
        if (this.current.status == api.EnduroStorageAipStatusEnum.Pending) {
          break;
        }
        await new Promise((resolve) => setTimeout(resolve, 1000));
      }
    },
    async reviewDeletion(approved: boolean) {
      if (!this.current) return;
      try {
        await client.storage.storageReviewAipDeletion({
          uuid: this.current.uuid,
          reviewAipDeletionRequestBody: { approved: approved },
        });
      } catch (error) {
        return error;
      }

      // TODO: Remove this once we have websocket in the storage service.
      while (true) {
        if (!this.current) return;
        this.fetchCurrent(this.current.uuid);
        if (
          this.current.status == api.EnduroStorageAipStatusEnum.Deleted ||
          this.current.status == api.EnduroStorageAipStatusEnum.Stored
        ) {
          break;
        }
        await new Promise((resolve) => setTimeout(resolve, 1000));
      }
    },
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useAipStore, import.meta.hot));
}
