import { mount } from "@vue/test-utils";
import VueDatePicker from "@vuepic/vue-datepicker";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";

import TimeDropdown from "../TimeDropdown.vue";

describe("TimeDropdown.vue", () => {
  let wrapper: ReturnType<typeof mount>;

  beforeEach(async () => {
    // set the test time to noon on 2025-01-01 (UTC).
    vi.useFakeTimers();
    const date = new Date(Date.UTC(2025, 0, 1, 12, 0, 0));

    // setSystemTime must be called before mounting the component or *bad*
    // things happen.
    vi.setSystemTime(date);

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
    await wrapper.find("#tdd-createdAt-preset").setValue(value);

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

      await wrapper.find("#tdd-createdAt-preset").setValue(value);

      expect(button.text()).toBe(want);
    },
  );

  it("clears all values when the reset button is clicked", async () => {
    const toggleBtn = wrapper.find("button");
    const preset = wrapper.find("#tdd-createdAt-preset");
    const reset = wrapper.find("#tdd-createdAt-reset");

    expect(reset.isVisible()).toBe(false);

    await preset.setValue("3h");

    expect(toggleBtn.text()).toBe("Started: 3h");
    expect(reset.isVisible()).toBe(true);

    await reset.trigger("click");

    expect(toggleBtn.text()).toBe("Started");
    expect(preset.element.getAttribute("value")).toBe(null);

    const changeEmitted = wrapper.emitted("change");
    if (changeEmitted === undefined) {
      throw new Error("change event was not emitted");
    } else {
      expect(changeEmitted.length).toBe(2);
      expect(changeEmitted[1]).toEqual(["createdAt", "", ""]);
    }
  });

  it("emits the correct event when a custom start date is selected", async () => {
    const button = wrapper.find("button");
    const datePicker = wrapper.findComponent<typeof VueDatePicker>(
      "#tdd-createdAt-start",
    );

    datePicker.vm.$emit(
      "update:model-value",
      new Date(Date.UTC(2025, 0, 1, 12, 0, 0)),
    );
    await nextTick();

    expect(button.text()).toBe("Started: Custom");
    expect(wrapper.emitted("change")).toEqual([
      ["createdAt", "2025-01-01T12:00:00Z", ""],
    ]);
  });

  it("emits the correct event when a custom end date is selected", async () => {
    const button = wrapper.find("button");
    const datePicker =
      wrapper.findComponent<typeof VueDatePicker>("#tdd-createdAt-end");

    datePicker.vm.$emit(
      "update:model-value",
      new Date(Date.UTC(2025, 0, 1, 12, 0, 0)),
    );
    await nextTick();

    expect(button.text()).toBe("Started: Custom");
    expect(wrapper.emitted("change")).toEqual([
      ["createdAt", "", "2025-01-01T12:00:00Z"],
    ]);
  });
});

describe("TimeDropdown.vue initialized with start and end times", () => {
  it("initializes with correct default values", async () => {
    const wrapper = mount(TimeDropdown, {
      props: {
        name: "createdAt",
        label: "Started",
        start: new Date("2025-01-01T00:00:00Z"),
        end: new Date("2025-01-31T23:59:59Z"),
      },
    });
    await nextTick();

    expect(wrapper.find(".dropdown-toggle").text()).toBe("Started: Custom");

    const start = wrapper.find("#tdd-createdAt-start input");
    const end = wrapper.find("#tdd-createdAt-end input");

    // Local times are offset -6 hours from UTC times.
    expect(start.element.getAttribute("value")).toEqual("12/31/2024, 18:00");
    expect(end.element.getAttribute("value")).toEqual("01/31/2025, 17:59");
  });
});
