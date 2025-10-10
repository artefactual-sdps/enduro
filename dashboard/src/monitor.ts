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

abstract class MonitorConnection {
  type: "ingest" | "storage";
  url: string;
  socket: WebSocket | null = null;
  isConnected: boolean = false;
  retry: RetryOptions;

  constructor(
    type: "ingest" | "storage",
    baseUrl: string,
    retry?: RetryOptions,
  ) {
    this.type = type;
    this.url = this.getWebSocketURL(baseUrl + "/" + this.type + "/monitor");
    this.retry = retry || {
      initialDelay: 1000, // 1 second.
      maxDelay: 30000, // 30 seconds.
      backoff: 2,
      maxAttempts: 10,
      jitterFn: () => Math.random() * 500, // add up to 500ms of jitter.
    };
  }

  abstract dial(): Promise<void>;

  getWebSocketURL(url: string): string {
    if (url.startsWith("https")) {
      url = "wss" + url.slice("https".length);
    } else if (url.startsWith("http")) {
      url = "ws" + url.slice("http".length);
    }

    return url;
  }

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
        const delay =
          Math.min(
            this.retry.initialDelay * this.retry.backoff ** attempt,
            this.retry.maxDelay,
          ) + this.retry.jitterFn();
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
    if (this.socket) {
      this.socket.onclose = null; // Disable reconnect on close.
      this.socket.close();
      this.socket = null;
    }
  }

  setupEventHandlers(): void {
    if (this.socket === null) {
      return;
    }

    this.socket.onopen = () => {
      this.isConnected = true;
      console.log(`${this.type} monitor socket connected`);
    };
    this.socket.onerror = () => {
      console.error(`${this.type} monitor socket error`);
    };
    this.socket.onclose = () => {
      this.isConnected = false;
      console.error(`${this.type} monitor socket closed, reconnecting...`);
      this.dial();
    };
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
            this.socket = new WebSocket(this.url);
            this.setupEventHandlers();
          } catch (err) {
            console.error("Failed to create Web Socket:", err);
            this.dial(); // Retry on failure.
          }
        })
        .catch((err) => {
          console.error("Ingest monitor request failed:", err);
          throw err;
        });
    });
  }

  setupEventHandlers(): void {
    if (this.socket === null) {
      return;
    }

    super.setupEventHandlers();

    // Handle incoming messages.
    this.socket.onmessage = (ev: MessageEvent) => {
      const body = JSON.parse(ev.data);
      const data = api.IngestEventFromJSON(body);
      if (data.ingestValue) {
        handleIngestEvent(data.ingestValue);
      }
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
            this.socket = new WebSocket(this.url);
            this.setupEventHandlers();
          } catch (err) {
            console.error("Failed to create Web Socket:", err);
            this.dial(); // Retry on failure.
          }
        })
        .catch((err) => {
          console.error("Storage monitor request failed:", err);
          throw err;
        });
    });
  }

  setupEventHandlers(): void {
    if (this.socket === null) {
      return;
    }

    super.setupEventHandlers();

    // Handle incoming messages.
    this.socket.onmessage = (ev: MessageEvent) => {
      const body = JSON.parse(ev.data);
      const data = api.StorageEventFromJSON(body);
      if (data.storageValue) {
        handleStorageEvent(data.storageValue);
      }
    };
  }
}
