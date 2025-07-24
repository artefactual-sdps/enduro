import { acceptHMRUpdate, defineStore } from "pinia";

import { api, client, getPath } from "@/client";
import { logError } from "@/helpers/logs";
import { IngestListSipsStatusEnum, ResponseError } from "@/openapi-generator";
import router from "@/router";
import { useLayoutStore } from "@/stores/layout";

const defaultPageSize = 20;

export const useSipStore = defineStore("sip", {
  state: () => ({
    // SIP currently displayed.
    current: null as api.EnduroIngestSip | null,

    // Workflows of the current SIP.
    currentWorkflows: null as api.SIPWorkflows | null,

    // A list of SIPs shown during searches.
    sips: [] as Array<api.EnduroIngestSip>,

    // Page is a subset of the total SIP list.
    page: { limit: defaultPageSize } as api.EnduroPage,

    filters: {
      name: "" as string | undefined,
      status: "" as IngestListSipsStatusEnum | undefined,
      earliestCreatedTime: undefined as Date | undefined,
      latestCreatedTime: undefined as Date | undefined,
      uploaderId: "" as string,
    },
    downloadError: null as string | null,
  }),
  getters: {
    isPending(): boolean {
      return this.current?.status == api.EnduroIngestSipStatusEnum.Pending;
    },
    getWorkflowById: (state) => {
      return (workflowId: string): api.EnduroIngestSipWorkflow | undefined => {
        const x = state.currentWorkflows?.workflows?.find(
          (workflow: api.EnduroIngestSipWorkflow) =>
            workflow.uuid === workflowId,
        );
        return x;
      };
    },
    getTaskById: (state) => {
      return (
        workflowId: string,
        taskId: string,
      ): api.EnduroIngestSipTask | undefined => {
        const workflow = state.currentWorkflows?.workflows?.find(
          (workflow: api.EnduroIngestSipWorkflow) =>
            workflow.uuid === workflowId,
        );
        if (!workflow) return;
        return workflow.tasks?.find(
          (task: api.EnduroIngestSipTask) => task.uuid === taskId,
        );
      };
    },
  },
  actions: {
    async fetchCurrent(id: string) {
      const layoutStore = useLayoutStore();
      let breadcrumb = "";
      return client.ingest
        .ingestShowSip({ uuid: id })
        .then((sip) => {
          this.current = sip;
          breadcrumb = this.current.name || "Unnamed";
        })
        .catch((e) => {
          this.current = null;
          breadcrumb = "Error";

          logError(e, "Error fetching SIP");
          throw new Error("Couldn't load SIP");
        })
        .finally(() => {
          // Update breadcrumb. TODO: should this be done in the component?
          layoutStore.updateBreadcrumb([
            { text: "Ingest" },
            { route: router.resolve("/ingest/sips/"), text: "SIPs" },
            { text: breadcrumb },
          ]);
        });
    },
    async fetchCurrentWorkflows(id: string) {
      return client.ingest
        .ingestListSipWorkflows({ uuid: id })
        .then((workflows) => {
          this.currentWorkflows = workflows;
        })
        .catch((e) => {
          this.currentWorkflows = null;
          logError(e, "Error fetching workflows");

          // Don't show an error if we get a 403 Forbidden response.
          if (e.response.status === 403) {
            return;
          }

          throw new Error("Couldn't load workflows");
        });
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
          uploaderUuid:
            this.filters.uploaderId !== ""
              ? this.filters.uploaderId
              : undefined,
        })
        .then((resp) => {
          this.sips = resp.items;
          this.page = resp.page;
        })
        .catch(async (err) => {
          this.sips = [];
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
    async fetchSipsDebounced(page: number) {
      return this.fetchSips(page);
    },
    confirm(locationId: string) {
      if (!this.current) return;
      client.ingest
        .ingestConfirmSip({
          uuid: this.current.uuid,
          confirmSipRequestBody: { locationUuid: locationId },
        })
        .then(() => {
          if (!this.current) return;
          this.current.status = api.EnduroIngestSipStatusEnum.Processing;
        });
    },
    reject() {
      if (!this.current) return;
      client.ingest.ingestRejectSip({ uuid: this.current.uuid }).then(() => {
        if (!this.current) return;
        this.current.status = api.EnduroIngestSipStatusEnum.Processing;
      });
    },
    async download() {
      if (!this.current) return;
      try {
        await client.ingest.ingestDownloadSipRequest({
          uuid: this.current.uuid,
        });
        window.open(
          getPath() + "/ingest/sips/" + this.current.uuid + "/download",
          "_blank",
        );
      } catch (err) {
        // Try to parse the error and save it for 5 seconds. It will
        // replace the download button with an alert including the
        // error message in the SipRelatedPackages component.
        let errorMsg = "Unexpected error downloading package";
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
  },
  debounce: {
    fetchSipsDebounced: [500, { isImmediate: false }],
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useSipStore, import.meta.hot));
}
