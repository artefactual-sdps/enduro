import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, client } from "@/client";
import type { Pager } from "@/stores/aip";
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

  it("isRejected", () => {
    const aipStore = useAipStore();
    const now = new Date();

    expect(aipStore.isRejected).toEqual(false);

    aipStore.$patch({
      current: {
        createdAt: now,
        uuid: "uuid-1",
        status: api.EnduroStorageAipStatusEnum.Rejected,
      },
    });
    expect(aipStore.isRejected).toEqual(true);
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

  it("hasNextPage", () => {
    const aipStore = useAipStore();

    aipStore.$patch({
      page: { limit: 20, offset: 0, total: 20 },
    });
    expect(aipStore.hasNextPage).toEqual(false);

    aipStore.$patch({
      page: { limit: 20, offset: 20, total: 40 },
    });
    expect(aipStore.hasNextPage).toEqual(false);

    aipStore.$patch({
      page: { limit: 20, offset: 0, total: 21 },
    });
    expect(aipStore.hasNextPage).toEqual(true);
  });

  it("hasPrevPage", () => {
    const aipStore = useAipStore();

    aipStore.$patch({
      page: { limit: 20, offset: 0, total: 40 },
    });
    expect(aipStore.hasPrevPage).toEqual(false);

    aipStore.$patch({
      page: { limit: 20, offset: 20, total: 40 },
    });
    expect(aipStore.hasPrevPage).toEqual(true);
  });

  it("returns lastResultOnPage", () => {
    const aipStore = useAipStore();

    aipStore.$patch({
      page: { limit: 20, offset: 0, total: 40 },
    });
    expect(aipStore.lastResultOnPage).toEqual(20);

    aipStore.$patch({
      page: { limit: 20, offset: 0, total: 7 },
    });
    expect(aipStore.lastResultOnPage).toEqual(7);

    aipStore.$patch({
      page: { limit: 20, offset: 20, total: 35 },
    });
    expect(aipStore.lastResultOnPage).toEqual(35);
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

  it("updates the pager", () => {
    const aipStore = useAipStore();

    aipStore.$patch({
      page: { limit: 20, offset: 60, total: 125 },
    });
    aipStore.updatePager();
    expect(aipStore.pager).toEqual(<Pager>{
      maxPages: 7,
      current: 4,
      first: 1,
      last: 7,
      total: 7,
      pages: [1, 2, 3, 4, 5, 6, 7],
    });

    aipStore.$patch({
      page: { limit: 20, offset: 160, total: 573 },
    });
    aipStore.updatePager();
    expect(aipStore.pager).toEqual(<Pager>{
      maxPages: 7,
      current: 9,
      first: 6,
      last: 12,
      total: 29,
      pages: [6, 7, 8, 9, 10, 11, 12],
    });

    aipStore.$patch({
      page: { limit: 20, offset: 540, total: 573 },
    });
    aipStore.updatePager();
    expect(aipStore.pager).toEqual(<Pager>{
      maxPages: 7,
      current: 28,
      first: 23,
      last: 29,
      total: 29,
      pages: [23, 24, 25, 26, 27, 28, 29],
    });
  });
});
