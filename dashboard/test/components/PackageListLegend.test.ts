import PackageListLegend from "../../src/components/PackageListLegend.vue";
import { fireEvent, render } from "@testing-library/vue";
import { describe, expect, it } from "vitest";

describe("PackageListLegend.vue", () => {
  it("renders when the package is moving", async () => {
    const {
      getByText,
      getByLabelText,
      queryByText,
      emitted,
      unmount,
      rerender,
    } = render(PackageListLegend, {
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

    unmount();
  });
});
