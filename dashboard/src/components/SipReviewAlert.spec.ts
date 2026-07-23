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
      global: { plugins: [createTestingPinia({ createSpy: vi.fn })] },
    });

    const reset = vi.spyOn(useAipStore(), "$reset");
    wrapper.unmount();
    expect(reset).toHaveBeenCalledOnce();
  });

  it("does not offer the obsolete task expansion action", () => {
    const wrapper = mount(SipReviewAlert, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: { sip: { current: { status: "pending" } } },
          }),
        ],
      },
    });

    expect(wrapper.text()).toContain("Task: Review AIP");
    expect(wrapper.text()).not.toContain("Expand");
  });
});
