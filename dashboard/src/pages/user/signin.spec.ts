import { createTestingPinia } from "@pinia/testing";
import { flushPromises, mount } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import SigninPage from "@/pages/user/signin.vue";
import { useAuthStore } from "@/stores/auth";
import { setTheme } from "@/test/theme";

function mountPage() {
  const pinia = createTestingPinia({ createSpy: vi.fn });
  const wrapper = mount(SigninPage, {
    attachTo: document.body,
    global: {
      plugins: [pinia],
    },
  });

  return { authStore: useAuthStore(), wrapper };
}

describe("signin.vue", () => {
  beforeEach(async () => {
    vi.clearAllMocks();
    await setTheme("light");
  });

  afterEach(() => {
    localStorage.clear();
    vi.unstubAllGlobals();
  });

  it("presents the OIDC sign-in action accessibly", () => {
    const { wrapper } = mountPage();
    const button = wrapper.get(".signin-button");

    expect(wrapper.get("h1").text()).toBe("Welcome to Enduro");
    expect(button.text()).toContain("Sign in with your organization");
    expect(button.attributes("type")).toBe("button");
    expect(button.attributes("aria-describedby")).toContain(
      "signin-description",
    );
    expect(wrapper.get(".signin-visual").attributes("aria-hidden")).toBe(
      "true",
    );
    expect(wrapper.get(".signin-footer img").attributes("alt")).toBe("");
    expect(wrapper.get(".signin-footer nav").attributes("aria-label")).toBe(
      "Enduro and Artefactual resources",
    );

    const links = wrapper.findAll(".signin-footer-link");
    expect(links).toHaveLength(2);
    expect(links[0].text()).toContain("Documentation");
    expect(links[0].attributes()).toMatchObject({
      "aria-label": "Read the Enduro documentation (opens in a new tab)",
      href: "https://enduro.readthedocs.io/",
      rel: "noopener noreferrer",
      target: "_blank",
    });
    expect(links[1].text()).toContain("Artefactual");
    expect(links[1].attributes()).toMatchObject({
      "aria-label": "Visit the Artefactual website (opens in a new tab)",
      href: "https://www.artefactual.com/",
      rel: "noopener noreferrer",
      target: "_blank",
    });

    const themeToggle = wrapper.get(".signin-theme-toggle");
    expect(themeToggle.attributes("aria-label")).toBe("Dark theme");
    expect(themeToggle.text()).toBe("");
    expect(themeToggle.element.parentElement).toBe(
      wrapper.get(".signin-panel").element,
    );
  });

  it("smoothly tracks the pointer across the document", () => {
    let animationFrame: FrameRequestCallback | undefined;
    vi.stubGlobal(
      "requestAnimationFrame",
      vi.fn((callback: FrameRequestCallback) => {
        animationFrame = callback;
        return 1;
      }),
    );
    vi.stubGlobal("cancelAnimationFrame", vi.fn());

    const { wrapper } = mountPage();
    const visual = wrapper.get(".signin-visual");

    window.dispatchEvent(
      new PointerEvent("pointermove", {
        clientX: window.innerWidth,
        clientY: 0,
        pointerType: "mouse",
      }),
    );
    animationFrame?.(0);

    const style = (visual.element as HTMLElement).style;
    expect(
      Number.parseFloat(style.getPropertyValue("--signin-watermark-x")),
    ).toBeCloseTo(-2.4);
    expect(
      Number.parseFloat(style.getPropertyValue("--signin-ambient-y")),
    ).toBeCloseTo(-2.2);
    expect(style.getPropertyValue("--signin-gradient-x")).toBe("65%");
  });

  it("does not track the pointer when reduced motion is preferred", () => {
    vi.stubGlobal(
      "matchMedia",
      vi.fn().mockReturnValue({
        matches: true,
        media: "(prefers-reduced-motion: reduce)",
        addEventListener: vi.fn(),
        removeEventListener: vi.fn(),
      } as unknown as MediaQueryList),
    );
    vi.stubGlobal("requestAnimationFrame", vi.fn());

    mountPage();
    window.dispatchEvent(
      new PointerEvent("pointermove", {
        clientX: window.innerWidth,
        clientY: 0,
        pointerType: "mouse",
      }),
    );

    expect(window.requestAnimationFrame).not.toHaveBeenCalled();
  });

  it("prevents repeated sign-in requests while redirecting", async () => {
    const { authStore, wrapper } = mountPage();
    const button = wrapper.get(".signin-button");

    await button.trigger("click");
    await button.trigger("click");

    expect(authStore.signinRedirect).toHaveBeenCalledOnce();
    expect(button.attributes("disabled")).toBeDefined();
    expect(button.text()).toContain("Redirecting to sign in");
    expect(wrapper.get('[role="status"]').text()).toContain(
      "Redirecting to your identity provider",
    );
  });

  it("shows an alert when the OIDC redirect cannot start", async () => {
    const { authStore, wrapper } = mountPage();
    vi.mocked(authStore.signinRedirect).mockRejectedValue(
      new Error("OIDC discovery failed"),
    );

    await wrapper.get(".signin-button").trigger("click");
    await flushPromises();

    expect(wrapper.get('[role="alert"]').text()).toContain(
      "We couldn't connect to the sign-in service",
    );
    expect(
      wrapper.get(".signin-button").attributes("disabled"),
    ).toBeUndefined();
    expect(document.activeElement).toBe(wrapper.get(".signin-button").element);
  });
});
