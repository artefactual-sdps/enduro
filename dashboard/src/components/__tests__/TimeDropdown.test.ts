import { flushPromises, mount } from "@vue/test-utils";
import VueDatePicker from "@vuepic/vue-datepicker";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import TimeDropdown from "../TimeDropdown.vue";

describe("TimeDropdown.vue", () => {
  let wrapper: ReturnType<typeof mount>;

  beforeEach(() => {
    vi.useFakeTimers();
    wrapper = mount(TimeDropdown, {
      attachTo: document.body,
      props: {
        name: "createdAt",
        label: "Started",
      },
    });
  });

  afterEach(() => {
    vi.useRealTimers();
    wrapper.unmount();
  });

  it("initializes with correct default values", async () => {
    expect(wrapper.find(".dropdown-toggle").text()).toBe("Started");
  });

  it.each([
    ["3h", "2025-01-01T09:00:00Z"], // Last 3 hours.
    ["6h", "2025-01-01T06:00:00Z"], // Last 6 hours.
    ["12h", "2025-01-01T00:00:00Z"], // Last 12 hours.
    ["24h", "2024-12-31T12:00:00Z"], // Last 24 hours.
    ["3d", "2024-12-29T12:00:00Z"], // Last 3 days.
    ["7d", "2024-12-25T12:00:00Z"], // Last 7 days.
  ])("emits the correct event when a date is selected", async (value, want) => {
    // set the test time to noon on 2025-01-01 (UTC).
    const date = new Date(Date.UTC(2025, 0, 1, 12, 0, 0));
    vi.setSystemTime(date);

    await wrapper.find("select").setValue(value);

    expect(wrapper.emitted("change")).toEqual([["createdAt", want, ""]]);
  });

  it.each([
    ["3h", "Started: 3h"],
    ["6h", "Started: 6h"],
    ["12h", "Started: 12h"],
    ["24h", "Started: 24h"],
    ["3d", "Started: 3d"],
    ["7d", "Started: 7d"],
  ])(
    "sets the button label correctly when a time is selected",
    async (value, want) => {
      const button = wrapper.find("button");

      await wrapper.find("select").setValue(value);

      expect(button.text()).toBe(want);
    },
  );

  it("clears all values when the clear button is clicked", async () => {
    const button = wrapper.find("button");
    const clearButton = wrapper.find("button[type='reset']");

    await clearButton.trigger("click");

    expect(button.text()).toBe("Started");
    expect(wrapper.emitted("change")).toEqual([["createdAt", "", ""]]);
  });

  it("emits the correct event when a custom date is selected", async () => {
    const button = wrapper.find("button");
    const datePicker = wrapper.findComponent<typeof VueDatePicker>(
      '[data-test="startTime"]',
    );

    datePicker.vm.$emit(
      "update:model-value",
      new Date(Date.UTC(2025, 0, 1, 12, 0, 0)),
    );
    await datePicker.vm.$nextTick();

    expect(button.text()).toBe("Started: Custom");
    expect(wrapper.emitted("change")).toEqual([
      ["createdAt", "2025-01-01T12:00:00Z", ""],
    ]);
  });
});
