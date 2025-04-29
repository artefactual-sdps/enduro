import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";

import WorkflowHelp from "../WorkflowHelp.vue";

describe("WorkflowHelp.vue", () => {
  it("renders the component correctly", () => {
    const wrapper = mount(WorkflowHelp);
    expect(wrapper.exists()).toBe(true);
  });

  it("displays help text when show is true", () => {
    const wrapper = mount(WorkflowHelp, {
      props: { show: true },
      attachTo: document.body,
    });

    expect(wrapper.get("#workflow-help").isVisible()).toBe(true);
    expect(wrapper.get("#workflow-description").text()).toContain(
      "A workflow is composed of one or more tasks performed on a SIP/AIP to support preservation.",
    );
  });

  it("hides help text when show is false", () => {
    const wrapper = mount(WorkflowHelp, {
      props: { show: false },
      attachTo: document.body,
    });

    expect(wrapper.get("#workflow-help").isVisible()).toBe(false);
  });

  it("emits update:show when close button is clicked", async () => {
    const wrapper = mount(WorkflowHelp, {
      props: { show: true },
      attachTo: document.body,
    });

    await wrapper.get("#workflow-help-close").trigger("click");
    expect(wrapper.emitted()["update:show"][0]).toEqual([false]);
  });
});
