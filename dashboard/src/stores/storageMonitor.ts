import { acceptHMRUpdate, defineStore } from "pinia";

import { getPath } from "@/client";
import { StorageMonitorConnection } from "@/monitor";

export const useStorageMonitorStore = defineStore("storageMonitor", {
  state: () => ({
    conn: new StorageMonitorConnection(getPath()),
  }),
  actions: {
    async connect() {
      if (this.conn.isConnected) return Promise.resolve();
      return this.conn.dial();
    },
  },
});

if (import.meta.hot) {
  import.meta.hot.accept(
    acceptHMRUpdate(useStorageMonitorStore, import.meta.hot),
  );
}
