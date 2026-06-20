// @vitest-environment jsdom
// DOMPurify and happy-dom don't work well together, so we use jsdom for these
// tests to ensure the sanitization logic is properly tested.
import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";

import SafeHtml from "@/components/SafeHtml.vue";

describe("SafeHtml", () => {
  it("renders sanitized HTML", () => {
    const wrapper = mount(SafeHtml, {
      props: {
        html: "<p>Hello <strong>world</strong></p>",
      },
    });

    expect(wrapper.html()).toMatchInlineSnapshot(`
      "<div>
        <p>Hello <strong>world</strong></p>
      </div>"
    `);
  });

  it("removes unsafe content", () => {
    const wrapper = mount(SafeHtml, {
      props: {
        html: `<div>Test<script>alert('xss')</script></div>`,
      },
    });

    expect(wrapper.html()).toMatchInlineSnapshot(`
      "<div>
        <div>Test</div>
      </div>"
    `);
  });

  it("renders nothing when html is empty", () => {
    const wrapper = mount(SafeHtml, {
      props: {
        html: "",
      },
    });

    expect(wrapper.html()).toMatchInlineSnapshot(`"<!--v-if-->"`);
  });
});
