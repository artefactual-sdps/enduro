import { createTestingPinia } from "@pinia/testing";
import { shallowMount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createMemoryHistory, createRouter } from "vue-router";

import Page from "@/pages/ingest/sips/[id].vue";
import { useSipStore } from "@/stores/sip";

describe("ingest/sips/[id].vue", () => {
  afterEach(() => vi.clearAllMocks());

  it("resets the SIP store on unmount", async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        {
          name: "/ingest/sips/[id]/",
          path: "/ingest/sips/:id/",
          component: {},
        },
      ],
    });
    await router.push("/ingest/sips/sip-uuid/");
    const pinia = createTestingPinia({ createSpy: vi.fn });
    vi.mocked(useSipStore(pinia).fetchCurrent).mockResolvedValue(undefined);
    const wrapper = shallowMount(Page, {
      global: { plugins: [pinia, router] },
    });

    const reset = vi.spyOn(useSipStore(), "$reset");
    wrapper.unmount();
    expect(reset).toHaveBeenCalledOnce();
  });
});
