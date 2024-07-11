import { UserManager, WebStorageStateStore } from "oidc-client-ts";
import type { User } from "oidc-client-ts";
import { defineStore } from "pinia";
import { Buffer } from "buffer";

type OIDCConfig = {
  enabled: boolean;
  provider: string;
  clientId: string;
  redirectUrl: string;
  extraScopes: string;
  extraQueryParams: string;
  abac: ABACConfig;
};

type ABACConfig = {
  enabled: boolean;
  claimPath: string;
  claimPathSeparator: string;
  claimValuePrefix: string;
};

export const useAuthStore = defineStore("auth", {
  state: () => ({
    config: {} as OIDCConfig,
    manager: null as UserManager | null,
    user: null as User | null,
    attributes: [] as string[],
  }),
  getters: {
    isEnabled(): boolean {
      return this.config.enabled;
    },
    isUserValid(): boolean {
      return !this.config.enabled || (this.user != null && !this.user.expired);
    },
    getUserDisplayName(): string | undefined {
      return (
        this.user?.profile.preferred_username ||
        this.user?.profile.name ||
        this.user?.profile.email
      );
    },
    getUserAccessToken(): string {
      return this.user ? this.user.access_token : "";
    },
    checkAttributes: (state) => {
      return (required: string[]): boolean => {
        if (
          !state.config.enabled ||
          !state.config.abac.enabled ||
          state.attributes.includes("*")
        ) {
          return true;
        }

        for (let attr of required) {
          while (true) {
            if (state.attributes.includes(attr)) {
              break;
            }
            const suffixIndex = attr.lastIndexOf(":*");
            if (suffixIndex !== -1) {
              attr = attr.substring(0, suffixIndex);
            }
            const lastColonIndex = attr.lastIndexOf(":");
            if (lastColonIndex === -1) {
              return false;
            }
            attr = attr.substring(0, lastColonIndex) + ":*";
          }
        }

        return true;
      };
    },
  },
  actions: {
    loadConfig() {
      // Config already loaded.
      if (Object.keys(this.config).length !== 0) {
        return;
      }

      this.config = {
        enabled: false,
        provider: "",
        clientId: "",
        redirectUrl: "",
        extraScopes: "",
        extraQueryParams: "",
        abac: {
          enabled: false,
          claimPath: "",
          claimPathSeparator: "",
          claimValuePrefix: "",
        },
      };

      const env = import.meta.env;
      if (env.VITE_OIDC_ENABLED) {
        this.config.enabled =
          env.VITE_OIDC_ENABLED.trim().toLowerCase() === "true";
      }
      if (env.VITE_OIDC_AUTHORITY) {
        this.config.provider = env.VITE_OIDC_AUTHORITY.trim();
      }
      if (env.VITE_OIDC_CLIENT_ID) {
        this.config.clientId = env.VITE_OIDC_CLIENT_ID.trim();
      }
      if (env.VITE_OIDC_REDIRECT_URI) {
        this.config.redirectUrl = env.VITE_OIDC_REDIRECT_URI.trim();
      }
      if (env.VITE_OIDC_EXTRA_SCOPES) {
        this.config.extraScopes = env.VITE_OIDC_EXTRA_SCOPES.trim();
      }
      if (env.VITE_OIDC_EXTRA_QUERY_PARAMS) {
        this.config.extraQueryParams = env.VITE_OIDC_EXTRA_QUERY_PARAMS.trim();
      }
      if (env.VITE_OIDC_ABAC_ENABLED) {
        this.config.abac.enabled =
          env.VITE_OIDC_ABAC_ENABLED.trim().toLowerCase() === "true";
      }
      if (env.VITE_OIDC_ABAC_CLAIM_PATH) {
        this.config.abac.claimPath = env.VITE_OIDC_ABAC_CLAIM_PATH.trim();
      }
      if (env.VITE_OIDC_ABAC_CLAIM_PATH_SEPARATOR) {
        this.config.abac.claimPathSeparator =
          env.VITE_OIDC_ABAC_CLAIM_PATH_SEPARATOR.trim();
      }
      if (env.VITE_OIDC_ABAC_CLAIM_VALUE_PREFIX) {
        this.config.abac.claimValuePrefix =
          env.VITE_OIDC_ABAC_CLAIM_VALUE_PREFIX.trim();
      }
    },
    loadManager() {
      // Manager already loaded.
      if (this.manager != null) {
        return;
      }

      // Manager not needed.
      this.loadConfig();
      if (!this.config.enabled) {
        return;
      }

      let scope = "openid email profile";
      if (this.config.extraScopes) {
        scope += " " + this.config.extraScopes;
      }

      let extraQueryParams: Record<string, string> = {};
      if (this.config.extraQueryParams) {
        this.config.extraQueryParams.split(",").forEach((param) => {
          const parts = param.trim().split("=");
          extraQueryParams[parts[0].trim()] = parts[1].trim();
        });
      }

      this.manager = new UserManager({
        authority: this.config.provider,
        client_id: this.config.clientId,
        redirect_uri: this.config.redirectUrl,
        extraQueryParams: extraQueryParams,
        scope: scope,
        userStore: new WebStorageStateStore({ store: window.localStorage }),
      });
    },
    signinRedirect() {
      this.loadManager();
      this.manager?.signinRedirect();
    },
    async signinCallback() {
      this.loadManager();
      this.setUser((await this.manager?.signinCallback()) || null);
    },
    // Load the currently authenticated user.
    async loadUser() {
      this.loadManager();
      this.setUser((await this.manager?.getUser()) || null);
    },
    async removeUser() {
      // TODO: end session upstream.
      this.loadManager();
      await this.manager?.removeUser();
      this.user = null;
      this.attributes = [];
    },
    setUser(user: User | null) {
      this.user = user;
      this.parseAttributes();
    },
    parseAttributes() {
      this.loadConfig();
      if (!this.config.enabled || !this.config.abac.enabled || !this.user) {
        return;
      }

      let data: Record<string, any>;
      try {
        data = JSON.parse(
          Buffer.from(
            this.user.access_token.split(".")[1],
            "base64",
          ).toString(),
        );
      } catch (err) {
        throw new Error(`Error decoding or parsing token: ${err}`);
      }

      const keys = this.config.abac.claimPathSeparator
        ? this.config.abac.claimPath.split(this.config.abac.claimPathSeparator)
        : [this.config.abac.claimPath];

      for (let i = 0; i < keys.length; i++) {
        const value = data[keys[i]];
        if (value === undefined) {
          throw new Error(
            `Attributes not found in token, claim path: ${this.config.abac.claimPath}`,
          );
        }

        if (i === keys.length - 1) {
          if (!Array.isArray(value)) {
            throw new Error(
              `Attributes are not part of a multivalue claim, claim path: ${this.config.abac.claimPath}`,
            );
          }

          this.attributes = value.reduce((acc: string[], item: any) => {
            if (
              typeof item === "string" &&
              item.startsWith(this.config.abac.claimValuePrefix)
            ) {
              acc.push(
                item.substring(this.config.abac.claimValuePrefix.length),
              );
            }
            return acc;
          }, []);

          return;
        }

        if (typeof value !== "object" || value === null) {
          throw new Error(
            `Attributes not found in token, claim path: ${this.config.abac.claimPath}`,
          );
        }

        data = value;
      }

      throw new Error("Unexpected error parsing attributes");
    },
  },
});
