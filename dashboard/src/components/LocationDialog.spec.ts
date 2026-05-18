import { createTestingPinia } from "@pinia/testing";
import { mount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";

import LocationDialog from "@/components/LocationDialog.vue";
import { useLocationStore } from "@/stores/location";

vi.mock("bootstrap/js/dist/modal", () => ({
  default: class {
    show() {}
  },
}));

vi.mock("vue3-promise-dialog", () => ({ closeDialog: vi.fn() }));

describe("LocationDialog.vue", () => {
  afterEach(() => vi.clearAllMocks());

  it("resets the location store on unmount", () => {
    const wrapper = mount(LocationDialog, {
      global: { plugins: [createTestingPinia({ createSpy: vi.fn })] },
    });

    const reset = vi.spyOn(useLocationStore(), "$reset");
    wrapper.unmount();
    expect(reset).toHaveBeenCalledOnce();
  });
});
