import { createTestingPinia } from "@pinia/testing";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, describe, expect, it, vi } from "vitest";

import AboutDialog from "@/components/AboutDialog.vue";
import { EnduroChildworkflowTypeEnum } from "@/openapi-generator";
import { useAboutStore } from "@/stores/about";

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

function renderDialog(initialState: Record<string, unknown> = {}) {
  return render(AboutDialog, {
    global: {
      plugins: [
        createTestingPinia({
          createSpy: vi.fn,
          initialState,
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

  it("loads application information and resolves when hidden", async () => {
    const { emitted, getByRole } = renderDialog();

    expect(showMock).toHaveBeenCalledOnce();
    expect(useAboutStore().load).toHaveBeenCalledOnce();

    await fireEvent(
      getByRole("dialog", { name: "Enduro" }),
      new Event("hidden.bs.modal"),
    );

    expect(emitted().resolve).toEqual([[]]);
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
