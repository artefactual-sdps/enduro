import { cleanup, render } from "@testing-library/vue";
import { RouterLinkStub } from "@vue/test-utils";
import { afterEach, describe, expect, it } from "vitest";

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
});
