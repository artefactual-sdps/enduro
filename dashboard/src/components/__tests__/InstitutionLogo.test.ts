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

    const logo = wrapper.get("img");

    expect(wrapper.get("a").attributes("href")).toEqual(
      "http://localhost:8080",
    );
    expect(logo.attributes("alt")).toEqual("Artefactual Systems Inc.");
    expect(logo.attributes("src")).toEqual(
      "http://localhost:8080/artefactual-logo.png",
    );
    expect(wrapper.html()).toMatchInlineSnapshot(
      `"<div data-v-5841bc2f="" class="d-none d-sm-block mx-3"><a data-v-5841bc2f="" href="http://localhost:8080" target="_blank" rel="external"><img data-v-5841bc2f="" src="http://localhost:8080/artefactual-logo.png" alt="Artefactual Systems Inc."></a></div>"`,
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

    const logo = wrapper.get("img");

    expect(wrapper.find("a").exists()).toBe(false);
    expect(logo.attributes("alt")).toEqual("Artefactual Systems Inc.");
    expect(logo.attributes("src")).toEqual(
      "http://localhost:8080/artefactual-logo.png",
    );
    expect(wrapper.html()).toMatchInlineSnapshot(
      `"<div data-v-5841bc2f="" class="d-none d-sm-block mx-3"><img data-v-5841bc2f="" src="http://localhost:8080/artefactual-logo.png" alt="Artefactual Systems Inc."></div>"`,
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
