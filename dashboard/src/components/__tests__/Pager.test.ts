import { mount } from "@vue/test-utils";
import { describe, expect, it } from "vitest";

import Pager from "../Pager.vue";

describe("Pager", () => {
  it("renders the correct number of pages when totalPages is less than maxVisiblePages", () => {
    const wrapper = mount(Pager, {
      props: {
        currentPage: 1,
        totalPages: 5,
        maxVisiblePages: 8,
      },
    });

    const pageItems = wrapper.findAll(".page-item");
    expect(pageItems.length).toBe(7); // 5 pages + prev + next.
    expect(wrapper.findAll(".page-item:not(.disabled)").length).toBe(6); // Active pages
  });

  it("renders the correct range of pages when totalPages is greater than maxVisiblePages", () => {
    const wrapper = mount(Pager, {
      props: {
        currentPage: 5,
        totalPages: 20,
        maxVisiblePages: 5,
      },
    });

    const pageItems = wrapper.findAll(".page-item");
    expect(pageItems.length).toBeGreaterThan(5); // Includes navigation buttons
    expect(wrapper.findAll(".page-item.active").length).toBe(1);
  });

  it("disables the Previous buttons on the first page", () => {
    const wrapper = mount(Pager, {
      props: {
        currentPage: 1,
        totalPages: 10,
      },
    });

    const prevButton = wrapper.find("li.page-item:first-child");

    expect(prevButton.classes()).toContain("disabled");
  });

  it("disables the Next buttons on the last page", () => {
    const wrapper = mount(Pager, {
      props: {
        currentPage: 10,
        totalPages: 10,
      },
    });

    const nextButton = wrapper.find("li.page-item:last-child");

    expect(nextButton.classes()).toContain("disabled");
  });

  it("emits 'page-change' event when a page is clicked", async () => {
    const wrapper = mount(Pager, {
      props: {
        currentPage: 1,
        totalPages: 10,
      },
    });

    await wrapper.find("#page-3").trigger("click");

    expect(wrapper.emitted("page-change")![0]).toEqual([3]);
  });

  it("does not emit 'page-change' event when clicking on the current page", async () => {
    const wrapper = mount(Pager, {
      props: {
        currentPage: 3,
        totalPages: 10,
      },
    });

    const currentPage = wrapper.find(".page-item.active");
    await currentPage.find("a").trigger("click");

    expect(wrapper.emitted("page-change")).toBeFalsy();
  });

  it("emits 'page-change' event when navigating to the next page", async () => {
    const wrapper = mount(Pager, {
      props: {
        currentPage: 3,
        totalPages: 10,
      },
    });

    await wrapper.find("#next-page").trigger("click");

    expect(wrapper.emitted("page-change")).toBeTruthy();
    expect(wrapper.emitted("page-change")![0]).toEqual([4]);
  });

  it("emits 'page-change' event when navigating to the previous page", async () => {
    const wrapper = mount(Pager, {
      props: {
        currentPage: 3,
        totalPages: 10,
      },
    });

    await wrapper.find("#prev-page").trigger("click");

    expect(wrapper.emitted("page-change")).toBeTruthy();
    expect(wrapper.emitted("page-change")![0]).toEqual([2]);
  });

  it("emits 'page-change' event when navigating to the first page", async () => {
    const wrapper = mount(Pager, {
      props: {
        currentPage: 5,
        totalPages: 10,
      },
    });

    await wrapper.find("#first-page").trigger("click");

    expect(wrapper.emitted("page-change")).toBeTruthy();
    expect(wrapper.emitted("page-change")![0]).toEqual([1]);
  });

  it("emits 'page-change' event when navigating to the last page", async () => {
    const wrapper = mount(Pager, {
      props: {
        currentPage: 5,
        totalPages: 10,
      },
    });

    await wrapper.find("#last-page").trigger("click");

    expect(wrapper.emitted("page-change")).toBeTruthy();
    expect(wrapper.emitted("page-change")![0]).toEqual([10]);
  });
});
