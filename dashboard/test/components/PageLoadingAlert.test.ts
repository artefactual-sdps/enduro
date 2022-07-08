import PageLoadingAlert from "../../src/components/PageLoadingAlert.vue";
import { mount, RouterLinkStub } from "@vue/test-utils";
import { describe, expect, it } from "vitest";

describe("PageLoadingAlert.vue", () => {
  it("should render", () => {
    expect(PageLoadingAlert).toBeTruthy();

    const wrapper = mount(PageLoadingAlert, {
      props: {
        error: { response: { status: 404 } },
      },
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    });

    expect(wrapper.html()).toContain("Page not found!");
  });
});
