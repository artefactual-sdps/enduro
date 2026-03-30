import { cleanup, fireEvent, render } from "@testing-library/vue";
import { RouterLinkStub } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";

import PageLoadingAlert from "@/components/PageLoadingAlert.vue";

describe("PageLoadingAlert.vue", () => {
  afterEach(() => cleanup());

  it("should render", () => {
    const { html } = render(PageLoadingAlert, {
      props: {
        error: { response: { status: 404 } },
      },
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    });

    expect(html()).toMatchInlineSnapshot(`
      "<!-- Not found. -->
      <div class="alert alert-warning" role="alert">
        <h4 class="alert-heading">Page not found!</h4>
        <p>We can't find the page you're looking for.</p>
        <hr><a class="btn btn-warning"> Take me home </a>
      </div>
      <!-- Other errors. -->
      <!--v-if-->"
    `);
  });

  it("should render a retry action for non-404 errors and call execute", async () => {
    const execute = vi.fn();
    const { getByRole, getByText } = render(PageLoadingAlert, {
      props: {
        error: new Error("network failure"),
        execute,
      },
      global: {
        stubs: {
          RouterLink: RouterLinkStub,
        },
      },
    });

    getByText("It was not possible to load this page.");
    await fireEvent.click(getByRole("button", { name: "Retry" }));

    expect(execute).toHaveBeenCalledTimes(1);
  });
});
