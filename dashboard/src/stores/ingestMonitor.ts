import { acceptHMRUpdate, defineStore } from "pinia";

import { getPath } from "@/client";
import { IngestMonitorConnection } from "@/monitor";

export const useIngestMonitorStore = defineStore("ingestMonitor", {
  state: () => ({
    conn: new IngestMonitorConnection(getPath()),
  }),
  actions: {
    async connect() {
      if (this.conn.isConnected) return;
      return this.conn.dial();
    },
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(
    acceptHMRUpdate(useIngestMonitorStore, import.meta.hot),
  );
}
