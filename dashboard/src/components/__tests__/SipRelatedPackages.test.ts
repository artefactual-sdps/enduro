import { createTestingPinia } from "@pinia/testing";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createRouter, createWebHistory } from "vue-router";

import { api } from "@/client";
import SipRelatedPackages from "@/components/SipRelatedPackages.vue";
import { useSipStore } from "@/stores/sip";

const router = createRouter({
  history: createWebHistory(),
  routes: [
    { name: "/", path: "/", component: {} },
    { name: "/storage/aips/[id]/", path: "/storage/aips/:id/", component: {} },
  ],
});

describe("SipRelatedPackages.vue", () => {
  afterEach(() => cleanup());

  it("renders nothing", () => {
    const { queryByText } = render(SipRelatedPackages, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: { current: null },
            },
          }),
          router,
        ],
      },
    });

    expect(queryByText("Related Packages")).toBeNull();
  });

  it("shows AIP and view button", () => {
    const { getByText, getByRole } = render(SipRelatedPackages, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: { current: { aipId: "aip-uuid" } },
              auth: {
                config: { enabled: true, abac: { enabled: true } },
                attributes: ["storage:aips:read"],
              },
            },
          }),
          router,
        ],
      },
    });

    getByText("Related Packages");
    getByText("AIP");
    getByText("aip-uuid");
    getByRole("link", { name: "View" });
  });

  it("shows AIP without view button", () => {
    const { getByText, queryByRole } = render(SipRelatedPackages, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: { current: { aipId: "aip-uuid" } },
              auth: {
                config: { enabled: true, abac: { enabled: true } },
                attributes: [],
              },
            },
          }),
          router,
        ],
      },
    });

    getByText("Related Packages");
    getByText("AIP");
    getByText("aip-uuid");
    expect(queryByRole("link", { name: "View" })).toBeNull();
  });

  it("shows failed SIP and download button", async () => {
    const { getByText, getByRole } = render(SipRelatedPackages, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: {
                current: {
                  failedAs: api.EnduroIngestSipFailedAsEnum.Sip,
                  failedKey: "failed-sip.zip",
                },
              },
              auth: {
                config: { enabled: true, abac: { enabled: true } },
                attributes: ["ingest:sips:download"],
              },
            },
          }),
          router,
        ],
      },
    });

    const sipStore = useSipStore();

    getByText("Related Packages");
    getByText("Failed SIP");
    getByText("failed-sip.zip");
    const btn = getByRole("button", { name: "Download" });
    await fireEvent.click(btn);
    expect(sipStore.download).toHaveBeenCalled();
  });

  it("shows failed PIP and download button", async () => {
    const { getByText, getByRole } = render(SipRelatedPackages, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: {
                current: {
                  failedAs: api.EnduroIngestSipFailedAsEnum.Pip,
                  failedKey: "failed-pip.zip",
                },
              },
              auth: {
                config: { enabled: true, abac: { enabled: true } },
                attributes: ["ingest:sips:download"],
              },
            },
          }),
          router,
        ],
      },
    });

    const sipStore = useSipStore();

    getByText("Related Packages");
    getByText("Failed PIP");
    getByText("failed-pip.zip");
    const btn = getByRole("button", { name: "Download" });
    await fireEvent.click(btn);
    expect(sipStore.download).toHaveBeenCalled();
  });

  it("shows failed PIP without download button", () => {
    const { getByText, queryByRole } = render(SipRelatedPackages, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: {
                current: {
                  failedAs: api.EnduroIngestSipFailedAsEnum.Pip,
                  failedKey: "failed-pip.zip",
                },
              },
              auth: {
                config: { enabled: true, abac: { enabled: true } },
                attributes: [],
              },
            },
          }),
          router,
        ],
      },
    });

    getByText("Related Packages");
    getByText("Failed PIP");
    getByText("failed-pip.zip");
    expect(queryByRole("button", { name: "Download" })).toBeNull();
  });

  it("shows failed PIP with error message", () => {
    const { getByText, getByRole, queryByRole } = render(SipRelatedPackages, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: {
                current: {
                  failedAs: api.EnduroIngestSipFailedAsEnum.Pip,
                  failedKey: "failed-pip.zip",
                },
                downloadError: "Download failed",
              },
              auth: {
                config: { enabled: true, abac: { enabled: true } },
                attributes: ["ingest:sips:download"],
              },
            },
          }),
          router,
        ],
      },
    });

    getByText("Related Packages");
    getByText("Failed PIP");
    getByText("failed-pip.zip");
    getByRole("alert");
    getByText("Download failed");
    expect(queryByRole("button", { name: "Download" })).toBeNull();
  });

  it("shows AIP and view button over failed SIP and download button", () => {
    const { getByText, getByRole, queryByText, queryByRole } = render(
      SipRelatedPackages,
      {
        global: {
          plugins: [
            createTestingPinia({
              createSpy: vi.fn,
              initialState: {
                sip: {
                  current: {
                    aipId: "aip-uuid",
                    failedAs: api.EnduroIngestSipFailedAsEnum.Sip,
                    failedKey: "failed-sip.zip",
                  },
                },
                auth: {
                  config: { enabled: true, abac: { enabled: true } },
                  attributes: ["storage:aips:read", "ingest:sips:download"],
                },
              },
            }),
            router,
          ],
        },
      },
    );

    getByText("Related Packages");
    getByText("AIP");
    getByText("aip-uuid");
    getByRole("link", { name: "View" });
    expect(queryByText("Failed SIP")).toBeNull();
    expect(queryByText("failed-sip.zip")).toBeNull();
    expect(queryByText("Failed SIP")).toBeNull();
    expect(queryByRole("button", { name: "Download" })).toBeNull();
  });
});
