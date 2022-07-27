import UUID from "@/components/UUID.vue";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

describe("UUID.vue", () => {
  beforeEach(() => {
    vi.useFakeTimers();
  });

  afterEach(() => {
    cleanup();
    vi.useRealTimers();
  });

  const uuid = "31ceb5d5-a9c1-488b-b4ee-40910e54109e";

  it("should render", () => {
    const { getByText } = render(UUID, { props: { id: uuid } });

    expect(getByText(uuid).className).toBe("font-monospace");
  });

  it("should copy to clipboard", async () => {
    let clipboardText = "";
    Object.assign(navigator, {
      clipboard: { writeText: (text: string) => (clipboardText = text) },
    });

    const { getByRole, findByRole } = render(UUID, { props: { id: uuid } });

    await fireEvent.click(getByRole("button", { name: "Copy to clipboard" }));

    expect(clipboardText).toEqual(uuid);
    getByRole("button", { name: "Copied!" });

    // Confirm that the button goes back into its original state.
    vi.runOnlyPendingTimers();
    await findByRole("button", { name: "Copy to clipboard" });
  });
});
