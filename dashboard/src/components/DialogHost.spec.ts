import { fireEvent, render } from "@testing-library/vue";
import { describe, expect, it } from "vitest";
import { defineComponent, h } from "vue";

import DialogHost from "@/components/DialogHost.vue";
import { DialogAlreadyOpenError, defineDialog } from "@/dialogs/dialog";

interface TestDialogProps {
  message: string;
}

const testDialogComponent = defineComponent({
  props: {
    message: { type: String, required: true },
  },
  emits: {
    resolve: (value: string) => typeof value === "string",
  },
  setup(props, { emit }) {
    return () =>
      h(
        "button",
        { onClick: () => emit("resolve", props.message) },
        props.message,
      );
  },
});

const openTestDialog = defineDialog<string, TestDialogProps>(
  testDialogComponent,
  "Dialog cancelled",
);

describe("DialogHost.vue", () => {
  it("renders a dialog and resolves its result", async () => {
    const { findByRole } = render(DialogHost);

    const result = openTestDialog({ message: "Dialog result" });
    await fireEvent.click(
      await findByRole("button", { name: "Dialog result" }),
    );

    await expect(result).resolves.toBe("Dialog result");
  });

  it("rejects a second dialog while one is active", async () => {
    const { findByRole } = render(DialogHost);

    const result = openTestDialog({ message: "First dialog" });
    await expect(
      openTestDialog({ message: "Second dialog" }),
    ).rejects.toBeInstanceOf(DialogAlreadyOpenError);

    await fireEvent.click(await findByRole("button", { name: "First dialog" }));
    await expect(result).resolves.toBe("First dialog");
  });

  it("resolves with cancellation when the host is unmounted", async () => {
    const { findByRole, unmount } = render(DialogHost);

    const result = openTestDialog({ message: "Active dialog" });
    await findByRole("button", { name: "Active dialog" });

    unmount();

    await expect(result).resolves.toBe("Dialog cancelled");
  });
});
