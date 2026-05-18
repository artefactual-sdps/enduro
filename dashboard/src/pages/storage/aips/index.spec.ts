import { createTestingPinia } from "@pinia/testing";
import { shallowMount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createMemoryHistory, createRouter } from "vue-router";

import Page from "@/pages/storage/aips/index.vue";
import { useAipStore } from "@/stores/aip";

vi.mock("bootstrap/js/dist/tooltip", () => ({ default: class {} }));

describe("storage/aips/index.vue", () => {
  afterEach(() => vi.clearAllMocks());

  it("resets the AIP store on unmount", async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { name: "/storage/aips/", path: "/storage/aips/", component: {} },
      ],
    });
    await router.push("/storage/aips/");
    const wrapper = shallowMount(Page, {
      global: { plugins: [createTestingPinia({ createSpy: vi.fn }), router] },
    });

    const reset = vi.spyOn(useAipStore(), "$reset");
    wrapper.unmount();
    expect(reset).toHaveBeenCalledOnce();
  });
});
