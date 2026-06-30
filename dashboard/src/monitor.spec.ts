import { createPinia, setActivePinia } from "pinia";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { handleUnauthorized } from "@/client";
import {
  IngestMonitorConnection,
  type RetryOptions,
  StorageMonitorConnection,
} from "@/monitor";
import { handleIngestEvent } from "@/monitor-ingest";
import { handleStorageEvent } from "@/monitor-storage";
import {
  type EnduroIngestBatch,
  type EnduroIngestSip,
  type EnduroIngestSipTask,
  type EnduroIngestSipWorkflows,
  type EnduroStorageAip,
  type EnduroStorageAipTask,
  type EnduroStorageAipWorkflows,
  IngestEventValueTypeEnum,
  StorageEventValueTypeEnum,
} from "@/openapi-generator";
import { useAipStore } from "@/stores/aip";
import { useAuthStore } from "@/stores/auth";
import { useBatchStore } from "@/stores/batch";
import { useSipStore } from "@/stores/sip";
import { FakeEventSource } from "@/test/fake-event-source";

vi.mock("@/client", async () => {
  const api = await vi.importActual("@/openapi-generator");
  return {
    api: { ...api },
    handleUnauthorized: vi.fn(() => Promise.resolve()),
  };
});

vi.mock("eventsource", async () => {
  const fakeEventSourceModule = await vi.importActual<
    typeof import("@/test/fake-event-source")
  >("@/test/fake-event-source");

  return { EventSource: fakeEventSourceModule.FakeEventSource };
});

vi.mock("@/monitor-ingest", () => ({
  handleIngestEvent: vi.fn(),
}));

vi.mock("@/monitor-storage", () => ({
  handleStorageEvent: vi.fn(),
}));

type MonitorConnectionFactory = (
  retry: RetryOptions,
) => IngestMonitorConnection | StorageMonitorConnection;

const monitorConnectionCases: Array<[string, MonitorConnectionFactory]> = [
  [
    "ingest",
    (retry) => new IngestMonitorConnection("http://localhost:1234", retry),
  ],
  [
    "storage",
    (retry) => new StorageMonitorConnection("http://localhost:1234", retry),
  ],
];

// Mock timers.
beforeEach(() => {
  setActivePinia(createPinia());
  vi.useFakeTimers();
});
afterEach(() => {
  vi.resetAllMocks();
  vi.unstubAllGlobals();
  FakeEventSource.reset();
});

describe("MonitorConnection", () => {
  describe("url", () => {
    it("uses the HTTP ingest monitor URL", () => {
      const connection = new IngestMonitorConnection("http://example.com");
      expect(connection.url).toBe("http://example.com/ingest/monitor");
    });

    it("uses the HTTP storage monitor URL", () => {
      const connection = new StorageMonitorConnection("https://example.com");
      expect(connection.url).toBe("https://example.com/storage/monitor");
    });
  });

  describe("isConnected", () => {
    it("returns false when eventSource is null", () => {
      const connection = new IngestMonitorConnection("http://example.com");
      expect(connection.isConnected).toBe(false);
    });

    it("returns true when the event stream is open", async () => {
      const conn = new IngestMonitorConnection("http://localhost:1234");

      conn.dial();
      await vi.runAllTimersAsync();
      FakeEventSource.latest().open();

      expect(conn.isConnected).toBe(true);

      conn.close();
    });
  });

  describe("retryBackoff", () => {
    it("retries with exponential backoff until success", async () => {
      const connection = new IngestMonitorConnection("http://example.com", {
        initialDelay: 100,
        maxDelay: 500,
        backoff: 2,
        maxAttempts: 3,
        jitterFn: () => 0,
      });

      const fn = vi
        .fn()
        .mockRejectedValueOnce(new Error("Failed once"))
        .mockRejectedValueOnce(new Error("Failed twice"))
        .mockResolvedValueOnce(undefined);

      connection.retryBackoff(fn);
      await vi.runAllTimersAsync();
      expect(fn).toHaveBeenCalledTimes(3);
    });

    it("throws after max attempts", async () => {
      const connection = new IngestMonitorConnection("http://example.com", {
        initialDelay: 100,
        maxDelay: 500,
        backoff: 2,
        maxAttempts: 2,
        jitterFn: () => 0,
      });
      const fn = vi.fn().mockRejectedValue(new Error("Always fails"));

      const promise = expect(connection.retryBackoff(fn)).rejects.toThrow(
        "Max attempts reached",
      );
      await vi.runAllTimersAsync();
      await promise;

      expect(fn).toHaveBeenCalledTimes(2);
    });
  });

  describe("reconnect", () => {
    it.each(monitorConnectionCases)(
      "reconnects %s streams after the configured delay",
      async (_type, createConnection) => {
        const conn = createConnection({
          initialDelay: 100,
          maxDelay: 500,
          backoff: 2,
          maxAttempts: 5,
          jitterFn: () => 0,
        });

        await conn.dial();
        const source = FakeEventSource.latest();
        source.open();

        expect(conn.eventSource).toBe(source);
        expect(source.readyState).toBe(FakeEventSource.OPEN);

        source.error();

        expect(source.readyState).toBe(FakeEventSource.CLOSED);
        expect(FakeEventSource.instances).toHaveLength(1);

        await vi.advanceTimersByTimeAsync(99);
        expect(FakeEventSource.instances).toHaveLength(1);

        await vi.advanceTimersByTimeAsync(1);

        const reconnected = FakeEventSource.latest();
        expect(FakeEventSource.instances).toHaveLength(2);
        expect(reconnected).not.toBe(source);
        expect(conn.eventSource).toBe(reconnected);
        expect(reconnected.readyState).toBe(FakeEventSource.CONNECTING);

        reconnected.open();
        expect(conn.isConnected).toBe(true);

        conn.close();
      },
    );
  });
});

