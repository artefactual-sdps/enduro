import { createTestingPinia } from "@pinia/testing";
import { render } from "@testing-library/vue";
import { User } from "oidc-client-ts";
import { describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";

import App from "@/App.vue";
import AboutDialog from "@/components/AboutDialog.vue";
import { openDialog } from "@/dialogs/dialog";
import { useAuthStore } from "@/stores/auth";

function createValidUser() {
  return new User({
    access_token: "",
    token_type: "",
    profile: { aud: "", exp: 0, iat: 0, iss: "", sub: "" },
    expires_at: Date.now() / 1000 + 60,
  });
}

describe("App.vue", () => {
  it("tears down an open dialog when the user session expires", async () => {
    const { findByRole, queryByRole } = render(App, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              auth: {
                config: { enabled: true },
                user: createValidUser(),
              },
            },
          }),
        ],
        stubs: {
          Header: true,
          RouterView: true,
          Sidebar: true,
        },
      },
    });
    const authStore = useAuthStore();

    const result = openDialog(AboutDialog);
    await findByRole("dialog", { name: "Enduro" });

    expect(document.body.classList.contains("modal-open")).toBe(true);
    expect(document.body.style.overflow).toBe("hidden");
    expect(document.querySelector(".modal-backdrop")).not.toBeNull();

    authStore.$patch({ user: null });
    await nextTick();

    await expect(result).resolves.toBeUndefined();
    expect(queryByRole("dialog", { name: "Enduro" })).toBeNull();
    expect(document.body.classList.contains("modal-open")).toBe(false);
    expect(document.body.style.overflow).toBe("");
    expect(document.querySelector(".modal-backdrop")).toBeNull();

    authStore.user = createValidUser();
    await nextTick();

    expect(queryByRole("dialog", { name: "Enduro" })).toBeNull();
  });
});
