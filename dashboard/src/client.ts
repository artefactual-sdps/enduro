import * as api from "./openapi-generator";
import * as runtime from "./openapi-generator/runtime";
import { useAuthStore } from "@/stores/auth";
import { usePackageStore } from "./stores/package";

export interface Client {
  package: api.PackageApi;
  storage: api.StorageApi;
  upload: api.UploadApi;
  connectPackageMonitor: () => void;
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

function storageServiceDownloadURL(aipId: string): string {
  return (
    getPath() +
    `/storage/package/{aip_id}/download`.replace(
      `{${"aip_id"}}`,
      encodeURIComponent(aipId),
    )
  );
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

function connectPackageMonitor() {
  const store = usePackageStore();
  const url = getWebSocketURL() + "/package/monitor";
  const socket = new WebSocket(url);
  socket.onmessage = (event: MessageEvent) => {
    const body = JSON.parse(event.data);
    const data = api.MonitorEventFromJSON(body);
    if (data.event) {
      store.handleEvent(data.event);
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
            useAuthStore().removeUser();
            return Promise.resolve();
          }
          return Promise.resolve(context.response);
        },
      },
    ],
  });
  return {
    package: new api.PackageApi(config),
    storage: new api.StorageApi(config),
    upload: new api.UploadApi(config),
    connectPackageMonitor,
  };
}

const client = createClient();

export { api, runtime, client, storageServiceDownloadURL };
