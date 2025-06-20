import { acceptHMRUpdate, defineStore } from "pinia";

import { api, client } from "@/client";
import { IngestListSipsStatusEnum, ResponseError } from "@/openapi-generator";
import router from "@/router";
import { useLayoutStore } from "@/stores/layout";

const defaultPageSize = 20;

function logError(e: Error, msg: string) {
  if (e instanceof ResponseError) {
    msg = msg + ":";
    console.error(msg, e.response.status, e.response.statusText);
  } else {
    // Unknown error type.
    console.error(msg, e.message);
  }
}

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
    },
  }),
  getters: {
    isPending(): boolean {
      return this.current?.status == api.EnduroIngestSipStatusEnum.Pending;
    },
    getWorkflowById: (state) => {
      return (workflowId: number): api.EnduroIngestSipWorkflow | undefined => {
        const x = state.currentWorkflows?.workflows?.find(
          (workflow: api.EnduroIngestSipWorkflow) => workflow.id === workflowId,
        );
        return x;
      };
    },
    getTaskById: (state) => {
      return (
        workflowId: number,
        taskId: number,
      ): api.EnduroIngestSipTask | undefined => {
        const workflow = state.currentWorkflows?.workflows?.find(
          (workflow: api.EnduroIngestSipWorkflow) => workflow.id === workflowId,
        );
        if (!workflow) return;
        return workflow.tasks?.find(
          (task: api.EnduroIngestSipTask) => task.id === taskId,
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
    confirm(locationId: string) {
      if (!this.current) return;
      client.ingest
        .ingestConfirmSip({
          uuid: this.current.uuid,
          confirmSipRequestBody: { locationId: locationId },
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
  },
  debounce: {
    fetchSipsDebounced: [500, { isImmediate: false }],
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(acceptHMRUpdate(useSipStore, import.meta.hot));
}
