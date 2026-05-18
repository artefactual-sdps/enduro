import { createTestingPinia } from "@pinia/testing";
import { shallowMount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createMemoryHistory, createRouter } from "vue-router";

import Page from "@/pages/ingest/sips/index.vue";
import { useSipStore } from "@/stores/sip";
import { useUserStore } from "@/stores/user";

vi.mock("bootstrap/js/dist/dropdown", () => ({ default: class {} }));
vi.mock("bootstrap/js/dist/tooltip", () => ({ default: class {} }));

describe("ingest/sips/index.vue", () => {
  afterEach(() => vi.clearAllMocks());

  it("resets the SIP and user stores on unmount", async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [{ name: "/ingest/sips/", path: "/ingest/sips/", component: {} }],
    });
    await router.push("/ingest/sips/");
    const wrapper = shallowMount(Page, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              auth: { config: { enabled: true, abac: { enabled: true } } },
            },
          }),
          router,
        ],
      },
    });

    const resetSip = vi.spyOn(useSipStore(), "$reset");
    const resetUser = vi.spyOn(useUserStore(), "$reset");
    wrapper.unmount();
    expect(resetSip).toHaveBeenCalledOnce();
    expect(resetUser).toHaveBeenCalledOnce();
  });
});
