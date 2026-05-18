import { createTestingPinia } from "@pinia/testing";
import { shallowMount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";

import Page from "@/pages/storage/locations/index.vue";
import { useLocationStore } from "@/stores/location";

describe("storage/locations/index.vue", () => {
  afterEach(() => vi.clearAllMocks());

  it("resets the location store on unmount", () => {
    const wrapper = shallowMount(Page, {
      global: {
        plugins: [createTestingPinia({ createSpy: vi.fn })],
        stubs: { RouterLink: true },
      },
    });

    const reset = vi.spyOn(useLocationStore(), "$reset");
    wrapper.unmount();
    expect(reset).toHaveBeenCalledOnce();
  });
});
