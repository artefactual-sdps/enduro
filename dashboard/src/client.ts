import * as api from "./openapi-generator";
import type { InjectionKey } from "vue";

export interface Client {
  package: api.PackageApi;
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
  };
}

const clientProviderKey = Symbol() as InjectionKey<Client>;

export { api, createClient, clientProviderKey };
