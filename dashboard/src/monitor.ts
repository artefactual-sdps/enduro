import { api, client } from "@/client";
import { handleIngestEvent } from "@/monitor-ingest";
import { handleStorageEvent } from "@/monitor-storage";

export type RetryOptions = {
  initialDelay: number;
  maxDelay: number;
  backoff: number;
  maxAttempts: number;
  jitterFn: () => number;
};

function isObject(value: unknown): value is Record<string, unknown> {
  return value !== null && typeof value === "object";
}

// Extract the raw monitor event envelope from the SSE payload.
// The generic only narrows the returned `type` for the caller; it does not
// validate that the runtime event type belongs to a specific enum.
function parseMonitorEvent<T extends string>(
  body: unknown,
): { type: T; value: unknown } | null {
  if (!isObject(body) || !isObject(body.value)) return null;

  const event = body.value;
  if (typeof event.type !== "string" || !("value" in event)) return null;

  return { type: event.type as T, value: event.value };
}

abstract class MonitorConnection {
  type: "ingest" | "storage";
  url: string;
  eventSource: EventSource | null = null;
  isConnected: boolean = false;
  retry: RetryOptions;
  private closed: boolean = false;
  private reconnectAttempts: number = 0;
  private reconnecting: boolean = false;
  private reconnectTimer: ReturnType<typeof setTimeout> | null = null;
  private reconnectTimerResolve: (() => void) | null = null;

  constructor(
    type: "ingest" | "storage",
    baseUrl: string,
    retry?: RetryOptions,
  ) {
    this.type = type;
    this.url = baseUrl + "/" + this.type + "/monitor";
    this.retry = retry || {
      initialDelay: 1000, // 1 second.
      maxDelay: 30000, // 30 seconds.
      backoff: 2,
      maxAttempts: 10,
      jitterFn: () => Math.random() * 500, // add up to 500ms of jitter.
    };
  }

  abstract dial(): Promise<void>;

  async retryBackoff(fn: () => Promise<void>) {
    for (let attempt = 0; attempt < this.retry.maxAttempts; attempt++) {
      try {
        await fn();
        return;
      } catch {
        if (attempt === this.retry.maxAttempts - 1) {
          throw new Error("Max attempts reached");
        }

        // Exponential backoff with jitter.
        const delay = this.retryDelay(attempt);
        console.log(
          `${this.type} monitor: reconnect attempt ${attempt + 1} in ${Math.round(delay)} ms...`,
        );
        await new Promise((r) => {
          setTimeout(r, delay);
        });
      }
    }
  }

  close(): void {
    this.closed = true;
    this.isConnected = false;
    this.reconnectAttempts = 0;
    this.reconnecting = false;

    if (this.reconnectTimer !== null) {
      clearTimeout(this.reconnectTimer);
      this.reconnectTimer = null;
      this.reconnectTimerResolve?.();
      this.reconnectTimerResolve = null;
    }

    if (this.eventSource) {
      this.eventSource.onerror = null;
      this.eventSource.close();
      this.eventSource = null;
    }
  }

  setupEventHandlers(): void {
    if (this.eventSource === null) {
      return;
    }

    this.eventSource.onopen = () => {
      this.isConnected = true;
      this.reconnectAttempts = 0;
      console.log(`${this.type} monitor event stream connected`);
    };
    this.eventSource.onerror = () => {
      this.isConnected = false;
      console.error(`${this.type} monitor event stream error, reconnecting...`);
      if (this.eventSource) {
        this.eventSource.onerror = null;
        this.eventSource.close();
        this.eventSource = null;
      }
      if (!this.closed) {
        this.reconnect();
      }
    };
  }

  protected openEventSource(): void {
    this.closed = false;
    this.eventSource = new EventSource(this.url);
    this.setupEventHandlers();
  }

  private retryDelay(attempt: number): number {
    return (
      Math.min(
        this.retry.initialDelay * this.retry.backoff ** attempt,
        this.retry.maxDelay,
      ) + this.retry.jitterFn()
    );
  }

  private async waitForReconnectDelay(delay: number): Promise<void> {
    await new Promise<void>((resolve) => {
      this.reconnectTimerResolve = resolve;
      this.reconnectTimer = setTimeout(() => {
        this.reconnectTimer = null;
        this.reconnectTimerResolve = null;
        resolve();
      }, delay);
    });
  }

  private reconnect(): void {
    if (this.reconnecting) return;

    this.reconnecting = true;
    void this.reconnectBackoff().finally(() => {
      this.reconnecting = false;
    });
  }

  private async reconnectBackoff(): Promise<void> {
    while (!this.closed) {
      if (this.reconnectAttempts >= this.retry.maxAttempts) {
        console.error(
          `${this.type} monitor reconnect failed:`,
          new Error("Max attempts reached"),
        );
        return;
      }

      const attempt = this.reconnectAttempts;
      const delay = this.retryDelay(attempt);
      this.reconnectAttempts++;

      console.log(
        `${this.type} monitor: reconnect attempt ${attempt + 1} in ${Math.round(delay)} ms...`,
      );
      await this.waitForReconnectDelay(delay);
      if (this.closed) return;

      try {
        await this.dial();
        return;
      } catch (err) {
        console.error("Failed to create monitor event stream:", err);
      }
    }
  }
}

export class IngestMonitorConnection extends MonitorConnection {
  constructor(baseUrl: string, retry?: RetryOptions) {
    super("ingest", baseUrl, retry);
  }

  async dial(): Promise<void> {
    return this.retryBackoff(async () => {
      return client.ingest
        .ingestMonitorRequest()
        .then(() => {
          try {
            this.openEventSource();
          } catch (err) {
            console.error("Failed to create monitor event stream:", err);
            throw err;
          }
        })
        .catch((err) => {
          console.error("Ingest monitor request failed:", err);
          throw err;
        });
    });
  }

  setupEventHandlers(): void {
    if (this.eventSource === null) {
      return;
    }

    super.setupEventHandlers();

    // Handle incoming messages.
    this.eventSource.onmessage = (ev: MessageEvent) => {
      const body = JSON.parse(ev.data);
      const data = parseMonitorEvent<api.IngestEventValueTypeEnum>(body);
      if (data) handleIngestEvent(data);
    };
  }
}

export class StorageMonitorConnection extends MonitorConnection {
  constructor(baseUrl: string, retry?: RetryOptions) {
    super("storage", baseUrl, retry);
  }

  async dial(): Promise<void> {
    return this.retryBackoff(async () => {
      return client.storage
        .storageMonitorRequest()
        .then(() => {
          try {
            this.openEventSource();
          } catch (err) {
            console.error("Failed to create monitor event stream:", err);
            throw err;
          }
        })
        .catch((err) => {
          console.error("Storage monitor request failed:", err);
          throw err;
        });
    });
  }

  setupEventHandlers(): void {
    if (this.eventSource === null) {
      return;
    }

    super.setupEventHandlers();

    // Handle incoming messages.
    this.eventSource.onmessage = (ev: MessageEvent) => {
      const body = JSON.parse(ev.data);
      const data = parseMonitorEvent<api.StorageEventValueTypeEnum>(body);
      if (data) handleStorageEvent(data);
    };
  }
}
