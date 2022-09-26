import { UserManager, WebStorageStateStore } from "oidc-client-ts";

export default new UserManager({
  authority: import.meta.env.VITE_OIDC_AUTHORITY,
  client_id: import.meta.env.VITE_OIDC_CLIENT_ID,
  redirect_uri: import.meta.env.VITE_OIDC_REDIRECT_URI,
  scope:
    "openid email profile offline_access audience:server:client_id:" +
    import.meta.env.VITE_OIDC_AUDIENCE_CLIENT_ID,
  userStore: new WebStorageStateStore({ store: window.localStorage }),
});
