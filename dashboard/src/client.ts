import * as api from "./openapi-generator";

export interface Client {
  package: api.PackageApi;
  storage: api.StorageApi;
}

function getPath(): string {
  const base = "/api";
  const location = window.location;
  const path =
    location.protocol +
    "//" +
    location.hostname +
    (location.port ? ":" + location.port : "") +
    base +
    (location.search ? location.search : "");

  return path.replace(/\/$/, "");
}

function createClient(): Client {
  const path = getPath();
  const config: api.Configuration = new api.Configuration({ basePath: path });

  // tslint:disable-next-line:no-console
  console.log("Enduro client created", path);

  return {
    package: new api.PackageApi(config),
    storage: new api.StorageApi(config),
  };
}

function storageServiceDownloadURL(aipId: string): string {
  return (
    getPath() +
    `/storage/{aip_id}/download`.replace(
      `{${"aip_id"}}`,
      encodeURIComponent(aipId)
    )
  );
}

const client = createClient();

export { api, client, storageServiceDownloadURL };
