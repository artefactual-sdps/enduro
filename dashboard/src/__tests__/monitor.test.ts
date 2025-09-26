import { createPinia, setActivePinia } from "pinia";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import WS from "vitest-websocket-mock";

import { client } from "@/client";
import { IngestMonitorConnection, StorageMonitorConnection } from "@/monitor";

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

      expect(() => connection.retryBackoff(fn)).rejects.toThrow(
        "Max attempts reached",
      );
      await vi.runAllTimersAsync();

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
        ingest_value: {
          Type: "ingest_ping_event",
          Value: "Ping",
        },
      }),
    );
    await vi.runAllTimersAsync();

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
    expect(conn.dial()).rejects.toThrow("Max attempts reached");
    await vi.runAllTimersAsync();

    expect(consoleMock).toHaveBeenCalledTimes(1);
    expect(consoleMock).toHaveBeenCalledWith(
      "Ingest monitor request failed:",
      new Error("401 Unauthorized"),
    );

    server.close();

    // Restore the mock for other tests.
    client.ingest.ingestMonitorRequest = vi.fn(() => Promise.resolve());
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
        storage_value: {
          Type: "storage_ping_event",
          Value: "Ping",
        },
      }),
    );
    await vi.runAllTimersAsync();

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
    expect(conn.dial()).rejects.toThrow("Max attempts reached");
    await vi.runAllTimersAsync();

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
