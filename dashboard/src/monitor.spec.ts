import { createPinia, setActivePinia } from "pinia";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import WS from "vitest-websocket-mock";

import { client } from "@/client";
import { IngestMonitorConnection, StorageMonitorConnection } from "@/monitor";
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
import { useBatchStore } from "@/stores/batch";
import { useSipStore } from "@/stores/sip";

vi.mock("@/client", async () => {
  const api = await vi.importActual("@/openapi-generator");
  return {
    client: {
      ingest: {
        ingestMonitorRequest: vi.fn(() => Promise.resolve()),
      },
      storage: {
        storageMonitorRequest: vi.fn(() => Promise.resolve()),
      },
    },
    api: { ...api },
  };
});

vi.mock("@/monitor-ingest", () => ({
  handleIngestEvent: vi.fn(),
}));

vi.mock("@/monitor-storage", () => ({
  handleStorageEvent: vi.fn(),
}));

// Mock timers.
beforeEach(() => {
  setActivePinia(createPinia());
  vi.useFakeTimers();
});
afterEach(() => {
  vi.resetAllMocks();
});

describe("MonitorConnection", () => {
  describe("getWebSocketURL", () => {
    it("converts http to ws", () => {
      const connection = new IngestMonitorConnection("http://example.com");
      expect(connection.getWebSocketURL("http://example.com/path")).toBe(
        "ws://example.com/path",
      );
    });

    it("converts https to wss", () => {
      const connection = new IngestMonitorConnection("https://example.com");
      expect(connection.getWebSocketURL("https://example.com/path")).toBe(
        "wss://example.com/path",
      );
    });
  });

  describe("isConnected", () => {
    it("returns false when socket is null", () => {
      const connection = new IngestMonitorConnection("http://example.com");
      expect(connection.isConnected).toBe(false);
    });

    it("returns true when socket is in OPEN state", async () => {
      const server = new WS("ws://localhost:1234/ingest/monitor");
      const conn = new IngestMonitorConnection("http://localhost:1234");

      conn.dial();
      await vi.runAllTimersAsync();

      expect(client.ingest.ingestMonitorRequest).toHaveBeenCalledTimes(1);
      expect(conn.isConnected).toBe(true);

      conn.close();
      server.close();
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
});

describe("IngestMonitorConnection", () => {
  it("connects to WebSocket and receives event", async () => {
    const server = new WS("ws://localhost:1234/ingest/monitor");
    const conn = new IngestMonitorConnection("http://localhost:1234");

    conn.dial();
    await vi.runAllTimersAsync();

    expect(client.ingest.ingestMonitorRequest).toHaveBeenCalledTimes(1);
    expect(conn.socket).toBeDefined();
    expect(conn.socket!.readyState).toBe(WebSocket.OPEN);

    server.send(
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
    server.close();

    expect(conn.socket).toBeNull();
  });

  it("reconnects after the WebSocket is closed", async () => {
    let server = new WS("ws://localhost:1234/ingest/monitor");
    const conn = new IngestMonitorConnection("http://localhost:1234", {
      initialDelay: 100,
      maxDelay: 500,
      backoff: 2,
      maxAttempts: 5,
      jitterFn: () => 0,
    });
    conn.dial();
    await vi.runAllTimersAsync();

    expect(client.ingest.ingestMonitorRequest).toHaveBeenCalledTimes(1);
    expect(conn.socket).toBeDefined();
    expect(conn.socket!.readyState).toBe(WebSocket.OPEN);

    // Close the server to trigger reconnect.
    server.close();
    await vi.advanceTimersByTimeAsync(100);

    expect(conn.socket).toBeDefined();
    expect(conn.socket!.readyState).toBe(WebSocket.CONNECTING);

    server = new WS("ws://localhost:1234/ingest/monitor");
    await vi.advanceTimersByTimeAsync(100);

    expect(conn.socket).toBeDefined();
    expect(conn.socket!.readyState).toBe(WebSocket.OPEN);

    conn.close();
    server.close();
  });

  it("logs an error when ingest monitor request fails", async () => {
    client.ingest.ingestMonitorRequest = vi.fn(() =>
      Promise.reject(new Error("401 Unauthorized")),
    );
    const consoleMock = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);

    const server = new WS("ws://localhost:1234/ingest/monitor");
    const conn = new IngestMonitorConnection("http://localhost:1234", {
      initialDelay: 100,
      maxDelay: 500,
      backoff: 2,
      maxAttempts: 1,
      jitterFn: () => 0,
    });
    const promise = expect(conn.dial()).rejects.toThrow("Max attempts reached");
    await vi.runAllTimersAsync();
    await promise;

    expect(consoleMock).toHaveBeenCalledTimes(1);
    expect(consoleMock).toHaveBeenCalledWith(
      "Ingest monitor request failed:",
      new Error("401 Unauthorized"),
    );

    server.close();

    // Restore the mock for other tests.
    client.ingest.ingestMonitorRequest = vi.fn(() => Promise.resolve());
  });

  it("ignores malformed websocket events", async () => {
    const server = new WS("ws://localhost:1234/ingest/monitor");
    const conn = new IngestMonitorConnection("http://localhost:1234");

    conn.dial();
    await vi.runAllTimersAsync();

    server.send(
      JSON.stringify({
        value: {
          type: "sip_workflow_updated_event",
        },
      }),
    );
    await vi.runAllTimersAsync();

    expect(handleIngestEvent).not.toHaveBeenCalled();

    conn.close();
    server.close();
  });
});

describe("StorageMonitorConnection", () => {
  it("connects to WebSocket and receives event", async () => {
    const server = new WS("ws://localhost:1234/storage/monitor");
    const conn = new StorageMonitorConnection("http://localhost:1234");
    conn.dial();
    await vi.runAllTimersAsync();

    expect(client.storage.storageMonitorRequest).toHaveBeenCalledTimes(1);
    expect(conn.socket).toBeDefined();
    expect(conn.socket!.readyState).toBe(WebSocket.OPEN);

    server.send(
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
    server.close();

    expect(conn.socket).toBeNull();
  });

  it("reconnects after the WebSocket is closed", async () => {
    let server = new WS("ws://localhost:1234/storage/monitor");
    const conn = new StorageMonitorConnection("http://localhost:1234", {
      initialDelay: 100,
      maxDelay: 500,
      backoff: 2,
      maxAttempts: 5,
      jitterFn: () => 0,
    });
    conn.dial();
    await vi.runAllTimersAsync();

    expect(conn.socket).toBeDefined();
    expect(conn.socket!.readyState).toBe(WebSocket.OPEN);

    // Close the server to trigger reconnect.
    server.close();
    await vi.advanceTimersByTimeAsync(100);

    expect(conn.socket).toBeDefined();
    expect(conn.socket!.readyState).toBe(WebSocket.CONNECTING);

    server = new WS("ws://localhost:1234/storage/monitor");
    await vi.advanceTimersByTimeAsync(100);

    expect(conn.socket).toBeDefined();
    expect(conn.socket!.readyState).toBe(WebSocket.OPEN);

    conn.close();
    server.close();
  });

  it("logs an error when storage monitor request fails", async () => {
    client.storage.storageMonitorRequest = vi.fn(() =>
      Promise.reject(new Error("401 Unauthorized")),
    );
    const consoleMock = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);

    const server = new WS("ws://localhost:1234/storage/monitor");
    const conn = new StorageMonitorConnection("http://localhost:1234", {
      initialDelay: 100,
      maxDelay: 500,
      backoff: 2,
      maxAttempts: 1,
      jitterFn: () => 0,
    });
    const promise = expect(conn.dial()).rejects.toThrow("Max attempts reached");
    await vi.runAllTimersAsync();
    await promise;

    expect(consoleMock).toHaveBeenCalledTimes(1);
    expect(consoleMock).toHaveBeenCalledWith(
      "Storage monitor request failed:",
      new Error("401 Unauthorized"),
    );

    server.close();

    // Restore the mock for other tests.
    client.storage.storageMonitorRequest = vi.fn(() => Promise.resolve());
  });
});

describe("monitor event handlers", () => {
  it("updates SIP workflows from raw websocket payloads", async () => {
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

  it("updates batches from raw websocket payloads", async () => {
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

  it("updates AIP status and location from raw websocket payloads", async () => {
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

  it("updates AIP workflows from raw websocket payloads", async () => {
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
