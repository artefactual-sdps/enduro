import { useLayoutStore } from "@/stores/layout";
import { flushPromises } from "@vue/test-utils";
import { User } from "oidc-client-ts";
import { setActivePinia, createPinia } from "pinia";
import { expect, describe, it, beforeEach } from "vitest";

describe("useLayoutStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("toggles the sidebarCollapsed property", () => {
    const layoutStore = useLayoutStore();
    layoutStore.sidebarCollapsed = false;

    layoutStore.toggleSidebar();
    expect(layoutStore.sidebarCollapsed).toEqual(true);
  });

  it("updates the breadcrumb property", () => {
    const layoutStore = useLayoutStore();
    const breadcrumb = [{ text: "Packages" }];

    layoutStore.updateBreadcrumb(breadcrumb);
    expect(layoutStore.breadcrumb).toEqual(breadcrumb);
  });

  it("sets and removes the user", async () => {
    const layoutStore = useLayoutStore();
    var user = new User({
      access_token: "",
      token_type: "",
      profile: { aud: "", exp: 0, iat: 0, iss: "", sub: "" },
    });
    expect(layoutStore.user).toEqual(null);
    layoutStore.setUser(user);
    expect(layoutStore.user).toEqual(user);
    layoutStore.removeUser();
    await flushPromises();
    expect(layoutStore.user).toEqual(null);
  });

  it("checks if the user is valid", () => {
    const layoutStore = useLayoutStore();

    // No user.
    expect(layoutStore.isUserValid).toEqual(false);

    // Expired user.
    var user = new User({
      access_token: "",
      token_type: "",
      profile: { aud: "", exp: 0, iat: 0, iss: "", sub: "" },
      expires_at: Date.now() / 1000 - 1,
    });
    layoutStore.setUser(user);
    expect(layoutStore.isUserValid).toEqual(false);

    // Valid user.
    user = new User({
      access_token: "",
      token_type: "",
      profile: { aud: "", exp: 0, iat: 0, iss: "", sub: "" },
      expires_at: Date.now() / 1000 + 1,
    });
    layoutStore.setUser(user);
    expect(layoutStore.isUserValid).toEqual(true);
  });

  it("gets the user display name", () => {
    const layoutStore = useLayoutStore();

    // No user.
    expect(layoutStore.getUserDisplayName).toEqual(undefined);

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
    layoutStore.setUser(user);
    expect(layoutStore.getUserDisplayName).toEqual("preferred_username");

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
    layoutStore.setUser(user);
    expect(layoutStore.getUserDisplayName).toEqual("name");

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
    layoutStore.setUser(user);
    expect(layoutStore.getUserDisplayName).toEqual("name@example.com");
  });
});
