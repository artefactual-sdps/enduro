import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import UUID from "@/components/UUID.vue";

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
    // TODO: figure out how to grant permissions via `navigator.permissions` so
    // useClipboard can use `navigator.clipboard.writeText` instead. Mock
    // `document.execCommand` since that's still a supported fallback.
    document.execCommand = () => true;
    const exec = vi
      .spyOn(document, "execCommand")
      .mockImplementation((command) => {
        return command === "copy";
      });

    const { getByRole, findByRole } = render(UUID, { props: { id: uuid } });

    await fireEvent.click(getByRole("button", { name: "Copy to clipboard" }));

    expect(exec).toHaveBeenCalledWith("copy");
    getByRole("button", { name: "Copied!" });

    // Confirm that the button goes back into its original state.
    vi.runOnlyPendingTimers();
    await findByRole("button", { name: "Copy to clipboard" });
  });
});
