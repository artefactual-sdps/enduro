import UUID from "@/components/UUID.vue";
import { render, cleanup } from "@testing-library/vue";
import { afterEach, describe, expect, it } from "vitest";

describe("UUID.vue", () => {
  afterEach(() => cleanup());

  it("should render", () => {
    const { getByText } = render(UUID, {
      props: {
        id: "31ceb5d5-a9c1-488b-b4ee-40910e54109e",
      },
    });

    const el = getByText("31ceb5d5-a9c1-488b-b4ee-40910e54109e");
    expect(el.className).toBe("font-monospace");
  });
});
