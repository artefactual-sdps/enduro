import UUID from "@/components/UUID.vue";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, describe, expect, it } from "vitest";

describe("UUID.vue", () => {
  const uuid = "31ceb5d5-a9c1-488b-b4ee-40910e54109e";

  afterEach(() => cleanup());

  it("should render", () => {
    const { getByText } = render(UUID, { props: { id: uuid } });

    expect(getByText(uuid).className).toBe("font-monospace");
  });

  it("should copy to clipboard", async () => {
    let clipboardText = "";
    Object.assign(navigator, {
      clipboard: { writeText: (text: string) => (clipboardText = text) },
    });

    const { getByRole } = render(UUID, { props: { id: uuid } });

    await fireEvent.click(getByRole("button", { name: "Copy to clipboard" }));

    expect(clipboardText).toEqual(uuid);
    getByRole("button", { name: "Copied!" });

    setTimeout(() => {
      getByRole("button", { name: "Copy to clipboard" });
    }, 500);
  });
});
