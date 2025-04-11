import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";

import ResultCounter from "../ResultCounter.vue";

describe("ResultCounter", () => {
  it("renders correctly with zero results", () => {
    const wrapper = mount(ResultCounter, {
      props: {
        total: 0,
      },
    });
    expect(wrapper.text()).toContain("No results found");
  });

  it("renders correctly with one result", () => {
    const wrapper = mount(ResultCounter, {
      props: {
        total: 1,
      },
    });
    expect(wrapper.text()).toContain("Found 1 result");
  });

  it("renders correctly with one page of results", () => {
    const wrapper = mount(ResultCounter, {
      props: {
        total: 20,
      },
    });
    expect(wrapper.text()).toContain("Found 20 results");
  });

  it("renders correctly more than one page of results", () => {
    const wrapper = mount(ResultCounter, {
      props: {
        total: 35,
      },
    });
    expect(wrapper.text()).toContain("Showing 1 - 20 of 35 results");
  });

  it("renders correctly on a page showing fewer results than the limit", () => {
    const wrapper = mount(ResultCounter, {
      props: {
        offset: 20,
        total: 35,
      },
    });
    expect(wrapper.text()).toContain("Showing 21 - 35 of 35 results");
  });
});
