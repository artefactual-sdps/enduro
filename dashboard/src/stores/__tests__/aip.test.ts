import { createPinia, setActivePinia } from "pinia";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { api, client } from "@/client";
import { ResponseError } from "@/openapi-generator";
import { useAipStore } from "@/stores/aip";
import { useLayoutStore } from "@/stores/layout";

vi.mock("@/client");

describe("useAipStore", () => {
  beforeEach(() => {
    vi.useFakeTimers();
    setActivePinia(createPinia());
  });

  afterEach(() => {
    vi.clearAllMocks();
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
    client.storage.storageListAipWorkflows = vi.fn().mockResolvedValue([]);

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

  it("throws a not found error", async () => {
    const consoleErr = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);
    client.storage.storageShowAip = vi.fn().mockRejectedValue(
      new ResponseError(
        <Response>{
          status: 404,
          statusText: "Not Found",
        },
        "Not Found",
      ),
    );

    const store = useAipStore();
    try {
      await store.fetchCurrent("uuid-1234");
    } catch (e) {
      expect(e).toEqual(new Error("AIP not found"));
    }

    expect(consoleErr).toHaveBeenCalledOnce();
    expect(store.current).toEqual(null);
    expect(store.locationChanging).toEqual(false);

    const layoutStore = useLayoutStore();
    expect(layoutStore.breadcrumb).toEqual([
      { text: "Storage" },
      { route: expect.any(Object), text: "AIPs" },
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

  it("cancels a deletion request", async () => {
    const pendingAIP: api.AIPResponse = {
      name: "AIP 1",
      createdAt: new Date("2025-01-01T00:00:00Z"),
      objectKey: "object-key-1",
      status: api.EnduroStorageAipStatusEnum.Pending,
      uuid: "aip-uuid-1",
    };

    const storedAIP: api.AIPResponse = {
      name: "AIP 1",
      createdAt: new Date("2025-01-01T00:00:00Z"),
      objectKey: "object-key-1",
      status: api.EnduroStorageAipStatusEnum.Stored,
      uuid: "aip-uuid-1",
    };

    const mockWorkflows: api.AIPWorkflows[] = [];

    const store = useAipStore();
    store.$patch({
      current: pendingAIP,
    });

    client.storage.storageCancelAipDeletion = vi.fn();
    client.storage.storageShowAip = vi
      .fn()
      .mockResolvedValueOnce(pendingAIP)
      .mockResolvedValueOnce(storedAIP);
    client.storage.storageListAipWorkflows = vi
      .fn()
      .mockResolvedValue(mockWorkflows);

    const p = store.cancelDeletionRequest();

    // Fast-forward the timer by 3.5 seconds to allow for three polling calls
    // to "Show AIP".
    await vi.advanceTimersByTimeAsync(3500);
    await p;

    expect(client.storage.storageCancelAipDeletion).toHaveBeenCalledWith({
      uuid: pendingAIP.uuid,
      cancelAipDeletionRequestBody: {},
    });

    // "Show AIP" should only be called twice, because the second call returns a
    // "stored" status which cancels polling.
    expect(client.storage.storageShowAip).toHaveBeenCalledWith({
      uuid: pendingAIP.uuid,
    });
    expect(client.storage.storageShowAip).toHaveBeenCalledTimes(2);

    expect(client.storage.storageListAipWorkflows).toHaveBeenCalledWith({
      uuid: pendingAIP.uuid,
    });
    expect(client.storage.storageListAipWorkflows).toHaveBeenCalledTimes(2);

    expect(store.current).toEqual(storedAIP);
  });

  it("checks if user can cancel deletion", async () => {
    const mockAip: api.AIPResponse = {
      name: "AIP 1",
      createdAt: new Date("2025-01-01T00:00:00Z"),
      objectKey: "object-key-1",
      status: api.EnduroStorageAipStatusEnum.Stored,
      uuid: "aip-uuid-1",
    };

    const store = useAipStore();
    store.$patch({
      current: mockAip,
    });

    const consoleErr = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);

    client.storage.storageCancelAipDeletion = vi
      .fn()
      .mockRejectedValueOnce(
        new ResponseError(
          <Response>{
            status: 401,
            statusText: "Unauthorized",
          },
          "Unauthorized",
        ),
      )
      .mockRejectedValueOnce(
        new ResponseError(
          <Response>{
            status: 403,
            statusText: "Forbidden",
          },
          "Forbidden",
        ),
      )
      .mockResolvedValueOnce({});

    // First call (401 Unauthorized) should return false and emit a console
    // error.
    let result = await store.canCancelDeletion();
    expect(client.storage.storageCancelAipDeletion).toBeCalledWith({
      uuid: "aip-uuid-1",
      cancelAipDeletionRequestBody: {
        check: true,
      },
    });
    expect(result).toBe(false);
    expect(consoleErr).toBeCalledWith(
      "Error checking user authorization to cancel deletion:",
      "Unauthorized",
    );

    // Second call (403 Forbidden) should return false.
    result = await store.canCancelDeletion();
    expect(client.storage.storageCancelAipDeletion).toBeCalledWith({
      uuid: "aip-uuid-1",
      cancelAipDeletionRequestBody: {
        check: true,
      },
    });
    expect(result).toBe(false);

    // Third call should return true.
    result = await store.canCancelDeletion();
    expect(client.storage.storageCancelAipDeletion).toBeCalledWith({
      uuid: "aip-uuid-1",
      cancelAipDeletionRequestBody: {
        check: true,
      },
    });
    expect(result).toBe(true);
  });
});
