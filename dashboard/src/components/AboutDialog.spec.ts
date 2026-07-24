import { createTestingPinia } from "@pinia/testing";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, describe, expect, it, vi } from "vitest";

import AboutDialog from "@/components/AboutDialog.vue";
import { EnduroChildworkflowTypeEnum } from "@/openapi-generator";
import { useAuthStore } from "@/stores/auth";

const closeDialogMock = vi.hoisted(() => vi.fn());
const showMock = vi.hoisted(() => vi.fn());

vi.mock("vue3-promise-dialog", () => ({ closeDialog: closeDialogMock }));
vi.mock("bootstrap/js/dist/modal", () => {
  return {
    default: class ModalMock {
      show = showMock;
    },
  };
});

function renderDialog(initialState: Record<string, unknown> = {}) {
  return render(AboutDialog, {
    global: {
      plugins: [
        createTestingPinia({
          createSpy: vi.fn,
          initialState: {
            auth: {
              config: { enabled: true },
              user: { expired: false },
            },
            ...initialState,
          },
        }),
      ],
    },
  });
}

describe("AboutDialog.vue", () => {
  afterEach(() => {
    cleanup();
    vi.resetAllMocks();
  });

  it("shows the modal on mount", () => {
    renderDialog();

    expect(showMock).toHaveBeenCalledOnce();
  });

  it("closes the dialog when the Bootstrap modal is hidden", async () => {
    const { getByRole } = renderDialog();

    const modalEl = getByRole("dialog", { name: "Enduro" });
    await fireEvent(modalEl, new Event("hidden.bs.modal"));

    expect(closeDialogMock).toHaveBeenCalledWith(null);
  });

  it("closes the dialog immediately when the user session expires", async () => {
    renderDialog();

    const authStore = useAuthStore();
    authStore.$patch({ user: null });

    await vi.waitFor(() => {
      expect(closeDialogMock).toHaveBeenCalledWith(null);
    });
  });

  it("does not show child workflows when none are configured", () => {
    const { queryByText } = renderDialog();

    expect(queryByText("Child workflows:")).toBeNull();
  });

  it("shows child workflows when configured", () => {
    const { getByText } = renderDialog({
      about: {
        childWorkflows: [
          {
            type: EnduroChildworkflowTypeEnum.Preprocessing,
            workflowName: "my-preprocessing-workflow",
          },
          {
            type: EnduroChildworkflowTypeEnum.Poststorage,
            workflowName: "my-poststorage-workflow",
          },
        ],
      },
    });

    getByText("Child workflows:");
    getByText("my-preprocessing-workflow");
    getByText("my-poststorage-workflow");
  });
});
