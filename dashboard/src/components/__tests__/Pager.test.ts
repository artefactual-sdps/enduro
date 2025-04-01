import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";

import Pager from "../Pager.vue";

describe("Pager", () => {
  it("renders the correct number of pages when there are less than maxVisiblePages", () => {
    const wrapper = mount(Pager, {
      props: {
        total: 80,
        maxVisiblePages: 5,
      },
    });

    // "< 1 2 3 4 >"
    const pageLinks = wrapper.findAll(".page-item a");
    expect(pageLinks.length).toBe(6);
    expect(pageLinks[0].attributes("title")).toBe("Previous page");
    expect(pageLinks[1].text()).toBe("1");
    expect(pageLinks[4].text()).toBe("4");
    expect(pageLinks[5].attributes("title")).toBe("Next page");
  });

  it("renders the correct range of pages when there are more than maxVisiblePages", () => {
    const wrapper = mount(Pager, {
      props: {
        offset: 100,
        total: 200,
        maxVisiblePages: 5,
      },
    });

    // "< 1 … 4 5 [6] 7 8 … 10 >"
    const pageLinks = wrapper.findAll(".page-item a");
    expect(pageLinks.length).toBe(11);
    expect(pageLinks[0].attributes("title")).toBe("Previous page");
    expect(pageLinks[1].text()).toBe("1");
    expect(pageLinks[2].text()).toBe("…");
    expect(pageLinks[3].text()).toBe("4");
    expect(wrapper.find(".page-item.active").text()).toBe("6");
    expect(pageLinks[7].text()).toBe("8");
    expect(pageLinks[8].text()).toBe("…");
    expect(pageLinks[9].text()).toBe("10");
    expect(pageLinks[10].attributes("title")).toBe("Next page");
  });

  it("disables the Previous button on the first page", () => {
    const wrapper = mount(Pager, { props: { total: 100 } });

    const prevButton = wrapper.find("li.page-item:first-child");

    expect(prevButton.classes()).toContain("disabled");
  });

  it("disables the Next button on the last page", () => {
    const wrapper = mount(Pager, {
      props: {
        offset: 80,
        total: 100,
      },
    });

    const nextButton = wrapper.find("li.page-item:last-child");

    expect(nextButton.classes()).toContain("disabled");
  });

  it("emits 'page-change' event when a page is clicked", async () => {
    const wrapper = mount(Pager, { props: { total: 100 } });

    await wrapper.find("#page-3").trigger("click");

    expect(wrapper.emitted("page-change")![0]).toEqual([3]);
  });

  it("does not emit 'page-change' event when clicking on the current page", async () => {
    const wrapper = mount(Pager, {
      props: {
        offset: 20,
        total: 100,
      },
    });

    const currentPage = wrapper.find(".page-item.active");
    await currentPage.find("a").trigger("click");

    expect(wrapper.emitted("page-change")).toBeFalsy();
  });

  it("emits 'page-change' event when navigating to the next page", async () => {
    const wrapper = mount(Pager, { props: { total: 100 } });

    await wrapper.find("#next-page").trigger("click");

    expect(wrapper.emitted("page-change")).toBeTruthy();
    expect(wrapper.emitted("page-change")![0]).toEqual([2]);
  });

  it("emits 'page-change' event when navigating to the previous page", async () => {
    const wrapper = mount(Pager, {
      props: {
        offset: 80,
        total: 100,
      },
    });

    await wrapper.find("#prev-page").trigger("click");

    expect(wrapper.emitted("page-change")).toBeTruthy();
    expect(wrapper.emitted("page-change")![0]).toEqual([4]);
  });

  it("emits 'page-change' event when navigating to the first page", async () => {
    const wrapper = mount(Pager, {
      props: {
        offset: 80,
        total: 100,
      },
    });

    await wrapper.find("#page-1").trigger("click");

    expect(wrapper.emitted("page-change")).toBeTruthy();
    expect(wrapper.emitted("page-change")![0]).toEqual([1]);
  });

  it("emits 'page-change' event when navigating to the last page", async () => {
    const wrapper = mount(Pager, { props: { total: 100 } });

    await wrapper.find("#page-5").trigger("click");

    expect(wrapper.emitted("page-change")).toBeTruthy();
    expect(wrapper.emitted("page-change")![0]).toEqual([5]);
  });
});
