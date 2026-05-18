import { createTestingPinia } from "@pinia/testing";
import { shallowMount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createMemoryHistory, createRouter } from "vue-router";

import Page from "@/pages/storage/aips/[id].vue";
import { useAipStore } from "@/stores/aip";

describe("storage/aips/[id].vue", () => {
  afterEach(() => vi.clearAllMocks());

  it("resets the AIP store on unmount", async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        {
          name: "/storage/aips/[id]/",
          path: "/storage/aips/:id/",
          component: {},
        },
      ],
    });
    await router.push("/storage/aips/aip-uuid/");
    const pinia = createTestingPinia({ createSpy: vi.fn });
    vi.mocked(useAipStore(pinia).fetchCurrent).mockResolvedValue(undefined);
    const wrapper = shallowMount(Page, {
      global: { plugins: [pinia, router] },
    });

    const reset = vi.spyOn(useAipStore(), "$reset");
    wrapper.unmount();
    expect(reset).toHaveBeenCalledOnce();
  });
});
