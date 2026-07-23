import { createTestingPinia } from "@pinia/testing";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, describe, expect, it, vi } from "vitest";

import AboutDialog from "@/components/AboutDialog.vue";
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

describe("AboutDialog.vue", () => {
  afterEach(() => {
    cleanup();
    vi.resetAllMocks();
  });

  it("loads application information and resolves when hidden", async () => {
    const { emitted, getByRole } = render(AboutDialog, {
      global: {
        plugins: [createTestingPinia({ createSpy: vi.fn })],
      },
    });

    expect(showMock).toHaveBeenCalledOnce();
    expect(useAboutStore().load).toHaveBeenCalledOnce();

    await fireEvent(
      getByRole("dialog", { name: "Enduro" }),
      new Event("hidden.bs.modal"),
    );

    expect(emitted().resolve).toEqual([[undefined]]);
  });
});
