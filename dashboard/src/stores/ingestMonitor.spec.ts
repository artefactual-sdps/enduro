import { createPinia, setActivePinia } from "pinia";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import WS from "vitest-websocket-mock";

import { useIngestMonitorStore } from "@/stores/ingestMonitor";

vi.mock("@/client", async () => {
  return {
    client: {
      ingest: {
        ingestMonitorRequest: vi.fn(() => Promise.resolve()),
      },
    },
    getPath: () => "http://localhost:1234",
  };
});

describe("useIngestMonitorStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.resetAllMocks();
  });

  it("connects the monitor", async () => {
    const server = new WS("ws://localhost:1234/ingest/monitor");
    const store = useIngestMonitorStore();

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
