import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, client } from "@/client";
import { useAipStore } from "@/stores/aip";
import { useLayoutStore } from "@/stores/layout";

vi.mock("@/client");

describe("useAipStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("isMovable", () => {
    const aipStore = useAipStore();
    const now = new Date();

    expect(aipStore.isMovable).toEqual(false);

    aipStore.$patch({
      current: {
        createdAt: now,
        uuid: "uuid-1",
        status: api.EnduroStorageAipStatusEnum.Stored,
      },
    });
    expect(aipStore.isMovable).toEqual(true);
  });

  it("isMoving", () => {
    const aipStore = useAipStore();

    expect(aipStore.isMoving).toEqual(false);

    aipStore.$patch({ locationChanging: true });
    expect(aipStore.isMoving).toEqual(true);
  });

  it("isStored", () => {
    const aipStore = useAipStore();
    const now = new Date();

    expect(aipStore.isStored).toEqual(false);

    aipStore.$patch({
      current: {
        createdAt: now,
        uuid: "uuid-1",
        status: api.EnduroStorageAipStatusEnum.Stored,
      },
    });
    expect(aipStore.isStored).toEqual(true);
  });

  it("fetches current", async () => {
    const mockAip: api.AIPResponse = {
      name: "AIP 1",
      createdAt: new Date("2025-01-01T00:00:00Z"),
      objectKey: "object-key-1",
      status: api.AIPResponseStatusEnum.Stored,
      uuid: "aip-uuid-1",
    };

    client.storage.storageShowAip = vi.fn().mockResolvedValue(mockAip);

    const store = useAipStore();
    await store.fetchCurrent("uuid-1234");

    expect(store.current).toEqual(mockAip);
    expect(store.locationChanging).toEqual(false);

    const layoutStore = useLayoutStore();
    expect(layoutStore.breadcrumb).toEqual([
      { text: "Storage" },
      { route: expect.any(Object), text: "AIPs" },
      { text: mockAip.name },
    ]);
  });

  it("fetches workflows", async () => {
    const mockWorkflows: api.AIPWorkflows = {
      workflows: [
        {
          uuid: "uuid-1",
          startedAt: new Date("2025-01-01T00:00:00Z"),
          status: api.EnduroStorageAipWorkflowStatusEnum.Done,
          type: api.EnduroStorageAipWorkflowTypeEnum.DeleteAip,
          temporalId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
        },
      ],
    };

    client.storage.storageListAipWorkflows = vi
      .fn()
      .mockResolvedValue(mockWorkflows);

    const store = useAipStore();
    await store.fetchWorkflows("uuid-1234");

    expect(store.currentWorkflows).toEqual(mockWorkflows);
  });

  it("fetches AIPs", async () => {
    const mockAips: api.AIPs = {
      items: [
        {
          name: "AIP 1",
          createdAt: new Date("2025-01-01T00:00:00Z"),
          objectKey: "object-key-1",
          status: api.AIPResponseStatusEnum.Stored,
          uuid: "aip-uuid-1",
        },
        {
          name: "AIP 2",
          createdAt: new Date("2025-01-02T00:00:00Z"),
          objectKey: "object-key-2",
          status: api.AIPResponseStatusEnum.Stored,
          uuid: "aip-uuid-2",
        },
      ],
      page: { limit: 20, offset: 0, total: 2 },
    };
    client.storage.storageListAips = vi.fn().mockResolvedValue(mockAips);

    const store = useAipStore();
    await store.fetchAips(1);

    expect(store.aips).toEqual(mockAips.items);
    expect(store.page).toEqual(mockAips.page);
  });
});
