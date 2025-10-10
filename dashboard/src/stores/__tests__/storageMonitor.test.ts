import { createPinia, setActivePinia } from "pinia";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import WS from "vitest-websocket-mock";

import { useStorageMonitorStore } from "@/stores/storageMonitor";

vi.mock("@/client", async () => {
  return {
    client: {
      storage: {
        storageMonitorRequest: vi.fn(() => Promise.resolve()),
      },
    },
    getPath: () => "http://localhost:1234",
  };
});

describe("useStorageMonitorStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  it("connects the monitor", async () => {
    const server = new WS("ws://localhost:1234/storage/monitor");
    const store = useStorageMonitorStore();

    store.connect();
    await vi.runAllTimersAsync();

    expect(store.conn.isConnected).toBe(true);

    store.connect(); // second call, should be no-op
    await vi.runAllTimersAsync();

    expect(store.conn.isConnected).toBe(true);

    store.conn.close();
    server.close();
  });
});
