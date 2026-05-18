import { createTestingPinia } from "@pinia/testing";
import { shallowMount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createMemoryHistory, createRouter } from "vue-router";

import Page from "@/pages/storage/locations/[id].vue";
import { useLocationStore } from "@/stores/location";

describe("storage/locations/[id].vue", () => {
  afterEach(() => vi.clearAllMocks());

  it("resets the location store on unmount", async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        {
          name: "/storage/locations/[id]/",
          path: "/storage/locations/:id/",
          component: {},
        },
        {
          name: "/storage/locations/[id]/aips",
          path: "/storage/locations/:id/aips",
          component: {},
        },
      ],
    });
    await router.push("/storage/locations/location-uuid/");
    const pinia = createTestingPinia({ createSpy: vi.fn });
    vi.mocked(useLocationStore(pinia).fetchCurrent).mockResolvedValue();
    const wrapper = shallowMount(Page, {
      global: { plugins: [pinia, router] },
    });

    const reset = vi.spyOn(useLocationStore(), "$reset");
    wrapper.unmount();
    expect(reset).toHaveBeenCalledOnce();
  });
});
