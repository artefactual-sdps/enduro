import { VueWrapper, mount } from "@vue/test-utils";
import { beforeEach, describe, expect, it } from "vitest";

import Task from "./Task.vue";

describe("Task.vue", () => {
  let wrapper: VueWrapper;

  beforeEach(() => {
    wrapper = mount(Task, {
      attachTo: document.body,
      props: {
        index: 1,
        task: {
          uuid: "task-uuid",
          name: "Task 1",
          startedAt: new Date("2020-02-25T17:21:03Z"),
          completedAt: new Date("2020-02-25T17:22:38Z"),
          status: "done",
          note: "This is a note\nwith multiple lines",
          workflowUuid: "workflow-uuid",
        },
      },
    });
  });

  it("renders the component", async () => {
    expect(wrapper.exists()).toBe(true);
  });

  it("applies the intended task alignment utilities", () => {
    expect(wrapper.get(".card-body > .d-flex").classes()).toContain(
      "align-items-start",
    );
    expect(wrapper.get(".card-body > .d-flex > .d-flex").exists()).toBe(true);
  });

  it("shows the time completed if the task is done", async () => {
    const time = wrapper.find("#pt-task-uuid-time span");

    expect(process.env.TZ).toEqual("America/Regina");
    expect(time.text()).toEqual("Completed: 2020-02-25 11:22:38");
  });

  it("shows the time started if the task is in progress", async () => {
    wrapper = mount(Task, {
      props: {
        index: 1,
        task: {
          uuid: "task-uuid",
          name: "Task 1",
          startedAt: new Date("2020-02-25T17:21:03Z"),
          status: "in progress",
          workflowUuid: "workflow-uuid",
        },
      },
    });

    const time = wrapper.find("#pt-task-uuid-time span");

    expect(process.env.TZ).toEqual("America/Regina");
    expect(time.text()).toEqual("Started: 2020-02-25 11:21:03");
  });

  it("shows the first line of the note by default", async () => {
    const note = wrapper.find("#pt-task-uuid-note");
    const more = wrapper.find("#pt-task-uuid-note-more");

    expect(note.text()).toEqual("This is a note");
    expect(more.isVisible()).toBe(false);
  });

  it("shows all lines of the note after expanding the card", async () => {
    const note = wrapper.find("#pt-task-uuid-note");
    const more = wrapper.find("#pt-task-uuid-note-more");
    const toggle = wrapper.find("#pt-task-uuid-note-toggle");

    expect(toggle.element.tagName).toBe("BUTTON");
    expect(toggle.attributes("aria-expanded")).toBe("false");
    await toggle.trigger("click");

    expect(note.text()).toEqual("This is a note");
    expect(more.isVisible()).toBe(true);
    expect(more.text()).toEqual("with multiple lines");
    expect(toggle.attributes("aria-expanded")).toBe("true");
  });

  it("doesn't have an expand control when the note is only one line", async () => {
    wrapper = mount(Task, {
      props: {
        index: 1,
        task: {
          uuid: "task-uuid",
          name: "Task 1",
          startedAt: new Date("2020-02-25T17:21:03Z"),
          completedAt: new Date("2020-02-25T17:22:38Z"),
          status: "done",
          note: "This note is only one line",
          workflowUuid: "workflow-uuid",
        },
      },
    });

    const note = wrapper.find("#pt-task-uuid-note");
    const more = wrapper.find("#pt-task-uuid-note-more");
    const toggle = wrapper.find("#pt-task-uuid-note-toggle");

    expect(note.text()).toEqual("This note is only one line");
    expect(more.exists()).toBe(false);
    expect(toggle.exists()).toEqual(false);
  });
});
