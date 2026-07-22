import { createTestingPinia } from "@pinia/testing";
import { mount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";

import LocationDialog from "@/components/LocationDialog.vue";
import { useLocationStore } from "@/stores/location";

const showMock = vi.hoisted(() => vi.fn());
const hideMock = vi.hoisted(() => vi.fn());
const disposeMock = vi.hoisted(() => vi.fn());

vi.mock("bootstrap/js/dist/modal", () => ({
  default: class {
    show = showMock;
    hide = hideMock;
    dispose = disposeMock;
  },
}));

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

  it("resolves the selected location after closing", async () => {
    const wrapper = mount(LocationDialog, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              location: {
                locations: [{ name: "Location 1", uuid: "location-1" }],
              },
            },
          }),
        ],
      },
    });

    await wrapper.get("button.btn-primary").trigger("click");
    expect(hideMock).toHaveBeenCalledOnce();

    wrapper
      .get('[role="dialog"]')
      .element.dispatchEvent(new Event("hidden.bs.modal"));

    expect(wrapper.emitted("resolve")).toEqual([["location-1"]]);
  });
});
