import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, client } from "@/client";
import { useBatchStore } from "@/stores/batch";
import { useLayoutStore } from "@/stores/layout";

vi.mock("@/client");

describe("useBatchStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("fetches current batch", async () => {
    const store = useBatchStore();
    const mockBatch: api.EnduroIngestBatch = {
      uuid: "batch-uuid-1",
      identifier: "Batch 1",
      createdAt: new Date("2025-01-01T00:00:00Z"),
      status: api.EnduroIngestBatchStatusEnum.Ingested,
      sipsCount: 2,
    };
    const mockSips: api.EnduroIngestSips = {
      items: [
        {
          uuid: "sip-uuid-1",
          name: "SIP 1",
          createdAt: new Date("2025-01-01T00:00:00Z"),
          status: api.EnduroIngestSipStatusEnum.Ingested,
        },
        {
          uuid: "sip-uuid-2",
          name: "SIP 2",
          createdAt: new Date("2025-01-02T00:00:00Z"),
          status: api.EnduroIngestSipStatusEnum.Ingested,
        },
      ],
      page: { limit: 20, offset: 0, total: 2 },
    };

    client.ingest.ingestShowBatch = vi.fn().mockResolvedValue(mockBatch);
    client.ingest.ingestListSips = vi.fn().mockResolvedValue(mockSips);

    await store.fetchCurrent("batch-uuid-1");

    expect(store.current).toEqual(mockBatch);
    expect(store.currentSips).toEqual(mockSips.items);

    const layoutStore = useLayoutStore();
    expect(layoutStore.breadcrumb).toEqual([
      { text: "Ingest" },
      //{ route: expect.any(Object), text: "Batches" },
      { text: "Batches" },
      { text: mockBatch.identifier },
    ]);
  });

  it("fetches batches", async () => {
    const store = useBatchStore();
    const mockBatches: api.EnduroIngestBatches = {
      items: [
        {
          uuid: "batch-uuid-1",
          identifier: "Batch 1",
          createdAt: new Date("2025-01-01T00:00:00Z"),
          status: api.EnduroIngestBatchStatusEnum.Ingested,
          sipsCount: 2,
        },
        {
          uuid: "batch-uuid-2",
          identifier: "Batch 2",
          createdAt: new Date("2025-01-02T00:00:00Z"),
          status: api.EnduroIngestBatchStatusEnum.Processing,
          sipsCount: 1,
        },
      ],
      page: { limit: 20, offset: 0, total: 2 },
    };

    client.ingest.ingestListBatches = vi.fn().mockResolvedValue(mockBatches);

    await store.fetchBatches(1);

    expect(store.batches).toEqual(mockBatches.items);
    expect(store.page).toEqual(mockBatches.page);
  });
});
