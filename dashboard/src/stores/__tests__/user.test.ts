import { createPinia, setActivePinia } from "pinia";
import { beforeEach, describe, expect, it, vi } from "vitest";

import { api, client } from "@/client";
import { ResponseError } from "@/openapi-generator";
import { useUserStore } from "@/stores/user";

vi.mock("@/client");

describe("hasUsers", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("returns expected results", () => {
    const userStore = useUserStore();
    expect(userStore.hasUsers).toEqual(false);

    userStore.users = <api.EnduroIngestUser[]>[
      {
        createdAt: new Date(),
        email: "nobody@example.com",
        name: "Nobody Example",
        uuid: "a499e8fc-7309-4e26-b39d-d8ab68466c27",
      },
    ];
    expect(userStore.hasUsers).toEqual(true);
  });
});

describe("fetchUsers", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("fetches users", async () => {
    const store = useUserStore();
    const mockUsers: api.Users = {
      items: [
        {
          createdAt: new Date("2025-01-01T00:00:00Z"),
          email: "user1@example.com",
          name: "User One",
          uuid: "a499e8fc-7309-4e26-b39d-d8ab68466c27",
        },
        {
          createdAt: new Date("2025-01-02T00:00:00Z"),
          email: "user2@example.com",
          name: "User Two",
          uuid: "30223842-0650-4f79-80bd-7bf43b810656",
        },
      ],
      page: { limit: 20, offset: 0, total: 2 },
    };

    client.ingest.ingestListUsers = vi.fn().mockResolvedValue(mockUsers);

    await store.fetchUsers();

    expect(store.users).toEqual(mockUsers.items);
    expect(store.page).toEqual(mockUsers.page);
  });

  it("errors when receiving and error response", async () => {
    const store = useUserStore();
    const logError = vi.spyOn(console, "error").mockImplementation(() => {});

    client.ingest.ingestListUsers = vi.fn().mockRejectedValue(
      new ResponseError(
        new Response("Forbidden", {
          status: 403,
          statusText: "Forbidden",
        }),
        "Response returned an error code",
      ),
    );

    await expect(store.fetchUsers()).resolves.toBeUndefined();
    expect(store.users).toEqual([]);
    expect(logError).toHaveBeenCalledWith(
      "Failed to fetch users:",
      403,
      "Forbidden",
    );
  });
});

describe("getHandle", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("returns the user name when available", () => {
    const store = useUserStore();
    const user = <api.EnduroIngestUser>{
      createdAt: new Date("2025-01-01T00:00:00Z"),
      email: "nobody@example.com",
      name: "Nobody Example",
      uuid: "a499e8fc-7309-4e26-b39d-d8ab68466c27",
    };

    expect(store.getHandle(user)).toEqual("Nobody Example");
  });

  it("returns the email if name isn't set", () => {
    const store = useUserStore();
    const user = <api.EnduroIngestUser>{
      createdAt: new Date("2025-01-01T00:00:00Z"),
      email: "nobody@example.com",
      name: "",
      uuid: "a499e8fc-7309-4e26-b39d-d8ab68466c27",
    };

    expect(store.getHandle(user)).toEqual("nobody@example.com");
  });

  it("returns the UUID if name and email aren't set", () => {
    const store = useUserStore();
    const user = <api.EnduroIngestUser>{
      createdAt: new Date("2025-01-01T00:00:00Z"),
      email: "",
      name: "",
      uuid: "a499e8fc-7309-4e26-b39d-d8ab68466c27",
    };

    expect(store.getHandle(user)).toEqual(
      "a499e8fc-7309-4e26-b39d-d8ab68466c27",
    );
  });
});
