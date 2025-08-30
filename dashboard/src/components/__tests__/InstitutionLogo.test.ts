import { cleanup } from "@testing-library/vue";
import { VueWrapper, mount } from "@vue/test-utils";
import { afterEach, describe, expect, it } from "vitest";

import InstitutionLogo from "@/components/InstitutionLogo.vue";

describe("InstitutionLogo.vue", () => {
  afterEach(() => cleanup());

  it("renders a logo with link and alt text", async () => {
    const wrapper: VueWrapper = mount(InstitutionLogo, {
      attachTo: document.body,
      props: {
        logo: "http://localhost:8080/artefactual-logo.png",
        name: "Artefactual Systems Inc.",
        url: "http://localhost:8080",
      },
    });

    const logo = wrapper.get("#institution-logo");

    expect(wrapper.get("a").attributes("href")).toEqual(
      "http://localhost:8080",
    );
    expect(logo.attributes("alt")).toEqual("Artefactual Systems Inc.");
    expect(logo.attributes("src")).toEqual(
      "http://localhost:8080/artefactual-logo.png",
    );
  });

  it("renders a logo with alt text", async () => {
    const wrapper: VueWrapper = mount(InstitutionLogo, {
      attachTo: document.body,
      props: {
        logo: "http://localhost:8080/artefactual-logo.png",
        name: "Artefactual Systems Inc.",
        url: "",
      },
    });

    const logo = wrapper.get("#institution-logo");

    expect(wrapper.find("a").exists()).toBe(false);
    expect(logo.attributes("alt")).toEqual("Artefactual Systems Inc.");
    expect(logo.attributes("src")).toEqual(
      "http://localhost:8080/artefactual-logo.png",
    );
  });

  it("renders nothing if logo and url are empty", async () => {
    const wrapper: VueWrapper = mount(InstitutionLogo, {
      attachTo: document.body,
      props: {
        logo: "",
        name: "Artefactual Systems Inc.",
        url: "",
      },
    });

    expect(wrapper.html()).toBe("<!--v-if-->");
  });
});
