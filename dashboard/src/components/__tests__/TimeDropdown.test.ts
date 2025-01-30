import { mount } from "@vue/test-utils";
import { ref } from "vue";
import TimeDropdown from "../TimeDropdown.vue";
import { afterEach, beforeEach, describe, it, expect, vi } from "vitest";

const changeEvent = ref<{ field: string; value: string } | null>(null);

const handleChange = (field: string, value: string) => {
  changeEvent.value = { field, value };
};

describe("TimeDropdown.vue", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    vi.useRealTimers();
  });

  it("renders correctly", () => {
    const wrapper = mount(TimeDropdown, {
      props: {
        fieldname: "testField",
        onChange: handleChange,
      },
    });
    expect(wrapper.exists()).toBe(true);
  });

  it("initializes with correct default values", () => {
    const wrapper = mount(TimeDropdown, {
      props: {
        fieldname: "testField",
        onChange: handleChange,
      },
    });
    const button = wrapper.find("button");
    expect(button.text()).toBe("Started");
  });

  it.each([
    [0, ""], // Any time (default).
    [1, "2025-01-01T09:00:00Z"], // Last 3 hours.
    [2, "2025-01-01T06:00:00Z"], // Last 6 hours.
    [3, "2025-01-01T00:00:00Z"], // Last 12 hours.
    [4, "2024-12-31T12:00:00Z"], // Last 24 hours.
    [5, "2024-12-29T12:00:00Z"], // Last 3 days.
    [6, "2024-12-25T12:00:00Z"], // Last 7 days.
  ])("emits the correct event when a date is selected", async (index, want) => {
    // set the test time to noon on 2025-01-01 (UTC).
    const date = new Date(Date.UTC(2025, 0, 1, 12, 0, 0));
    vi.setSystemTime(date);

    const wrapper = mount(TimeDropdown, {
      props: {
        fieldname: "testField",
        onChange: handleChange,
      },
    });

    const options = wrapper.findAll(".dropdown-item");
    await options[index].trigger("click");
    expect(changeEvent.value).toEqual({
      field: "testField",
      value: want,
    });
  });

  it.each([
    [0, "Started"], // Any time.
    [1, "Started: The last 3 hours"],
    [2, "Started: The last 6 hours"],
    [3, "Started: The last 12 hours"],
    [4, "Started: The last 24 hours"],
    [5, "Started: The last 3 days"],
    [6, "Started: The last 7 days"],
  ])(
    "sets the button label correctly when a time is selected",
    async (index, want) => {
      const wrapper = mount(TimeDropdown, {
        props: {
          fieldname: "testField",
          onChange: handleChange,
        },
      });

      const button = wrapper.find("button");
      const option = wrapper.findAll(".dropdown-item");

      await option[index].trigger("click");
      expect(button.text()).toBe(want);
    },
  );

  it("resets the button label when 'Any time' is selected", async () => {
    const wrapper = mount(TimeDropdown, {
      props: {
        fieldname: "testField",
        onChange: handleChange,
      },
    });

    const button = wrapper.find("button");
    const option = wrapper.findAll(".dropdown-item");

    await option[1].trigger("click");
    expect(button.text()).toBe("Started: The last 3 hours");

    await option[0].trigger("click");
    expect(button.text()).toBe("Started");
  });
});
