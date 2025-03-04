import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, describe, expect, it } from "vitest";

import SipListLegend from "@/components/SipListLegend.vue";

describe("SipListLegend.vue", () => {
  afterEach(() => cleanup());

  it("renders", async () => {
    const { getByText, getByLabelText, queryByText, emitted, rerender } =
      render(SipListLegend, {
        props: {
          modelValue: true,
        },
      });

    getByText("DONE");
    getByText("ERROR");
    getByText("IN PROGRESS");
    getByText("QUEUED");
    getByText("PENDING");

    // Closing emits an event.
    const button = getByLabelText("Close");
    await fireEvent.click(button);
    expect(emitted()["update:modelValue"][0]).toStrictEqual([false]);

    // And setting the prop to false should hide the legend.
    await rerender({ modelValue: false });
    expect(queryByText("DONE")).toBeNull();
  });
});
