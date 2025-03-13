import { createTestingPinia } from "@pinia/testing";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";

import { api } from "@/client";
import AipLocationCard from "@/components/AipLocationCard.vue";
import { useSipStore } from "@/stores/sip";

describe("AipLocationCard.vue", () => {
  afterEach(() => cleanup());

  it("renders when the AIP is stored", async () => {
    const { html } = render(AipLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: {
                current: {
                  status: api.EnduroIngestSipStatusEnum.Done,
                  locationId: "f8635e46-a320-4152-9a2c-98a28eeb50d1",
                } as api.EnduroIngestSip,
              },
            },
          }),
        ],
      },
    });

    expect(html()).toMatchInlineSnapshot(`
      "<div class="card mb-3">
        <div class="card-body">
          <!--v-if-->
          <!--v-if-->
          <h4 class="card-title">Location</h4>
          <p class="card-text"><span><div class="d-flex align-items-start gap-2"><span class="font-monospace">f8635e46-a320-4152-9a2c-98a28eeb50d1</span><button class="btn btn-sm btn-link link-secondary p-0" data-bs-toggle="tooltip" data-bs-title="Copy to clipboard">
              <!-- Copied visual hint. -->
              <!-- Copy icon. --><span><svg viewBox="0 0 24 24" width="1.2em" height="1.2em" aria-hidden="true"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M8 4v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7.242a2 2 0 0 0-.602-1.43L16.083 2.57A2 2 0 0 0 14.685 2H10a2 2 0 0 0-2 2"></path><path d="M16 18v2a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V9a2 2 0 0 1 2-2h2"></path></g></svg><span class="visually-hidden">Copy to clipboard</span></span>
            </button>
        </div></span></p>
        <div class="d-flex flex-wrap gap-2"><button type="button" class="btn btn-primary btn-sm">Choose storage location</button>
          <!--v-if-->
        </div>
      </div>
      </div>"
    `);
  });

  it("renders without move button based on auth attributes", async () => {
    const { html } = render(AipLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: {
                current: {
                  status: api.EnduroIngestSipStatusEnum.Done,
                  locationId: "f8635e46-a320-4152-9a2c-98a28eeb50d1",
                } as api.EnduroIngestSip,
              },
              auth: {
                config: { enabled: true, abac: { enabled: true } },
                attributes: [],
              },
            },
          }),
        ],
      },
    });

    expect(html()).toMatchInlineSnapshot(`
      "<div class="card mb-3">
        <div class="card-body">
          <!--v-if-->
          <!--v-if-->
          <h4 class="card-title">Location</h4>
          <p class="card-text"><span><div class="d-flex align-items-start gap-2"><span class="font-monospace">f8635e46-a320-4152-9a2c-98a28eeb50d1</span><button class="btn btn-sm btn-link link-secondary p-0" data-bs-toggle="tooltip" data-bs-title="Copy to clipboard">
              <!-- Copied visual hint. -->
              <!-- Copy icon. --><span><svg viewBox="0 0 24 24" width="1.2em" height="1.2em" aria-hidden="true"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M8 4v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7.242a2 2 0 0 0-.602-1.43L16.083 2.57A2 2 0 0 0 14.685 2H10a2 2 0 0 0-2 2"></path><path d="M16 18v2a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V9a2 2 0 0 1 2-2h2"></path></g></svg><span class="visually-hidden">Copy to clipboard</span></span>
            </button>
        </div></span></p>
        <div class="d-flex flex-wrap gap-2">
          <!--v-if-->
          <!--v-if-->
        </div>
      </div>
      </div>"
    `);
  });

  it("renders when the AIP location is moved", async () => {
    const { getByText } = render(AipLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            stubActions: false,
            initialState: {
              sip: {
                current: {
                  status: api.EnduroIngestSipStatusEnum.Done,
                  locationId: "f8635e46-a320-4152-9a2c-98a28eeb50d1",
                } as api.EnduroIngestSip,
              },
            },
          }),
        ],
      },
    });

    getByText("f8635e46-a320-4152-9a2c-98a28eeb50d1");

    const sipStore = useSipStore();

    const moveMock = vi.fn().mockImplementation(sipStore.move);
    moveMock.mockImplementation(async () => {
      sipStore.$patch((state) => {
        if (!state.current) return;
        state.current.status = api.EnduroIngestSipStatusEnum.InProgress;
        state.locationChanging = true;
      });
    });
    sipStore.move = moveMock;

    vi.mock("@/dialogs", () => {
      return {
        openLocationDialog: () => "fe675e52-c761-46d0-8605-fae4bd10303e",
      };
    });

    const button = getByText("Choose storage location");
    await fireEvent.click(button);

    getByText("The AIP is being moved into a new location.");
  });

  it("renders when the AIP location is not available", async () => {
    const { html } = render(AipLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: {
                current: {
                  status: api.EnduroIngestSipStatusEnum.InProgress,
                } as api.EnduroIngestSip,
              },
            },
          }),
        ],
      },
    });

    expect(html()).toMatchInlineSnapshot(`
      "<div class="card mb-3">
        <div class="card-body">
          <!--v-if-->
          <!--v-if-->
          <h4 class="card-title">Location</h4>
          <p class="card-text"><span>Not available yet.</span></p>
          <div class="d-flex flex-wrap gap-2"><button type="button" class="btn btn-primary btn-sm" disabled="">Choose storage location</button>
            <!--v-if-->
          </div>
        </div>
      </div>"
    `);
  });

  it("renders when the AIP is rejected", async () => {
    const { html } = render(AipLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: {
                current: {
                  status: api.EnduroIngestSipStatusEnum.Done,
                  locationId: undefined,
                } as api.EnduroIngestSip,
              },
            },
          }),
        ],
      },
    });

    expect(html()).toMatchInlineSnapshot(`
      "<div class="card mb-3">
        <div class="card-body">
          <!--v-if-->
          <!--v-if-->
          <h4 class="card-title">Location</h4>
          <p class="card-text"><span>AIP rejected.</span></p>
          <div class="d-flex flex-wrap gap-2">
            <!--v-if-->
            <!--v-if-->
          </div>
        </div>
      </div>"
    `);
  });

  it("renders when the AIP is moving", async () => {
    const { html } = render(AipLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: {
                current: {
                  status: api.EnduroIngestSipStatusEnum.InProgress,
                  locationId: "f8635e46-a320-4152-9a2c-98a28eeb50d1",
                } as api.EnduroIngestSip,
              },
            },
          }),
        ],
      },
    });

    expect(html()).toMatchInlineSnapshot(`
      "<div class="card mb-3">
        <div class="card-body">
          <!--v-if-->
          <!--v-if-->
          <h4 class="card-title">Location</h4>
          <p class="card-text"><span><div class="d-flex align-items-start gap-2"><span class="font-monospace">f8635e46-a320-4152-9a2c-98a28eeb50d1</span><button class="btn btn-sm btn-link link-secondary p-0" data-bs-toggle="tooltip" data-bs-title="Copy to clipboard">
              <!-- Copied visual hint. -->
              <!-- Copy icon. --><span><svg viewBox="0 0 24 24" width="1.2em" height="1.2em" aria-hidden="true"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M8 4v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7.242a2 2 0 0 0-.602-1.43L16.083 2.57A2 2 0 0 0 14.685 2H10a2 2 0 0 0-2 2"></path><path d="M16 18v2a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V9a2 2 0 0 1 2-2h2"></path></g></svg><span class="visually-hidden">Copy to clipboard</span></span>
            </button>
        </div></span></p>
        <div class="d-flex flex-wrap gap-2"><button type="button" class="btn btn-primary btn-sm" disabled="">Choose storage location</button>
          <!--v-if-->
        </div>
      </div>
      </div>"
    `);
  });

  it("watches download requests from the store", async () => {
    render(AipLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: {
                current: {
                  aipId: "89229d18-5554-4e0d-8c4e-d0d88afd3bae",
                  status: api.EnduroIngestSipStatusEnum.Pending,
                } as api.EnduroIngestSip,
              },
            },
          }),
        ],
      },
    });

    vi.stubGlobal("open", vi.fn());

    // Someone requests the download of the AIP via the ingest store.
    const sipStore = useSipStore();
    sipStore.ui.download.request();
    await nextTick();

    // Then we observe that the component download function is executed.
    expect(window.open).toBeCalledWith(
      "http://localhost:3000/api/storage/aips/89229d18-5554-4e0d-8c4e-d0d88afd3bae/download",
      "_blank",
    );
  });

  it("shows the download button", async () => {
    const { getByRole } = render(AipLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: {
                current: {
                  aipId: "89229d18-5554-4e0d-8c4e-d0d88afd3bae",
                  status: api.EnduroIngestSipStatusEnum.Done,
                } as api.EnduroIngestSip,
              },
            },
          }),
        ],
      },
    });

    getByRole("button", { name: "Download" });
  });

  it("hides the download button", async () => {
    const { queryByRole } = render(AipLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: {
                current: {
                  aipId: "89229d18-5554-4e0d-8c4e-d0d88afd3bae",
                  status: api.EnduroIngestSipStatusEnum.Done,
                } as api.EnduroIngestSip,
              },
              auth: {
                config: { enabled: true, abac: { enabled: true } },
                attributes: [],
              },
            },
          }),
        ],
      },
    });

    expect(queryByRole("button", { name: "Download" })).toBeNull();
  });
});
