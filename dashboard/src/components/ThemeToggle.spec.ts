import { mount } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";

import IconMoon from "~icons/clarity/moon-line";
import IconSun from "~icons/clarity/sun-line";

import ThemeToggle from "@/components/ThemeToggle.vue";
import { setTheme } from "@/test/theme";

describe("ThemeToggle.vue", () => {
  beforeEach(async () => {
    localStorage.clear();
    await setTheme("light");
  });

  afterEach(() => {
    localStorage.clear();
  });

  it("offers dark mode with a moon icon while light mode is active", () => {
    const wrapper = mount(ThemeToggle);

    expect(wrapper.get("button").text()).toBe("Dark theme");
    expect(wrapper.findComponent(IconMoon).exists()).toBe(true);
    expect(wrapper.findComponent(IconSun).exists()).toBe(false);
  });

  it("offers light mode with a sun icon in compact mode", async () => {
    await setTheme("dark");
    const wrapper = mount(ThemeToggle, { props: { compact: true } });

    expect(wrapper.get("button").attributes("aria-label")).toBe("Light theme");
    expect(wrapper.get("button").text()).toBe("");
    expect(wrapper.findComponent(IconMoon).exists()).toBe(false);
    expect(wrapper.findComponent(IconSun).exists()).toBe(true);
  });

  it("toggles and persists the opposite theme", async () => {
    const wrapper = mount(ThemeToggle);

    await wrapper.get("button").trigger("click");

    expect(document.documentElement.dataset.bsTheme).toBe("dark");
    expect(localStorage.getItem("enduro-theme")).toBe("dark");
    expect(wrapper.get("button").text()).toBe("Light theme");
    expect(wrapper.findComponent(IconSun).exists()).toBe(true);
  });
});
