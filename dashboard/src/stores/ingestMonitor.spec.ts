import { createPinia, setActivePinia } from "pinia";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { useIngestMonitorStore } from "@/stores/ingestMonitor";
import { FakeEventSource } from "@/test/fake-event-source";

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
    vi.stubGlobal("EventSource", FakeEventSource);
  });

  afterEach(() => {
    vi.resetAllMocks();
    vi.unstubAllGlobals();
    FakeEventSource.reset();
  });

  it("connects the monitor", async () => {
    const store = useIngestMonitorStore();

    store.connect();
    await vi.runAllTimersAsync();
    const source = FakeEventSource.latest();
    source.open();

    expect(store.conn.isConnected).toBe(true);
    expect(FakeEventSource.instances).toHaveLength(1);

    store.connect(); // second call, should be no-op
    await vi.runAllTimersAsync();

    expect(store.conn.isConnected).toBe(true);
    expect(FakeEventSource.instances).toHaveLength(1);
    expect(FakeEventSource.latest()).toBe(source);

    store.conn.close();
  });
});
