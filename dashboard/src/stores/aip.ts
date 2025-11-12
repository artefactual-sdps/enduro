import { acceptHMRUpdate, defineStore } from "pinia";

import { api, client, getPath } from "@/client";
import { logError } from "@/helpers/logs";
import { ResponseError, StorageListAipsStatusEnum } from "@/openapi-generator";
import router from "@/router";
import { useLayoutStore } from "@/stores/layout";

const defaultPageSize = 20;

export const useAipStore = defineStore("aip", {
  state: () => ({
    // AIP currently displayed.
    current: null as api.EnduroStorageAip | null,

    // Workflows of the current AIP.
    currentWorkflows: null as api.EnduroStorageAipWorkflows | null,

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
    downloadError: null as string | null,
  }),
  getters: {
    isDeleted(): boolean {
      return this.current?.status == api.EnduroStorageAipStatusEnum.Deleted;
    },
    isMovable(): boolean {
      return this.isStored && !this.isProcessing;
    },
    isMoving(): boolean {
      return this.locationChanging;
    },
    isPending(): boolean {
      return this.current?.status == api.EnduroStorageAipStatusEnum.Pending;
    },
    isStored(): boolean {
      return this.current?.status == api.EnduroStorageAipStatusEnum.Stored;
    },
    isProcessing(): boolean {
      return this.current?.status == api.EnduroStorageAipStatusEnum.Processing;
    },
    getWorkflowById: (state) => {
      return (workflowId: string): api.EnduroStorageAipWorkflow | undefined => {
        const x = state.currentWorkflows?.workflows?.find(
          (workflow: api.EnduroStorageAipWorkflow) =>
            workflow.uuid === workflowId,
        );
        return x;
      };
    },
    getTaskById: (state) => {
      return (
        workflowId: string,
        taskId: string,
      ): api.EnduroStorageAipTask | undefined => {
        const workflow = state.currentWorkflows?.workflows?.find(
          (workflow: api.EnduroStorageAipWorkflow) =>
            workflow.uuid === workflowId,
        );
        if (!workflow) return;
        return workflow.tasks?.find(
          (task: api.EnduroStorageAipTask) => task.uuid === taskId,
        );
      };
    },
  },
  actions: {
    async fetchCurrent(id: string) {
      const layoutStore = useLayoutStore();
      let breadcrumb = "";
      return client.storage
        .storageShowAip({ uuid: id })
        .then((resp) => {
          this.current = resp;
          breadcrumb = resp.name;
        })
        .catch((e) => {
          this.current = null;
          this.currentWorkflows = null;
          breadcrumb = "Error";

          logError(e, "Error fetching AIP");

          if (e instanceof ResponseError && e.response.status === 404) {
            // The AIPDeletionReviewAlert component has special handling for
            // 404 errors, so rethrow the error.
            throw e;
          }

          throw new Error("Couldn't load AIP");
        })
        .then(() => {
          // Fetch workflows for the current AIP.
          if (!this.current) return;
          return this.fetchWorkflows(this.current.uuid);
        })
        .finally(() => {
          // Update breadcrumb. TODO: should this be done in the component?
          layoutStore.updateBreadcrumb([
            { text: "Storage" },
            { route: router.resolve("/storage/aips/"), text: "AIPs" },
            { text: breadcrumb },
          ]);
        });
    },
    async fetchWorkflows(id: string) {
      return client.storage
        .storageListAipWorkflows({
          uuid: id,
        })
        .then((resp) => {
          if (resp && resp.workflows) {
            resp.workflows.reverse();
          }
          this.currentWorkflows = resp;
        })
        .catch((e) => {
          this.currentWorkflows = null;

          logError(e, "Error fetching workflows");

          // Don't show an error if we don't have permission to view workflows.
          if (e instanceof ResponseError && e.response.status === 403) {
            return;
          }

          throw new Error("Couldn't load workflows");
        });
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

          if (err instanceof RangeError) {
            // An invalid date parameter (e.g. earliestCreatedTime) returns a
            // RangeError with a message like "invalid date".
            console.error("Error fetching AIPs", "Range error: " + err.message);
            throw new Error(err.message);
          } else {
            logError(err, "Error fetching AIPs");
          }

          throw new Error("Couldn't load AIPs");
        });
    },
    async move(locationId: string) {
      if (!this.current) return;
      try {
        await client.storage.storageMoveAip({
          uuid: this.current.uuid,
          confirmSipRequestBody: { locationUuid: locationId },
        });
      } catch (error) {
        return error;
      }
      this.$patch((state) => {
        if (!state.current) return;
        state.current.status = api.EnduroStorageAipStatusEnum.Processing;
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
      return client.storage
        .storageRequestAipDeletion({
          uuid: this.current.uuid,
          requestAipDeletionRequestBody: { reason: reason },
        })
        .catch((e) => {
          console.error("Error requesting deletion", e.message);
          throw new Error("Couldn't create deletion request");
        });
    },
    async reviewDeletion(approved: boolean) {
      if (!this.current) return;
      return client.storage
        .storageReviewAipDeletion({
          uuid: this.current.uuid,
          reviewAipDeletionRequestBody: { approved: approved },
        })
        .catch((e) => {
          console.error("Error reviewing deletion", e.message);
          throw new Error("Couldn't update deletion request");
        });
    },
    async cancelDeletionRequest(): Promise<void> {
      if (!this.current) return;
      return client.storage
        .storageCancelAipDeletion({
          uuid: this.current.uuid,
          cancelAipDeletionRequestBody: {},
        })
        .catch((e) => {
          logError(e, "Error cancelling deletion request");
          throw new Error("Couldn't cancel deletion request");
        });
    },
    async canCancelDeletion(): Promise<boolean> {
      if (!this.current) return false;
      return client.storage
        .storageCancelAipDeletion({
          uuid: this.current.uuid,
          cancelAipDeletionRequestBody: {
            check: true,
          },
        })
        .then(() => {
          return true;
        })
        .catch((e) => {
          // A 403 Forbidden response means this user is not authorized to
          // cancel the deletion request.
          if (e instanceof ResponseError && e.response.status === 403) {
            return false;
          }

          logError(e, "Error checking user authorization to cancel deletion");

          return false;
        });
    },
    async download() {
      if (!this.current) return;
      try {
        await client.storage.storageDownloadAipRequest({
          uuid: this.current.uuid,
        });
        window.open(
          getPath() + "/storage/aips/" + this.current.uuid + "/download",
          "_blank",
        );
      } catch (err) {
        // Try to parse the error and save it for 5 seconds. It will
        // replace the download button with an alert including the
        // error message in the AipLocationCard component.
        let errorMsg = "Unexpected error downloading AIP";
        if (err instanceof ResponseError) {
          const body = await err.response.json();
          if (body.message) {
            errorMsg = body.message;
          }
        }
        this.downloadError = errorMsg;
        setTimeout(() => (this.downloadError = null), 5000);
      }
    },
    async fetchAipsDebounced(page: number) {
      return this.fetchAips(page);
    },
  },
  debounce: {
    fetchAipsDebounced: [500, { isImmediate: false }],
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useAipStore, import.meta.hot));
}
