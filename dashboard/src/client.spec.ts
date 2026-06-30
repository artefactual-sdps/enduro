import { createPinia, setActivePinia } from "pinia";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { handleUnauthorized } from "@/client";
import router from "@/router";
import { useAuthStore } from "@/stores/auth";

describe("handleUnauthorized", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("removes the user and navigates to sign in", async () => {
    const authStore = useAuthStore();
    authStore.user = { access_token: "expired" } as typeof authStore.user;
    const pushSpy = vi.spyOn(router, "push").mockResolvedValue(undefined);

    await handleUnauthorized();

    expect(authStore.user).toBeNull();
    expect(pushSpy).toHaveBeenCalledWith({ name: "/user/signin" });
  });
});
