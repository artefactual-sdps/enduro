import { createTestingPinia } from "@pinia/testing";
import { flushPromises, mount } from "@vue/test-utils";
import { User } from "oidc-client-ts";
import { describe, expect, it, vi } from "vitest";

import IndexPage from "@/pages/index.vue";

describe("index.vue", () => {
  it("shows default content", async () => {
    const wrapper = mount(IndexPage, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              auth: { config: { enabled: false } },
              custom: { loaded: true, manifest: null },
            },
          }),
        ],
      },
    });

    await flushPromises();
    expect(wrapper.find('[role="status"]').exists()).toBe(false);
    expect(wrapper.text()).toContain("Welcome!");
    expect(wrapper.text()).toContain("Enduro is a new application");

    wrapper.unmount();
  });

  it("shows default content with user name", async () => {
    const wrapper = mount(IndexPage, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
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
              custom: { loaded: true, manifest: null },
            },
          }),
        ],
      },
    });

    await flushPromises();
    expect(wrapper.text()).toContain("Welcome, John Doe!");
    expect(wrapper.text()).toContain("Enduro is a new application");

    wrapper.unmount();
  });

  it("shows loading spinner while fetching custom content", async () => {
    const wrapper = mount(IndexPage, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              auth: { config: { enabled: false } },
              custom: {
                loaded: true,
                manifest: { homeUrl: "/custom/home.html" },
                homeLoading: true,
                homeContent: null,
                homeError: null,
              },
            },
          }),
        ],
      },
    });

    await flushPromises();
    expect(wrapper.find('[role="status"]').exists()).toBe(true);
    expect(wrapper.text()).toContain("Loading...");
    expect(wrapper.text()).not.toContain("Welcome!");
    expect(wrapper.text()).not.toContain("Enduro is a new application");

    wrapper.unmount();
  });

  it("shows custom content", async () => {
    const wrapper = mount(IndexPage, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              auth: { config: { enabled: false } },
              custom: {
                loaded: true,
                manifest: { homeUrl: "/custom/home.html" },
                homeLoading: false,
                homeContent: "<h1>Custom Home Page</h1><p>Custom content</p>",
                homeError: null,
              },
            },
          }),
        ],
      },
    });

    await flushPromises();
    expect(wrapper.text()).toContain("Custom Home Page");
    expect(wrapper.text()).toContain("Custom content");
    expect(wrapper.text()).not.toContain("Welcome!");
    expect(wrapper.text()).not.toContain("Enduro is a new application");

    wrapper.unmount();
  });

  it("shows error and default content when custom content fails to load", async () => {
    const wrapper = mount(IndexPage, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              auth: { config: { enabled: false } },
              custom: {
                loaded: true,
                manifest: { homeUrl: "/custom/home.html" },
                homeLoading: false,
                homeContent: null,
                homeError: "Failed to load custom home content.",
              },
            },
          }),
        ],
      },
    });

    await flushPromises();
    expect(wrapper.find('[role="alert"]').exists()).toBe(true);
    expect(wrapper.text()).toContain("Failed to load custom home content.");
    expect(wrapper.text()).toContain("Welcome!");
    expect(wrapper.text()).toContain("Enduro is a new application");

    wrapper.unmount();
  });
});
