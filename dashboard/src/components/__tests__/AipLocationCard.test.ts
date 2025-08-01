import { createTestingPinia } from "@pinia/testing";
import { cleanup, render } from "@testing-library/vue";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api } from "@/client";
import AipLocationCard from "@/components/AipLocationCard.vue";
import { useAipStore } from "@/stores/aip";

describe("AipLocationCard.vue", () => {
  afterEach(() => cleanup());

  it("renders when the AIP is stored", async () => {
    const { html } = render(AipLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              aip: {
                current: {
                  status: api.EnduroStorageAipStatusEnum.Stored,
                  locationUuid: "f8635e46-a320-4152-9a2c-98a28eeb50d1",
                } as api.EnduroStorageAip,
              },
            },
          }),
        ],
      },
    });

    expect(html()).toMatchInlineSnapshot(`
      "<div data-v-09f60ec4="" class="card mb-3">
        <div data-v-09f60ec4="" class="card-body">
          <!--v-if-->
          <!--v-if-->
          <h4 data-v-09f60ec4="" class="card-title">Location</h4>
          <p data-v-09f60ec4="" class="card-text"><span data-v-09f60ec4=""><div data-v-09f60ec4="" class="d-flex align-items-start gap-2"><span class="font-monospace">f8635e46-a320-4152-9a2c-98a28eeb50d1</span><button class="btn btn-sm btn-link link-secondary p-0" data-bs-toggle="tooltip" data-bs-title="Copy to clipboard">
              <!-- Copied visual hint. -->
              <!-- Copy icon. --><span><svg viewBox="0 0 24 24" width="1.2em" height="1.2em" aria-hidden="true"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M8 4v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7.242a2 2 0 0 0-.602-1.43L16.083 2.57A2 2 0 0 0 14.685 2H10a2 2 0 0 0-2 2"></path><path d="M16 18v2a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V9a2 2 0 0 1 2-2h2"></path></g></svg><span class="visually-hidden">Copy to clipboard</span></span>
            </button>
        </div></span></p>
        <div data-v-09f60ec4="">
          <transition-stub data-v-09f60ec4="" mode="out-in" appear="false" persisted="false" css="true">
            <div data-v-09f60ec4="" class="d-flex flex-wrap gap-2"><button data-v-09f60ec4="" type="button" class="btn btn-primary btn-sm"> Download </button>
              <!--v-if--><button data-v-09f60ec4="" type="button" class="btn btn-primary btn-sm"> Delete </button>
            </div>
          </transition-stub>
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
              aip: {
                current: {
                  status: api.EnduroStorageAipStatusEnum.Stored,
                  locationUuid: "f8635e46-a320-4152-9a2c-98a28eeb50d1",
                } as api.EnduroStorageAip,
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
      "<div data-v-09f60ec4="" class="card mb-3">
        <div data-v-09f60ec4="" class="card-body">
          <!--v-if-->
          <!--v-if-->
          <h4 data-v-09f60ec4="" class="card-title">Location</h4>
          <p data-v-09f60ec4="" class="card-text"><span data-v-09f60ec4=""><div data-v-09f60ec4="" class="d-flex align-items-start gap-2"><span class="font-monospace">f8635e46-a320-4152-9a2c-98a28eeb50d1</span><button class="btn btn-sm btn-link link-secondary p-0" data-bs-toggle="tooltip" data-bs-title="Copy to clipboard">
              <!-- Copied visual hint. -->
              <!-- Copy icon. --><span><svg viewBox="0 0 24 24" width="1.2em" height="1.2em" aria-hidden="true"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M8 4v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7.242a2 2 0 0 0-.602-1.43L16.083 2.57A2 2 0 0 0 14.685 2H10a2 2 0 0 0-2 2"></path><path d="M16 18v2a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V9a2 2 0 0 1 2-2h2"></path></g></svg><span class="visually-hidden">Copy to clipboard</span></span>
            </button>
        </div></span></p>
        <div data-v-09f60ec4="">
          <transition-stub data-v-09f60ec4="" mode="out-in" appear="false" persisted="false" css="true">
            <div data-v-09f60ec4="" class="d-flex flex-wrap gap-2">
              <!--v-if-->
              <!--v-if-->
              <!--v-if-->
            </div>
          </transition-stub>
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
              aip: {
                current: {
                  status: api.EnduroStorageAipStatusEnum.Stored,
                  locationUuid: "f8635e46-a320-4152-9a2c-98a28eeb50d1",
                } as api.EnduroStorageAip,
              },
            },
          }),
        ],
      },
    });

    getByText("f8635e46-a320-4152-9a2c-98a28eeb50d1");

    const aipStore = useAipStore();

    const moveMock = vi.fn().mockImplementation(aipStore.move);
    moveMock.mockImplementation(async () => {
      aipStore.$patch((state) => {
        if (!state.current) return;
        state.current.status = api.EnduroStorageAipStatusEnum.Processing;
        state.locationChanging = true;
      });
    });
    aipStore.move = moveMock;

    vi.mock("@/dialogs", () => {
      return {
        openLocationDialog: () => "fe675e52-c761-46d0-8605-fae4bd10303e",
      };
    });

    //const button = getByText("Choose storage location");
    //await fireEvent.click(button);

    //getByText("The AIP is being moved into a new location.");
  });

  it("renders when the AIP location is not available", async () => {
    const { html } = render(AipLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              aip: {
                current: {
                  status: api.EnduroStorageAipStatusEnum.Processing,
                } as api.EnduroStorageAip,
              },
            },
          }),
        ],
      },
    });

    expect(html()).toMatchInlineSnapshot(`
      "<div data-v-09f60ec4="" class="card mb-3">
        <div data-v-09f60ec4="" class="card-body">
          <!--v-if-->
          <!--v-if-->
          <h4 data-v-09f60ec4="" class="card-title">Location</h4>
          <p data-v-09f60ec4="" class="card-text"><span data-v-09f60ec4="">Not available yet.</span></p>
          <div data-v-09f60ec4="">
            <transition-stub data-v-09f60ec4="" mode="out-in" appear="false" persisted="false" css="true">
              <div data-v-09f60ec4="" class="d-flex flex-wrap gap-2">
                <!--v-if-->
                <!--v-if-->
                <!--v-if-->
              </div>
            </transition-stub>
          </div>
        </div>
      </div>"
    `);
  });

  it("renders when the AIP is deleted", async () => {
    const { html } = render(AipLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              aip: {
                current: {
                  status: api.EnduroStorageAipStatusEnum.Deleted,
                  locationUuid: undefined,
                } as api.EnduroStorageAip,
              },
            },
          }),
        ],
      },
    });

    expect(html()).toMatchInlineSnapshot(`
      "<div data-v-09f60ec4="" class="card mb-3">
        <div data-v-09f60ec4="" class="card-body">
          <!--v-if-->
          <!--v-if-->
          <h4 data-v-09f60ec4="" class="card-title">Location</h4>
          <p data-v-09f60ec4="" class="card-text"><span data-v-09f60ec4="">AIP deleted.</span></p>
          <!--v-if-->
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
              aip: {
                current: {
                  status: api.EnduroStorageAipStatusEnum.Processing,
                  locationUuid: "f8635e46-a320-4152-9a2c-98a28eeb50d1",
                } as api.EnduroStorageAip,
              },
            },
          }),
        ],
      },
    });

    expect(html()).toMatchInlineSnapshot(`
      "<div data-v-09f60ec4="" class="card mb-3">
        <div data-v-09f60ec4="" class="card-body">
          <!--v-if-->
          <!--v-if-->
          <h4 data-v-09f60ec4="" class="card-title">Location</h4>
          <p data-v-09f60ec4="" class="card-text"><span data-v-09f60ec4=""><div data-v-09f60ec4="" class="d-flex align-items-start gap-2"><span class="font-monospace">f8635e46-a320-4152-9a2c-98a28eeb50d1</span><button class="btn btn-sm btn-link link-secondary p-0" data-bs-toggle="tooltip" data-bs-title="Copy to clipboard">
              <!-- Copied visual hint. -->
              <!-- Copy icon. --><span><svg viewBox="0 0 24 24" width="1.2em" height="1.2em" aria-hidden="true"><g fill="none" stroke="currentColor" stroke-linecap="round" stroke-linejoin="round" stroke-width="2"><path d="M8 4v12a2 2 0 0 0 2 2h8a2 2 0 0 0 2-2V7.242a2 2 0 0 0-.602-1.43L16.083 2.57A2 2 0 0 0 14.685 2H10a2 2 0 0 0-2 2"></path><path d="M16 18v2a2 2 0 0 1-2 2H6a2 2 0 0 1-2-2V9a2 2 0 0 1 2-2h2"></path></g></svg><span class="visually-hidden">Copy to clipboard</span></span>
            </button>
        </div></span></p>
        <div data-v-09f60ec4="">
          <transition-stub data-v-09f60ec4="" mode="out-in" appear="false" persisted="false" css="true">
            <div data-v-09f60ec4="" class="d-flex flex-wrap gap-2">
              <!--v-if-->
              <!--v-if-->
              <!--v-if-->
            </div>
          </transition-stub>
        </div>
      </div>
      </div>"
    `);
  });

  it("shows the download button", async () => {
    const { getByRole } = render(AipLocationCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              aip: {
                current: {
                  status: api.EnduroStorageAipStatusEnum.Stored,
                  locationUuid: "f8635e46-a320-4152-9a2c-98a28eeb50d1",
                } as api.EnduroStorageAip,
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
              aip: {
                current: {
                  status: api.EnduroStorageAipStatusEnum.Stored,
                  locationUuid: "f8635e46-a320-4152-9a2c-98a28eeb50d1",
                } as api.EnduroStorageAip,
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
