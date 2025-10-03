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

export interface MonitorConnector {
  connect(): Promise<void>;
  isConnected(): boolean;
  getWebSocketURL(url: string): string;
}

export class MonitorConnection {
  type: "ingest" | "storage";
  url: string;
  socket: WebSocket | null = null;
  retry: RetryOptions;

  constructor(
    type: "ingest" | "storage",
    baseUrl: string,
    retry?: RetryOptions,
  ) {
    this.type = type;
    this.url = this.getWebSocketURL(baseUrl + "/" + this.type + "/monitor");
    this.retry = retry || {
      initialDelay: 1000,
      maxDelay: 30000,
      backoff: 2,
      maxAttempts: 10,
      jitterFn: () => Math.random() * 500,
    };
  }

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

  isConnected(): boolean {
    return this.socket !== null && this.socket.readyState === WebSocket.OPEN;
  }

  async connectSocket(url: string): Promise<void> {
    this.socket = new WebSocket(url);

    this.socket.onopen = () => {
      console.log(`${this.type} monitor socket connected`);
    };
    this.socket.onerror = () => {
      console.error(`${this.type} monitor socket error`);
    };
  }
}

export class IngestMonitorConnection
  extends MonitorConnection
  implements MonitorConnector
{
  constructor(baseUrl: string, retry?: RetryOptions) {
    super("ingest", baseUrl, retry);
  }

  async connect(): Promise<void> {
    return this.retryBackoff(async () => {
      return client.ingest.ingestMonitorRequest().then(async () => {
        return this.connectSocket(this.url).then(() => {
          if (this.socket) {
            // Reconnect on close.
            this.socket.onclose = () => {
              console.error("ingest monitor socket closed, reconnecting...");
              this.connect();
            };

            // Handle incoming messages.
            this.socket.onmessage = (ev: MessageEvent) => {
              const body = JSON.parse(ev.data);
              const data = api.IngestEventFromJSON(body);
              if (data.ingestValue) {
                handleIngestEvent(data.ingestValue);
              }
            };
          }
        });
      });
    });
  }
}

export class StorageMonitorConnection
  extends MonitorConnection
  implements MonitorConnector
{
  constructor(baseUrl: string, retry?: RetryOptions) {
    super("storage", baseUrl, retry);
  }

  async connect(): Promise<void> {
    return this.retryBackoff(async () => {
      return client.storage.storageMonitorRequest().then(async () => {
        return this.connectSocket(this.url).then(() => {
          if (this.socket) {
            // Reconnect on close.
            this.socket.onclose = () => {
              console.error("storage monitor socket closed, reconnecting...");
              this.connect();
            };

            // Handle incoming messages.
            this.socket.onmessage = (ev: MessageEvent) => {
              const body = JSON.parse(ev.data);
              const data = api.StorageEventFromJSON(body);
              if (data.storageValue) {
                handleStorageEvent(data.storageValue);
              }
            };
          }
        });
      });
    });
  }
}
