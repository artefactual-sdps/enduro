import { cleanup, render } from "@testing-library/vue";
import { afterEach, describe, expect, it } from "vitest";
import { defineComponent, h } from "vue";

import useBootstrapModal from "@/composables/useBootstrapModal";

const modalHarness = defineComponent({
  setup() {
    const { element } = useBootstrapModal(() => {});

    return () =>
      h("div", { ref: element, class: "modal" }, [
        h("div", { class: "modal-dialog" }),
      ]);
  },
});

describe("useBootstrapModal", () => {
  afterEach(() => cleanup());

  it("restores document state when unmounted while visible", () => {
    const { unmount } = render(modalHarness);

    expect(document.body.classList.contains("modal-open")).toBe(true);
    expect(document.body.style.overflow).toBe("hidden");
    expect(document.querySelector(".modal-backdrop")).not.toBeNull();

    unmount();

    expect(document.body.classList.contains("modal-open")).toBe(false);
    expect(document.body.style.overflow).toBe("");
    expect(document.querySelector(".modal-backdrop")).toBeNull();
  });
});
