import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, describe, expect, it } from "vitest";

import { api } from "@/client";
import StatusLegend from "@/components/StatusLegend.vue";
import type { LegendItem } from "@/components/StatusLegend.vue";

describe("StatusLegend.vue", () => {
  afterEach(() => cleanup());

  it("renders", async () => {
    const { getByText, getByLabelText, queryByText, emitted, rerender } =
      render(StatusLegend, {
        props: {
          show: true,
          items: <LegendItem[]>[
            {
              status: api.EnduroIngestSipStatusEnum.Ingested,
              description: "Ingested description",
            },
            {
              status: api.EnduroIngestSipStatusEnum.Error,
              description: "Error description",
            },
          ],
        },
      });

    getByText("INGESTED");
    getByText("Ingested description");
    getByText("ERROR");
    getByText("Error description");

    // Closing emits an event.
    const button = getByLabelText("Close");
    await fireEvent.click(button);
    expect(emitted()["update:show"][0]).toStrictEqual([false]);

    // And setting the prop to false should hide the legend.
    await rerender({ show: false });
    expect(queryByText("INGESTED")).toBeNull();
  });
});
