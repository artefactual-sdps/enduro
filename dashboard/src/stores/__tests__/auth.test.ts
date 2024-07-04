import { useAuthStore } from "@/stores/auth";
import { flushPromises } from "@vue/test-utils";
import { User, UserManager } from "oidc-client-ts";
import { setActivePinia, createPinia } from "pinia";
import { expect, describe, it, vi, beforeEach } from "vitest";

describe("useAuthStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("sets and removes the user and attributes", async () => {
    var user = new User({
      access_token: "",
      token_type: "",
      profile: { aud: "", exp: 0, iat: 0, iss: "", sub: "" },
    });

    const manager = new UserManager({
      authority: "",
      client_id: "",
      redirect_uri: "",
    });

    const getUserMock = vi.fn().mockImplementation(manager.getUser);
    getUserMock.mockImplementation(async () => user);
    manager.getUser = getUserMock;

    const removeUserMock = vi.fn().mockImplementation(manager.removeUser);
    removeUserMock.mockImplementation(async () => null);
    manager.removeUser = removeUserMock;

    const authStore = useAuthStore();
    authStore.$patch((state) => (state.manager = manager));

    const parseAttrMock = vi.fn().mockImplementation(authStore.parseAttributes);
    parseAttrMock.mockImplementation(() => {
      authStore.$patch((state) => (state.attributes = ["*"]));
    });
    authStore.parseAttributes = parseAttrMock;

    expect(authStore.user).toEqual(null);
    expect(authStore.attributes).toEqual([]);
    authStore.loadUser();
    await flushPromises();
    expect(authStore.user).toEqual(user);
    expect(authStore.attributes).toEqual(["*"]);
    authStore.removeUser();
    await flushPromises();
    expect(authStore.user).toEqual(null);
    expect(authStore.attributes).toEqual([]);
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
