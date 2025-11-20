import { VueWrapper, mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";

import EmailLinkedText from "../EmailLinkedText.vue";

describe("EmailLinkedText.vue", () => {
  let wrapper: VueWrapper;

  it("renders simple text without emails", () => {
    wrapper = mount(EmailLinkedText, {
      props: {
        text: "Hello world",
      },
    });
    expect(wrapper.find("a").exists()).toBe(false);
    expect(wrapper.html()).toMatchInlineSnapshot(`"<span>Hello world</span>"`);
  });

  it("links a single email address", () => {
    wrapper = mount(EmailLinkedText, {
      props: {
        text: "Contact us at support@example.com",
      },
    });
    const link = wrapper.find("a");
    expect(link.exists()).toBe(true);
    expect(link.attributes("href")).toBe("mailto:support@example.com");
    expect(link.text()).toBe("support@example.com");
    expect(wrapper.html()).toMatchInlineSnapshot(
      `"<span>Contact us at <a href="mailto:support@example.com">support@example.com</a></span>"`,
    );
  });

  it("links multiple email addresses", () => {
    wrapper = mount(EmailLinkedText, {
      props: {
        text: "Email alice@example.com or bob@example.com for help.",
      },
    });
    const links = wrapper.findAll("a");
    expect(links).toHaveLength(2);
    expect(links[0].attributes("href")).toBe("mailto:alice@example.com");
    expect(links[0].text()).toBe("alice@example.com");
    expect(links[1].attributes("href")).toBe("mailto:bob@example.com");
    expect(links[1].text()).toBe("bob@example.com");
    expect(wrapper.html()).toMatchInlineSnapshot(
      `"<span>Email <a href="mailto:alice@example.com">alice@example.com</a> or <a href="mailto:bob@example.com">bob@example.com</a> for help.</span>"`,
    );
  });

  it("handles text starting with an email", () => {
    wrapper = mount(EmailLinkedText, {
      props: {
        text: "info@example.com is our email.",
      },
    });
    const link = wrapper.find("a");
    expect(link.exists()).toBe(true);
    expect(link.attributes("href")).toBe("mailto:info@example.com");
    expect(wrapper.html()).toMatchInlineSnapshot(
      `"<span><a href="mailto:info@example.com">info@example.com</a> is our email.</span>"`,
    );
  });

  it("handles text ending with an email", () => {
    wrapper = mount(EmailLinkedText, {
      props: {
        text: "Write to me@example.com",
      },
    });
    const link = wrapper.find("a");
    expect(link.exists()).toBe(true);
    expect(link.attributes("href")).toBe("mailto:me@example.com");
    expect(wrapper.html()).toMatchInlineSnapshot(
      `"<span>Write to <a href="mailto:me@example.com">me@example.com</a></span>"`,
    );
  });
});
