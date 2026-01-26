import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, describe, expect, it, vi } from "vitest";

import BatchReviewConfirmDialog from "@/components/BatchReviewConfirmDialog.vue";

const closeDialogMock = vi.hoisted(() => vi.fn());
const showMock = vi.hoisted(() => vi.fn());
const hideMock = vi.hoisted(() => vi.fn());

vi.mock("vue3-promise-dialog", () => ({ closeDialog: closeDialogMock }));
vi.mock("bootstrap/js/dist/modal", () => {
  return {
    default: class ModalMock {
      show = showMock;
      hide = hideMock;
    },
  };
});

describe("BatchReviewConfirmDialog.vue", () => {
  afterEach(() => {
    cleanup();
    vi.resetAllMocks();
  });

  it("renders the dialog and confirms", async () => {
    const { getByRole, getByText } = render(BatchReviewConfirmDialog, {
      props: {
        heading: "Cancel batch",
        bodyHtml: "<p>Are you sure you want to cancel?</p>",
        confirmClass: "btn-danger",
      },
    });

    expect(showMock).toHaveBeenCalled();
    getByText("Cancel batch");
    getByText("Are you sure you want to cancel?");

    const yesButton = getByRole("button", { name: "Yes" });
    expect(yesButton.className).toContain("btn-danger");
    const noButton = getByRole("button", { name: "No" });
    expect(noButton.className).toContain("btn-secondary");

    await fireEvent.click(yesButton);
    expect(hideMock).toHaveBeenCalled();

    const modalEl = getByRole("dialog", { name: "Cancel batch" });
    await fireEvent(modalEl, new Event("hidden.bs.modal"));

    expect(closeDialogMock).toHaveBeenCalledWith(true);
  });

  it("closes without confirmation", async () => {
    const { getByRole, getByText } = render(BatchReviewConfirmDialog, {
      props: {
        heading: "Continue partial ingest",
        bodyHtml: "<p>Are you sure you want to continue processing?</p>",
        confirmClass: "btn-primary",
      },
    });

    expect(showMock).toHaveBeenCalled();
    getByText("Continue partial ingest");
    getByText("Are you sure you want to continue processing?");

    const yesButton = getByRole("button", { name: "Yes" });
    expect(yesButton.className).toContain("btn-primary");
    const noButton = getByRole("button", { name: "No" });
    expect(noButton.className).toContain("btn-secondary");

    await fireEvent.click(noButton);
    expect(hideMock).toHaveBeenCalled();

    const modalEl = getByRole("dialog", { name: "Continue partial ingest" });
    await fireEvent(modalEl, new Event("hidden.bs.modal"));

    expect(closeDialogMock).toHaveBeenCalledWith(false);
  });
});
