import { api } from "../../src/client";
import PackageDetailsCard from "../../src/components/PackageDetailsCard.vue";
import { usePackageStore } from "../../src/stores/package";
import { createTestingPinia } from "@pinia/testing";
import { render } from "@testing-library/vue";
import { flushPromises } from "@vue/test-utils";
import { expect, describe, it, vi } from "vitest";

describe("PackageDetailsCard.vue", () => {
  it("watches download requests from the store", async () => {
    render(PackageDetailsCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              package: {
                current: {
                  aipId: "89229d18-5554-4e0d-8c4e-d0d88afd3bae",
                  status: api.PackageShowResponseBodyStatusEnum.Pending,
                } as api.PackageShowResponseBody,
              },
            },
          }),
        ],
      },
    });

    vi.stubGlobal("open", vi.fn());

    // Someone requests the download of the AIP via the package store.
    const packageStore = usePackageStore();
    packageStore.ui.download.request();
    await flushPromises();

    // Then we observe that the component download function is executed.
    expect(window.open).toBeCalledWith(
      "///api/storage/89229d18-5554-4e0d-8c4e-d0d88afd3bae/download",
      "_blank"
    );
  });

  it("renders when the package is in pending status", async () => {
    const { html } = render(PackageDetailsCard, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              package: {
                current: {
                  aipId: "89229d18-5554-4e0d-8c4e-d0d88afd3bae",
                  status: api.PackageShowResponseBodyStatusEnum.Pending,
                } as api.PackageShowResponseBody,
              },
            },
          }),
        ],
      },
    });

    expect(html()).toMatchInlineSnapshot(`
      "<div class=\\"card mb-3\\">
        <div class=\\"card-body\\">
          <h5 class=\\"card-title\\">Package details</h5>
          <dl>
            <dt>Original objects</dt>
            <dd>N/A</dd>
            <dt>Package size</dt>
            <dd>N/A</dd>
            <dt>Last workflow outcome</dt>
            <dd><span><span class=\\"badge text-bg-secondary\\">PENDING</span><span class=\\"badge text-dark fw-normal\\">(Create and Review AIP)</span></span></dd>
          </dl>
          <div class=\\"d-flex flex-wrap gap-2\\"><button class=\\"btn btn-secondary btn-sm disabled\\"> View metadata summary </button><button class=\\"btn btn-primary btn-sm\\" type=\\"button\\"> Download </button></div>
        </div>
      </div>"
    `);
  });
});
