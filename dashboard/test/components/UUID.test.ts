import UUID from "../../src/components/UUID.vue";
import { render } from "@testing-library/vue";
import { RouterLinkStub } from "@vue/test-utils";
import { describe, expect, it } from "vitest";

describe("UUID.vue", () => {
  it("should render", () => {
    const { getByText, unmount } = render(UUID, {
      props: {
        id: "31ceb5d5-a9c1-488b-b4ee-40910e54109e",
      },
    });

    const el = getByText("31ceb5d5-a9c1-488b-b4ee-40910e54109e");
    expect(el.className).toBe("font-monospace");

    unmount();
  });
});
