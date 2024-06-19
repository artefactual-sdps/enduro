import { UserManager, WebStorageStateStore } from "oidc-client-ts";

let scope = "openid email profile";
if (import.meta.env.VITE_OIDC_EXTRA_SCOPES != undefined) {
  scope += " " + import.meta.env.VITE_OIDC_EXTRA_SCOPES;
}

export default new UserManager({
  authority: import.meta.env.VITE_OIDC_AUTHORITY,
  client_id: import.meta.env.VITE_OIDC_CLIENT_ID,
  redirect_uri: import.meta.env.VITE_OIDC_REDIRECT_URI,
  scope: scope,
  userStore: new WebStorageStateStore({ store: window.localStorage }),
});
