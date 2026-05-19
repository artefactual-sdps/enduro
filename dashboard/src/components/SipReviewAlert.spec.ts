import { createTestingPinia } from "@pinia/testing";
import { mount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";

import SipReviewAlert from "@/components/SipReviewAlert.vue";
import { useAipStore } from "@/stores/aip";

vi.mock("vue3-promise-dialog", () => ({ openDialog: vi.fn() }));

describe("SipReviewAlert.vue", () => {
  afterEach(() => vi.clearAllMocks());

  it("resets the AIP store on unmount", () => {
    const wrapper = mount(SipReviewAlert, {
      props: { expandCounter: 0 },
      global: { plugins: [createTestingPinia({ createSpy: vi.fn })] },
    });

    const reset = vi.spyOn(useAipStore(), "$reset");
    wrapper.unmount();
    expect(reset).toHaveBeenCalledOnce();
  });
});
