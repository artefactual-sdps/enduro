import { handleIngestEvent } from "./monitor-ingest";
import { handleStorageEvent } from "./monitor-storage";
import * as api from "./openapi-generator";
import * as runtime from "./openapi-generator/runtime";

import router from "@/router";
import { useAuthStore } from "@/stores/auth";

export interface Client {
  about: api.AboutApi;
  ingest: api.IngestApi;
  storage: api.StorageApi;
  connectIngestMonitor: () => void;
  connectStorageMonitor: () => void;
}

function getPath(): string {
  const location = window.location;
  const path =
    location.protocol +
    "//" +
    location.hostname +
    (location.port ? ":" + location.port : "") +
    "/api";

  return path.replace(/\/$/, "");
}

function getWebSocketURL(): string {
  let url = getPath();

  if (url.startsWith("https")) {
    url = "wss" + url.slice("https".length);
  } else if (url.startsWith("http")) {
    url = "ws" + url.slice("http".length);
  }

  return url;
}

function connectIngestMonitor() {
  const url = getWebSocketURL() + "/ingest/monitor";
  const socket = new WebSocket(url);
  socket.onmessage = (event: MessageEvent) => {
    const body = JSON.parse(event.data);
    const data = api.IngestEventFromJSON(body);
    if (data.ingestValue) {
      handleIngestEvent(data.ingestValue);
    }
  };
}

function connectStorageMonitor() {
  const url = getWebSocketURL() + "/storage/monitor";
  const socket = new WebSocket(url);
  socket.onmessage = (event: MessageEvent) => {
    const body = JSON.parse(event.data);
    const data = api.StorageEventFromJSON(body);
    if (data.storageValue) {
      handleStorageEvent(data.storageValue);
    }
  };
}

function createClient(): Client {
  const config: api.Configuration = new api.Configuration({
    basePath: getPath(),
    accessToken: () => useAuthStore().getUserAccessToken,
    middleware: [
      {
        post(context) {
          if (context.response.status == 401) {
            useAuthStore()
              .removeUser()
              .then(() => router.push({ name: "/user/signin" }));
            return Promise.resolve();
          }
          return Promise.resolve(context.response);
        },
      },
    ],
  });
  return {
    about: new api.AboutApi(config),
    ingest: new api.IngestApi(config),
    storage: new api.StorageApi(config),
    connectIngestMonitor: connectIngestMonitor,
    connectStorageMonitor: connectStorageMonitor,
  };
}

const client = createClient();

export { api, client, getPath, runtime };
