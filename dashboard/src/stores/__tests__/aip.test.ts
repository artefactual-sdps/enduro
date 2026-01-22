import { createPinia, setActivePinia } from "pinia";
import { afterEach, beforeEach, describe, expect, it, vi } from "vitest";

import { api, client } from "@/client";
import { ResponseError } from "@/openapi-generator";
import { useAipStore } from "@/stores/aip";
import { useLayoutStore } from "@/stores/layout";

const http500Error = new ResponseError(
  new Response(JSON.stringify({}), {
    status: 500,
    statusText: "Internal Server Error",
  }),
  "Response returned an error code",
);

vi.mock("@/client");

beforeEach(() => {
  vi.useFakeTimers();
  setActivePinia(createPinia());
});

afterEach(() => {
  vi.clearAllMocks();
});

describe("getters", () => {
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
});

describe("fetch current", () => {
  it("fetches current AIP", async () => {
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
    const http404Error = new ResponseError(
      <Response>{
        status: 404,
        statusText: "Not Found",
      },
      "Not Found",
    );
    const consoleErr = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);

    client.storage.storageShowAip = vi.fn().mockRejectedValue(http404Error);

    const store = useAipStore();
    try {
      await store.fetchCurrent("uuid-1234");
    } catch (e) {
      expect(e).toEqual(http404Error);
    }

    expect(consoleErr).toBeCalledWith("Error fetching AIP:", 404, "Not Found");
    expect(store.current).toEqual(null);
    expect(store.locationChanging).toEqual(false);

    const layoutStore = useLayoutStore();
    expect(layoutStore.breadcrumb).toEqual([
      { text: "Storage" },
      { route: expect.any(Object), text: "AIPs" },
      { text: "Error" },
    ]);
  });

  it("fetches workflows", async () => {
    const mockWorkflows: api.EnduroStorageAipWorkflows = {
      workflows: [
        {
          uuid: "uuid-1",
          startedAt: new Date("2025-01-01T00:00:00Z"),
          status: api.EnduroStorageAipWorkflowStatusEnum.Done,
          type: api.EnduroStorageAipWorkflowTypeEnum.DeleteAip,
          temporalId: "c18d00f2-a1c4-4161-820c-6fc6ce707811",
          aipUuid: "aip-uuid-1",
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

  it("throws a workflow error", async () => {
    const http400Error = new ResponseError(
      <Response>{
        status: 400,
        statusText: "Bad Request",
      },
      "Bad Request",
    );
    const consoleErr = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);
    const mockAip: api.AIPResponse = {
      name: "AIP 1",
      createdAt: new Date("2025-01-01T00:00:00Z"),
      objectKey: "object-key-1",
      status: api.AIPResponseStatusEnum.Stored,
      uuid: "aip-uuid-1",
    };

    client.storage.storageShowAip = vi.fn().mockResolvedValue(mockAip);
    client.storage.storageListAipWorkflows = vi
      .fn()
      .mockRejectedValue(http400Error);

    const store = useAipStore();
    try {
      await store.fetchCurrent("aip-uuid-1");
    } catch (e) {
      expect(e).toEqual(new Error("Couldn't load workflows"));
    }

    expect(consoleErr).toHaveBeenCalledOnce();
    expect(consoleErr).toHaveBeenCalledWith(
      "Error fetching workflows:",
      400,
      "Bad Request",
    );
    expect(store.current).toEqual(mockAip);
    expect(store.currentWorkflows).toEqual(null);
  });
});

describe("fetch AIPs", () => {
  it("fetches AIPs", async () => {
    const mockAips: api.EnduroStorageAips = {
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

  it("reports a range error", async () => {
    const consoleErr = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);
    const store = useAipStore();

    client.storage.storageListAips = vi
      .fn()
      .mockRejectedValue(new RangeError("invalid date"));

    try {
      await store.fetchAips(1);
    } catch (e) {
      expect(e).toEqual(new Error("invalid date"));
    }

    expect(consoleErr).toHaveBeenCalledWith(
      "Error fetching AIPs (range error):",
      "invalid date",
    );
  });

  it("throws an error when fetching AIPs fails", async () => {
    const store = useAipStore();
    const consoleErr = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);
    client.storage.storageListAips = vi.fn().mockRejectedValue(http500Error);

    try {
      await store.fetchAips(1);
    } catch (e) {
      expect(e).toEqual(new Error("Couldn't load AIPs"));
    }

    expect(consoleErr).toHaveBeenCalledWith(
      "Error fetching AIPs:",
      500,
      "Internal Server Error",
    );
    expect(store.aips).toEqual([]);
  });
});

describe("cancel deletion request", () => {
  it("cancels a deletion request", async () => {
    const store = useAipStore();
    const pendingAIP: api.AIPResponse = {
      name: "AIP 1",
      createdAt: new Date("2025-01-01T00:00:00Z"),
      objectKey: "object-key-1",
      status: api.EnduroStorageAipStatusEnum.Pending,
      uuid: "aip-uuid-1",
    };

    store.$patch({
      current: pendingAIP,
    });

    client.storage.storageCancelAipDeletion = vi.fn().mockResolvedValue({});

    store.cancelDeletionRequest();

    expect(client.storage.storageCancelAipDeletion).toHaveBeenCalledWith({
      uuid: pendingAIP.uuid,
      cancelAipDeletionRequestBody: {},
    });
  });

  it("throws an error cancelling a deletion request", async () => {
    const store = useAipStore();
    const aip: api.AIPResponse = {
      name: "AIP 1",
      createdAt: new Date("2025-01-01T00:00:00Z"),
      objectKey: "object-key-1",
      status: api.EnduroStorageAipStatusEnum.Pending,
      uuid: "aip-uuid-1",
    };
    const consoleErr = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);

    client.storage.storageCancelAipDeletion = vi
      .fn()
      .mockRejectedValue(http500Error);

    store.$patch({
      current: aip,
    });

    try {
      await store.cancelDeletionRequest();
    } catch (e) {
      expect(e).toEqual(new Error("Couldn't cancel deletion request"));
    }

    expect(consoleErr).toHaveBeenCalledWith(
      "Error cancelling deletion request:",
      500,
      "Internal Server Error",
    );
  });
});

describe("canCancelDeletion", () => {
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
          new Response("Unauthorized", {
            status: 401,
            statusText: "Unauthorized",
          }),
          "Response returned an error code",
        ),
      )
      .mockRejectedValueOnce(
        new ResponseError(
          new Response("Forbidden", {
            status: 403,
            statusText: "Forbidden",
          }),
          "Response returned an error code",
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
      401,
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

  it("logs an error if canCancelDeletion returns an error response", async () => {
    const aip: api.AIPResponse = {
      name: "AIP 1",
      createdAt: new Date("2025-01-01T00:00:00Z"),
      objectKey: "object-key-1",
      status: api.EnduroStorageAipStatusEnum.Pending,
      uuid: "aip-uuid-1",
    };

    const consoleErr = vi
      .spyOn(console, "error")
      .mockImplementation(() => undefined);
    client.storage.storageCancelAipDeletion = vi
      .fn()
      .mockRejectedValue(http500Error);

    const store = useAipStore();
    store.$patch({
      current: aip,
    });

    const res = await store.canCancelDeletion();

    expect(consoleErr).toHaveBeenCalledWith(
      "Error checking user authorization to cancel deletion:",
      500,
      "Internal Server Error",
    );
    expect(res).toBe(false);
  });
});
