import { cleanup } from "@testing-library/vue";
import { mount } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";

import InstitutionLogo from "@/components/InstitutionLogo.vue";
import { useTheme } from "@/composables/useTheme";
import { setTheme } from "@/test/theme";

function stubInstitution({
  dark = "",
  legacy = "",
  light = "",
  name = "Example institution",
  url = "https://example.com",
}: {
  dark?: string;
  legacy?: string;
  light?: string;
  name?: string;
  url?: string;
} = {}) {
  vi.stubEnv("VITE_INSTITUTION_LOGO", legacy);
  vi.stubEnv("VITE_INSTITUTION_LOGO_LIGHT", light);
  vi.stubEnv("VITE_INSTITUTION_LOGO_DARK", dark);
  vi.stubEnv("VITE_INSTITUTION_NAME", name);
  vi.stubEnv("VITE_INSTITUTION_URL", url);
}

describe("InstitutionLogo.vue", () => {
  beforeEach(async () => {
    await setTheme("light");
  });

  afterEach(() => {
    cleanup();
    localStorage.clear();
    vi.unstubAllEnvs();
  });

  it("switches the configured logo with the active theme", async () => {
    stubInstitution({
      dark: "https://example.com/logo-dark.png",
      legacy: "https://example.com/logo.png",
      light: "https://example.com/logo-light.png",
    });
    const wrapper = mount(InstitutionLogo, { attachTo: document.body });

    const logo = wrapper.get("img");

    expect(wrapper.get("a").attributes("href")).toBe("https://example.com");
    expect(logo.attributes("alt")).toBe("Example institution");
    expect(logo.attributes("src")).toBe("https://example.com/logo-light.png");

    useTheme().toggle();
    await nextTick();

    expect(logo.attributes("src")).toBe("https://example.com/logo-dark.png");
  });

  it("falls back to the legacy institution logo", () => {
    stubInstitution({ legacy: "https://example.com/logo.png" });
    const wrapper = mount(InstitutionLogo);

    expect(wrapper.get("img").attributes("src")).toBe(
      "https://example.com/logo.png",
    );
  });

  it("does not use the light logo as the dark logo fallback", async () => {
    await setTheme("dark");
    stubInstitution({
      light: "https://example.com/logo-light.png",
    });
    const wrapper = mount(InstitutionLogo);

    expect(wrapper.find("img").exists()).toBe(false);
  });

  it("renders an unlinked logo when no institution URL is configured", () => {
    stubInstitution({
      light: "https://example.com/logo-light.png",
      url: "",
    });
    const wrapper = mount(InstitutionLogo);

    expect(wrapper.find("a").exists()).toBe(false);
    expect(wrapper.get("img").attributes()).toMatchObject({
      alt: "Example institution",
      src: "https://example.com/logo-light.png",
    });
  });

  it("renders nothing when no logo is configured for the active theme", () => {
    stubInstitution();
    const wrapper = mount(InstitutionLogo);

    expect(wrapper.find("img").exists()).toBe(false);
    expect(wrapper.html()).toBe("<!--v-if-->");
  });
});
