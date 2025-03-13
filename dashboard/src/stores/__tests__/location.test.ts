import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, client } from "@/client";
import { useLayoutStore } from "@/stores/layout";
import { useLocationStore } from "@/stores/location";

vi.mock("@/client");

describe("useLocationStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("fetches current", async () => {
    const mockLocation: api.LocationResponse = {
      name: "Location 1",
      createdAt: new Date("2025-01-01T00:00:00Z"),
      purpose: api.LocationResponsePurposeEnum.AipStore,
      source: api.LocationResponseSourceEnum.Amss,
      uuid: "uuid-1",
    };
    const mockAips: api.AIPResponse[] = [
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
    ];
    client.storage.storageShowLocation = vi
      .fn()
      .mockResolvedValue(mockLocation);
    client.storage.storageListLocationAips = vi
      .fn()
      .mockResolvedValue(mockAips);

    const store = useLocationStore();
    await store.fetchCurrent("uuid-1234");

    expect(store.current).toEqual(mockLocation);
    expect(store.currentAips).toEqual(mockAips);

    const layoutStore = useLayoutStore();
    expect(layoutStore.breadcrumb).toEqual([
      { text: "Storage" },
      { route: expect.any(Object), text: "Locations" },
      { text: mockLocation.name },
    ]);
  });

  it("fetches locations", async () => {
    const mockLocations: api.LocationResponse[] = [
      {
        name: "Location 1",
        createdAt: new Date("2025-01-01T00:00:00Z"),
        purpose: api.LocationResponsePurposeEnum.AipStore,
        source: api.LocationResponseSourceEnum.Amss,
        uuid: "uuid-1",
      },
      {
        name: "Location 2",
        createdAt: new Date("2025-01-02T00:00:00Z"),
        purpose: api.LocationResponsePurposeEnum.AipStore,
        source: api.LocationResponseSourceEnum.Amss,
        uuid: "uuid-2",
      },
    ];
    client.storage.storageListLocations = vi
      .fn()
      .mockResolvedValue(mockLocations);

    const store = useLocationStore();
    await store.fetchLocations();

    expect(store.locations).toEqual(mockLocations);
  });
});