describe("IngestMonitorConnection", () => {
  it("sends the current bearer token in the event stream request", async () => {
    const authStore = useAuthStore();
    authStore.user = { access_token: "access-token" } as typeof authStore.user;
    const fetchMock = vi.fn<
      (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>
    >(() =>
      Promise.resolve(
        new Response("", {
          headers: { "content-type": "text/event-stream" },
          status: 200,
        }),
      ),
    );
    vi.stubGlobal("fetch", fetchMock);

    const conn = new IngestMonitorConnection("http://localhost:1234");
    await conn.dial();

    const source = FakeEventSource.latest();
    await source.connect();

    expect(fetchMock).toHaveBeenCalledTimes(1);
    const init = fetchMock.mock.calls[0][1] as RequestInit;
    const headers = new Headers(init.headers);
    expect(headers.get("Accept")).toBe("text/event-stream");
    expect(headers.get("Authorization")).toBe("Bearer access-token");

    conn.close();
  });

  it("retries server errors after the configured delay", async () => {
    const fetchMock = vi.fn<
      (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>
    >(() => Promise.resolve(new Response("", { status: 503 })));
    vi.stubGlobal("fetch", fetchMock);

    const conn = new IngestMonitorConnection("http://localhost:1234", {
      initialDelay: 100,
      maxDelay: 500,
      backoff: 2,
      maxAttempts: 5,
      jitterFn: () => 0,
    });

    await conn.dial();
    await FakeEventSource.latest().connect();

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(FakeEventSource.instances).toHaveLength(1);

    await vi.advanceTimersByTimeAsync(99);
    expect(FakeEventSource.instances).toHaveLength(1);

    await vi.advanceTimersByTimeAsync(1);

    expect(FakeEventSource.instances).toHaveLength(2);
    expect(FakeEventSource.latest().readyState).toBe(
      FakeEventSource.CONNECTING,
    );

    conn.close();
  });

  it("stops reconnecting and signs out after an unauthorized response", async () => {
    const fetchMock = vi.fn<
      (input: RequestInfo | URL, init?: RequestInit) => Promise<Response>
    >(() => Promise.resolve(new Response("", { status: 401 })));
    vi.stubGlobal("fetch", fetchMock);

    const conn = new IngestMonitorConnection("http://localhost:1234", {
      initialDelay: 100,
      maxDelay: 500,
      backoff: 2,
      maxAttempts: 5,
      jitterFn: () => 0,
    });

    await conn.dial();
    await FakeEventSource.latest().connect();

    expect(fetchMock).toHaveBeenCalledTimes(1);
    expect(handleUnauthorized).toHaveBeenCalledTimes(1);
    expect(conn.eventSource).toBeNull();

    await vi.runAllTimersAsync();

    expect(FakeEventSource.instances).toHaveLength(1);
  });

  it("connects to the event stream and receives event", async () => {
    const conn = new IngestMonitorConnection("http://localhost:1234");

    conn.dial();
    await vi.runAllTimersAsync();
    const source = FakeEventSource.latest();
    source.open();

    expect(conn.eventSource).toBe(source);
    expect(source.readyState).toBe(FakeEventSource.OPEN);

    source.message(
      JSON.stringify({
        value: {
          type: "sip_workflow_updated_event",
          value: {
            uuid: "workflow-1",
            item: {
              sip_uuid: "sip-1",
              status: "processing",
            },
          },
        },
      }),
    );
    await vi.runAllTimersAsync();

    expect(handleIngestEvent).toHaveBeenCalledWith({
      type: "sip_workflow_updated_event",
      value: {
        uuid: "workflow-1",
        item: {
          sip_uuid: "sip-1",
          status: "processing",
        },
      },
    });

    conn.close();

    expect(conn.eventSource).toBeNull();
  });

  it("ignores malformed SSE events", async () => {
    const conn = new IngestMonitorConnection("http://localhost:1234");

    conn.dial();
    await vi.runAllTimersAsync();
    const source = FakeEventSource.latest();

    source.message(
      JSON.stringify({
        value: {
          type: "sip_workflow_updated_event",
        },
      }),
    );
    await vi.runAllTimersAsync();

    expect(handleIngestEvent).not.toHaveBeenCalled();

    conn.close();
  });
});

describe("StorageMonitorConnection", () => {
  it("connects to the event stream and receives event", async () => {
    const conn = new StorageMonitorConnection("http://localhost:1234");
    conn.dial();
    await vi.runAllTimersAsync();
    const source = FakeEventSource.latest();
    source.open();

    expect(conn.eventSource).toBe(source);
    expect(source.readyState).toBe(FakeEventSource.OPEN);

    source.message(
      JSON.stringify({
        value: {
          type: "aip_location_updated_event",
          value: {
            uuid: "aip-1",
            location_uuid: "location-1",
          },
        },
      }),
    );
    await vi.runAllTimersAsync();

    expect(handleStorageEvent).toHaveBeenCalledWith({
      type: "aip_location_updated_event",
      value: {
        uuid: "aip-1",
        location_uuid: "location-1",
      },
    });

    conn.close();

    expect(conn.eventSource).toBeNull();
  });
});

describe("monitor event handlers", () => {
  it("updates SIP workflows from raw monitor event payloads", async () => {
    const { handleIngestEvent } =
      await vi.importActual<typeof import("@/monitor-ingest")>(
        "@/monitor-ingest",
      );

    const sipStore = useSipStore();
    const existingTasks = [
      { uuid: "task-1", status: "done" },
    ] as unknown as Array<EnduroIngestSipTask>;
    sipStore.current = { uuid: "sip-1" } as unknown as EnduroIngestSip;
    sipStore.currentWorkflows = {
      workflows: [
        {
          sipUuid: "sip-1",
          startedAt: new Date("2026-04-14T00:00:00Z"),
          status: "queued",
          tasks: existingTasks,
          temporalId: "temporal-1",
          type: "create aip",
          uuid: "workflow-1",
        },
      ],
    } as EnduroIngestSipWorkflows;
    const storedTasks = sipStore.currentWorkflows!.workflows![0].tasks;

    handleIngestEvent({
      type: IngestEventValueTypeEnum.SipWorkflowUpdatedEvent,
      value: {
        uuid: "workflow-1",
        item: {
          sip_uuid: "sip-1",
          started_at: "2026-04-14T01:00:00Z",
          status: "done",
          temporal_id: "temporal-2",
          type: "create aip",
          uuid: "workflow-1",
        },
      },
    });

    const workflow = sipStore.currentWorkflows!.workflows?.[0];
    expect(workflow?.status).toBe("done");
    expect(workflow?.temporalId).toBe("temporal-2");
    expect(workflow?.startedAt).toEqual(new Date("2026-04-14T01:00:00Z"));
    expect(workflow?.tasks).toBe(storedTasks);
  });

  it("updates batches from raw monitor event payloads", async () => {
    const { handleIngestEvent } =
      await vi.importActual<typeof import("@/monitor-ingest")>(
        "@/monitor-ingest",
      );

    const batchStore = useBatchStore();
    batchStore.current = {
      createdAt: new Date("2026-04-14T00:00:00Z"),
      identifier: "batch-1",
      sipsCount: 1,
      status: "queued",
      uuid: "batch-1",
    } as EnduroIngestBatch;
    batchStore.fetchBatchesDebounced = vi.fn();

    handleIngestEvent({
      type: IngestEventValueTypeEnum.BatchUpdatedEvent,
      value: {
        uuid: "batch-1",
        item: {
          created_at: "2026-04-14T01:00:00Z",
          identifier: "batch-1-updated",
          sips_count: 2,
          status: "processing",
          uuid: "batch-1",
        },
      },
    });

    expect(batchStore.fetchBatchesDebounced).toHaveBeenCalledWith(1);
    expect(batchStore.current!.identifier).toBe("batch-1-updated");
    expect(batchStore.current!.sipsCount).toBe(2);
    expect(batchStore.current!.status).toBe("processing");
    expect(batchStore.current!.createdAt).toEqual(
      new Date("2026-04-14T01:00:00Z"),
    );
  });

  it("updates AIP status and location from raw monitor event payloads", async () => {
    const { handleStorageEvent } =
      await vi.importActual<typeof import("@/monitor-storage")>(
        "@/monitor-storage",
      );

    const aipStore = useAipStore();
    aipStore.current = {
      createdAt: new Date("2026-04-14T00:00:00Z"),
      locationUuid: "location-1",
      name: "aip-1",
      objectKey: "object-1",
      status: "stored",
      uuid: "aip-1",
    } as EnduroStorageAip;
    aipStore.fetchAipsDebounced = vi.fn();

    handleStorageEvent({
      type: StorageEventValueTypeEnum.AipStatusUpdatedEvent,
      value: {
        uuid: "aip-1",
        status: "deleted",
      },
    });
    handleStorageEvent({
      type: StorageEventValueTypeEnum.AipLocationUpdatedEvent,
      value: {
        uuid: "aip-1",
        location_uuid: "location-2",
      },
    });

    expect(aipStore.fetchAipsDebounced).toHaveBeenCalledTimes(2);
    expect(aipStore.fetchAipsDebounced).toHaveBeenCalledWith(1);
    expect(aipStore.current!.status).toBe("deleted");
    expect(aipStore.current!.locationUuid).toBe("location-2");
  });

  it("updates AIP workflows from raw monitor event payloads", async () => {
    const { handleStorageEvent } =
      await vi.importActual<typeof import("@/monitor-storage")>(
        "@/monitor-storage",
      );

    const aipStore = useAipStore();
    const existingTasks = [
      { uuid: "task-1", status: "done" },
    ] as unknown as Array<EnduroStorageAipTask>;
    aipStore.current = { uuid: "aip-1" } as unknown as EnduroStorageAip;
    aipStore.currentWorkflows = {
      workflows: [
        {
          aipUuid: "aip-1",
          status: "queued",
          tasks: existingTasks,
          temporalId: "temporal-1",
          type: "delete aip",
          uuid: "workflow-1",
        },
      ],
    } as EnduroStorageAipWorkflows;
    const storedTasks = aipStore.currentWorkflows!.workflows![0].tasks;

    handleStorageEvent({
      type: StorageEventValueTypeEnum.AipWorkflowUpdatedEvent,
      value: {
        uuid: "workflow-1",
        item: {
          aip_uuid: "aip-1",
          status: "done",
          temporal_id: "temporal-2",
          type: "delete aip",
          uuid: "workflow-1",
        },
      },
    });

    const workflow = aipStore.currentWorkflows!.workflows?.[0];
    expect(workflow?.status).toBe("done");
    expect(workflow?.temporalId).toBe("temporal-2");
    expect(workflow?.tasks).toBe(storedTasks);
  });
});
