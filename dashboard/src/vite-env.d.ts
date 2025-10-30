/// <reference types="vite/client" />
/// <reference types="vue/macros-global" />
/// <reference types="unplugin-icons/types/vue" />
/// <reference types="unplugin-vue-router/client" />

declare module "*.vue" {
  import type { DefineComponent } from "vue";
  const component: DefineComponent<object, object, unknown>;
  export default component;
}

interface ImportMetaEnv {
  readonly VITE_OIDC_ENABLED: string;
  readonly VITE_OIDC_BASE_URL: string;
  readonly VITE_OIDC_AUTHORITY: string;
  readonly VITE_OIDC_CLIENT_ID: string;
  readonly VITE_OIDC_SCOPES: string;
  readonly VITE_OIDC_EXTRA_QUERY_PARAMS: string;
  readonly VITE_OIDC_ABAC_ENABLED: string;
  readonly VITE_OIDC_ABAC_CLAIM_PATH: string;
  readonly VITE_OIDC_ABAC_CLAIM_PATH_SEPARATOR: string;
  readonly VITE_OIDC_ABAC_CLAIM_VALUE_PREFIX: string;
  readonly VITE_OIDC_ABAC_USE_ROLES: string;
  readonly VITE_OIDC_ABAC_ROLES_MAPPING: string;
  readonly VITE_INSTITUTION_LOGO: string;
  readonly VITE_INSTITUTION_NAME: string;
  readonly VITE_INSTITUTION_URL: string;
  readonly VITE_CUSTOM_MANIFEST: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
