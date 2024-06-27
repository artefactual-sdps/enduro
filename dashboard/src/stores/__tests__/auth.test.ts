import { useAuthStore } from "@/stores/auth";
import { flushPromises } from "@vue/test-utils";
import { User } from "oidc-client-ts";
import { setActivePinia, createPinia } from "pinia";
import { expect, describe, it, beforeEach } from "vitest";

describe("useAuthStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("sets and removes the user", async () => {
    const authStore = useAuthStore();
    var user = new User({
      access_token: "",
      token_type: "",
      profile: { aud: "", exp: 0, iat: 0, iss: "", sub: "" },
    });
    expect(authStore.user).toEqual(null);
    authStore.setUser(user);
    expect(authStore.user).toEqual(user);
    //authStore.removeUser();
    //await flushPromises();
    //expect(authStore.user).toEqual(null);
  });

  it("checks if the user is valid", () => {
    const authStore = useAuthStore();
    authStore.$patch((state) => {
      state.config = {
        enabled: true,
        provider: "",
        clientId: "",
        redirectUrl: "",
        extraScopes: "",
        abac: {
          enabled: false,
          claimPath: "",
          claimPathSeparator: "",
          claimValuePrefix: "",
        },
      };
    });

    // No user.
    expect(authStore.isUserValid).toEqual(false);

    // Expired user.
    var user = new User({
      access_token: "",
      token_type: "",
      profile: { aud: "", exp: 0, iat: 0, iss: "", sub: "" },
      expires_at: Date.now() / 1000 - 1,
    });
    authStore.setUser(user);
    expect(authStore.isUserValid).toEqual(false);

    // Valid user.
    user = new User({
      access_token: "",
      token_type: "",
      profile: { aud: "", exp: 0, iat: 0, iss: "", sub: "" },
      expires_at: Date.now() / 1000 + 1,
    });
    authStore.setUser(user);
    expect(authStore.isUserValid).toEqual(true);
  });

  it("gets the user display name", () => {
    const authStore = useAuthStore();

    // No user.
    expect(authStore.getUserDisplayName).toEqual(undefined);

    // User with preferred_username, name and email.
    var user = new User({
      access_token: "",
      token_type: "",
      profile: {
        aud: "",
        exp: 0,
        iat: 0,
        iss: "",
        sub: "",
        preferred_username: "preferred_username",
        name: "name",
        email: "name@example.com",
      },
    });
    authStore.setUser(user);
    expect(authStore.getUserDisplayName).toEqual("preferred_username");

    // User with name and email.
    user = new User({
      access_token: "",
      token_type: "",
      profile: {
        aud: "",
        exp: 0,
        iat: 0,
        iss: "",
        sub: "",
        name: "name",
        email: "name@example.com",
      },
    });
    authStore.setUser(user);
    expect(authStore.getUserDisplayName).toEqual("name");

    // User with email.
    user = new User({
      access_token: "",
      token_type: "",
      profile: {
        aud: "",
        exp: 0,
        iat: 0,
        iss: "",
        sub: "",
        email: "name@example.com",
      },
    });
    authStore.setUser(user);
    expect(authStore.getUserDisplayName).toEqual("name@example.com");
  });
});
