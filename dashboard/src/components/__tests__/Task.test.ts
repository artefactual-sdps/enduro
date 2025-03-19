import { VueWrapper, mount } from "@vue/test-utils";
import { afterEach, beforeEach, describe, expect, it } from "vitest";

import Task from "../Task.vue";

describe("Task.vue", () => {
  let wrapper: VueWrapper;

  beforeEach(() => {
    wrapper = mount(Task, {
      attachTo: document.body,
      props: {
        index: 1,
        task: {
          id: 1,
          name: "Task 1",
          startedAt: new Date("2020-02-25T17:21:03Z"),
          completedAt: new Date("2020-02-25T17:22:38Z"),
          status: "done",
          taskId: "9614f71f-c21e-4da9-a07e-1661bd010d1f",
          note: "This is a note\nwith multiple lines",
        },
      },
    });
  });

  afterEach(() => {
    wrapper.unmount();
  });

  it("renders the component", async () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("shows the time completed if the task is done", async () => {
    const time = wrapper.find("#pt-1-time span");

    expect(process.env.TZ).toEqual("America/Regina");
    expect(time.text()).toEqual("Completed: 2020-02-25 11:22:38");
  });

  it("shows the time started if the task is in progress", async () => {
    wrapper = mount(Task, {
      props: {
        index: 1,
        task: {
          id: 1,
          name: "Task 1",
          startedAt: new Date("2020-02-25T17:21:03Z"),
          status: "in progress",
          taskId: "9614f71f-c21e-4da9-a07e-1661bd010d1f",
        },
      },
    });

    const time = wrapper.find("#pt-1-time span");

    expect(process.env.TZ).toEqual("America/Regina");
    expect(time.text()).toEqual("Started: 2020-02-25 11:21:03");
  });

  it("shows the first line of the note by default", async () => {
    const note = wrapper.find("#pt-1-note");
    const more = wrapper.find("#pt-1-note-more");

    expect(note.text()).toEqual("This is a note");
    expect(more.isVisible()).toBe(false);
  });

  it("shows all lines of the note after expanding the card", async () => {
    const note = wrapper.find("#pt-1-note");
    const more = wrapper.find("#pt-1-note-more");
    const toggle = wrapper.find("#pt-1-note-toggle");

    await toggle.trigger("click");

    expect(note.text()).toEqual("This is a note");
    expect(more.isVisible()).toBe(true);
    expect(more.text()).toEqual("with multiple lines");
  });

  it("doesn't have an expand link when the note is only one line", async () => {
    wrapper = mount(Task, {
      props: {
        index: 1,
        task: {
          id: 1,
          name: "Task 1",
          startedAt: new Date("2020-02-25T17:21:03Z"),
          completedAt: new Date("2020-02-25T17:22:38Z"),
          status: "done",
          taskId: "9614f71f-c21e-4da9-a07e-1661bd010d1f",
          note: "This note is only one line",
        },
      },
    });

    const note = wrapper.find("#pt-1-note");
    const more = wrapper.find("#pt-1-note-more");
    const toggle = wrapper.find("#pt-1-note-toggle");

    expect(note.text()).toEqual("This note is only one line");
    expect(more.exists()).toBe(false);
    expect(toggle.exists()).toEqual(false);
  });
});
