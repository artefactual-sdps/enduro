import { createTestingPinia } from "@pinia/testing";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { User } from "oidc-client-ts";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createRouter, createWebHistory } from "vue-router";

import AboutDialog from "@/components/AboutDialog.vue";
import Header from "@/components/Header.vue";
import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";

const openDialogMock = vi.hoisted(() => vi.fn());
vi.mock("@/dialogs/dialog", () => ({
  openDialog: openDialogMock,
}));

const router = createRouter({
  history: createWebHistory(),
  routes: [{ name: "/", path: "/", component: {} }],
});

describe("Header.vue", () => {
  afterEach(() => {
    cleanup();
    vi.resetAllMocks();
  });

  it("collapses and expands the sidebar", async () => {
    const { getByRole } = render(Header, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              layout: {
                sidebarCollapsed: false,
              },
            },
            stubActions: false,
          }),
          router,
        ],
      },
    });

    const layoutStore = useLayoutStore();

    const expandButton = getByRole("button", {
      name: "Collapse navigation",
    });

    await fireEvent.click(expandButton);
    expect(layoutStore.sidebarCollapsed).toEqual(true);

    const collapseButton = getByRole("button", {
      name: "Expand navigation",
    });

    await fireEvent.click(collapseButton);
    expect(layoutStore.sidebarCollapsed).toEqual(false);
  });

  it("displays the breadcrumb navigation", async () => {
    const { getByRole } = render(Header, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              layout: { breadcrumb: [{ text: "SIPs" }] },
            },
          }),
          router,
        ],
      },
    });

    getByRole("navigation", { name: "Breadcrumb" });
  });

  it("opens the About dialog from the user menu", async () => {
    openDialogMock.mockResolvedValue(undefined);
    const { getByRole } = render(Header, {
      global: {
        plugins: [createTestingPinia({ createSpy: vi.fn }), router],
      },
    });

    await fireEvent.click(getByRole("button", { name: "Open user menu" }));
    await fireEvent.click(getByRole("button", { name: "About" }));

    expect(openDialogMock).toHaveBeenCalledWith(AboutDialog);
  });

  it("shows the authenticated user menu and logs out", async () => {
    const pinia = createTestingPinia({
      createSpy: vi.fn,
      initialState: {
        auth: {
          config: { enabled: true },
          user: new User({
            access_token: "access-token",
            token_type: "Bearer",
            profile: {
              aud: "enduro",
              exp: 0,
              iat: 0,
              iss: "https://keycloak.example.com",
              sub: "user-id",
              email: "operator@example.com",
            },
          }),
        },
      },
    });
    const { getByRole, getByText } = render(Header, {
      global: { plugins: [pinia, router] },
    });

    await fireEvent.click(getByRole("button", { name: "Open user menu" }));
    getByText("operator@example.com");
    await fireEvent.click(getByRole("link", { name: "Sign out" }));

    expect(useAuthStore().signoutRedirect).toHaveBeenCalledOnce();
  });

  it("shows an unauthenticated menu without sign out when auth is disabled", async () => {
    const pinia = createTestingPinia({
      createSpy: vi.fn,
      initialState: { auth: { config: { enabled: false }, user: null } },
    });
    const { getByRole, getByText, queryByRole } = render(Header, {
      global: { plugins: [pinia, router] },
    });

    await fireEvent.click(getByRole("button", { name: "Open user menu" }));
    getByText("Unauthenticated");
    expect(queryByRole("link", { name: "Sign out" })).toBeNull();
  });
});
