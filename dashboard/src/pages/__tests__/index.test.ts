import { createTestingPinia } from "@pinia/testing";
import { flushPromises, mount } from "@vue/test-utils";
import { User } from "oidc-client-ts";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import SafeHtml from "@/components/SafeHtml.vue";
import IndexPage from "@/pages/index.vue";

function mountPage(initialState: Record<string, unknown> = {}) {
  return mount(IndexPage, {
    global: {
      components: { SafeHtml },
      plugins: [
        createTestingPinia({
          createSpy: vi.fn,
          initialState: {
            auth: {
              config: { enabled: false },
              user: null,
              ...(initialState.auth ?? {}),
            },
          },
        }),
      ],
    },
  });
}

describe("index.vue", () => {
  beforeEach(() => {
    vi.stubGlobal("fetch", vi.fn());
    vi.stubEnv("VITE_CUSTOM_HOME_URL", "");
  });

  afterEach(() => {
    vi.restoreAllMocks();
  });

  it("shows default content", async () => {
    const wrapper = mountPage();

    await flushPromises();
    expect(global.fetch).not.toHaveBeenCalled();
    expect(wrapper.find('[role="status"]').exists()).toBe(false);
    expect(wrapper.text()).toContain("Welcome!");
    expect(wrapper.text()).toContain("Enduro is a new application");

    wrapper.unmount();
  });

  it("shows default content with user name", async () => {
    const wrapper = mountPage({
      auth: {
        config: { enabled: true },
        user: new User({
          access_token: "",
          token_type: "",
          profile: {
            aud: "",
            exp: 0,
            iat: 0,
            iss: "",
            sub: "",
            name: "John Doe",
          },
        }),
      },
    });

    await flushPromises();
    expect(global.fetch).not.toHaveBeenCalled();
    expect(wrapper.text()).toContain("Welcome, John Doe!");
    expect(wrapper.text()).toContain("Enduro is a new application");

    wrapper.unmount();
  });

  it("shows loading spinner while fetching custom content", async () => {
    vi.stubEnv("VITE_CUSTOM_HOME_URL", "http://example.com/custom.html");
    global.fetch = vi.fn().mockReturnValue(new Promise(() => {}));

    const wrapper = mountPage();

    await flushPromises();
    expect(global.fetch).toHaveBeenCalledWith("http://example.com/custom.html");
    expect(wrapper.find('[role="status"]').exists()).toBe(true);
    expect(wrapper.text()).toContain("Loading...");
    expect(wrapper.text()).not.toContain("Welcome!");
    expect(wrapper.text()).not.toContain("Enduro is a new application");

    wrapper.unmount();
  });

  it("shows custom content", async () => {
    vi.stubEnv("VITE_CUSTOM_HOME_URL", "http://example.com/custom.html");
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      text: async () => "<h1>Custom Home Page</h1><p>Custom content</p>",
    });

    const wrapper = mountPage();

    await flushPromises();
    expect(global.fetch).toHaveBeenCalledWith("http://example.com/custom.html");
    expect(wrapper.text()).toContain("Custom Home Page");
    expect(wrapper.text()).toContain("Custom content");
    expect(wrapper.text()).not.toContain("Welcome!");
    expect(wrapper.text()).not.toContain("Enduro is a new application");

    wrapper.unmount();
  });

  it("sanitizes custom content", async () => {
    vi.stubEnv("VITE_CUSTOM_HOME_URL", "http://example.com/custom.html");
    global.fetch = vi.fn().mockResolvedValue({
      ok: true,
      text: async () =>
        "<h1>Title</h1><script>alert('XSS')</script><p>Content</p>",
    });

    const wrapper = mountPage();

    await flushPromises();
    expect(global.fetch).toHaveBeenCalledWith("http://example.com/custom.html");
    expect(wrapper.text()).toContain("Title");
    expect(wrapper.text()).toContain("Content");
    expect(wrapper.text()).not.toContain("alert");
    expect(wrapper.text()).not.toContain("Welcome!");
    expect(wrapper.text()).not.toContain("Enduro is a new application");

    wrapper.unmount();
  });

  it("shows error when custom content fails to load", async () => {
    vi.stubEnv("VITE_CUSTOM_HOME_URL", "http://example.com/custom.html");
    global.fetch = vi.fn().mockResolvedValue({ ok: false, status: 404 });

    const wrapper = mountPage();

    await flushPromises();
    expect(global.fetch).toHaveBeenCalledWith("http://example.com/custom.html");
    expect(wrapper.find('[role="alert"]').exists()).toBe(true);
    expect(wrapper.text()).toContain("Failed to load custom home content.");
    expect(wrapper.text()).toContain("Welcome!");
    expect(wrapper.text()).toContain("Enduro is a new application");

    wrapper.unmount();
  });

  it("shows error when custom content fetch throws an error", async () => {
    vi.stubEnv("VITE_CUSTOM_HOME_URL", "http://example.com/custom.html");
    global.fetch = vi.fn().mockRejectedValue(new Error());

    const wrapper = mountPage();

    await flushPromises();
    expect(global.fetch).toHaveBeenCalledWith("http://example.com/custom.html");
    expect(wrapper.find('[role="alert"]').exists()).toBe(true);
    expect(wrapper.text()).toContain("Failed to load custom home content.");
    expect(wrapper.text()).toContain("Welcome!");
    expect(wrapper.text()).toContain("Enduro is a new application");

    wrapper.unmount();
  });
});
