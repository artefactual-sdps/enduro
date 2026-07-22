import { createTestingPinia } from "@pinia/testing";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, describe, expect, it, vi } from "vitest";

import AipDeletionRequestDialog from "@/components/AipDeletionRequestDialog.vue";

const showMock = vi.hoisted(() => vi.fn());
const hideMock = vi.hoisted(() => vi.fn());
const disposeMock = vi.hoisted(() => vi.fn());

vi.mock("bootstrap/js/dist/modal", () => ({
  default: class {
    show = showMock;
    hide = hideMock;
    dispose = disposeMock;
  },
}));

const renderDialog = () =>
  render(AipDeletionRequestDialog, {
    global: {
      plugins: [
        createTestingPinia({
          createSpy: vi.fn,
          initialState: {
            aip: { current: { name: "Test AIP" } },
          },
        }),
      ],
    },
  });

describe("AipDeletionRequestDialog.vue", () => {
  afterEach(() => {
    cleanup();
    vi.resetAllMocks();
  });

  it("resolves the deletion reason after submission", async () => {
    const { emitted, getByRole } = renderDialog();

    await fireEvent.update(
      getByRole("textbox", { name: "Reason:" }),
      "No longer required",
    );
    await fireEvent.click(getByRole("button", { name: "Request deletion" }));
    expect(hideMock).toHaveBeenCalledOnce();

    await fireEvent(
      getByRole("dialog", { name: "Delete AIP" }),
      new Event("hidden.bs.modal"),
    );

    expect(emitted().resolve).toEqual([["No longer required"]]);
  });

  it("resolves null when cancelled", async () => {
    const { emitted, getByRole } = renderDialog();

    await fireEvent(
      getByRole("dialog", { name: "Delete AIP" }),
      new Event("hidden.bs.modal"),
    );

    expect(emitted().resolve).toEqual([[null]]);
  });
});
