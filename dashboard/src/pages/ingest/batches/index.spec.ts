import { createTestingPinia } from "@pinia/testing";
import { shallowMount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createMemoryHistory, createRouter } from "vue-router";

import Page from "@/pages/ingest/batches/index.vue";
import { useBatchStore } from "@/stores/batch";
import { useUserStore } from "@/stores/user";

vi.mock("bootstrap/js/dist/dropdown", () => ({ default: class {} }));
vi.mock("bootstrap/js/dist/tooltip", () => ({ default: class {} }));

describe("ingest/batches/index.vue", () => {
  afterEach(() => vi.clearAllMocks());

  it("resets the batch and user stores on unmount", async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        { name: "/ingest/batches/", path: "/ingest/batches/", component: {} },
      ],
    });
    await router.push("/ingest/batches/");
    const wrapper = shallowMount(Page, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              auth: { config: { enabled: true, abac: { enabled: true } } },
            },
          }),
          router,
        ],
      },
    });

    const resetBatch = vi.spyOn(useBatchStore(), "$reset");
    const resetUser = vi.spyOn(useUserStore(), "$reset");
    wrapper.unmount();
    expect(resetBatch).toHaveBeenCalledOnce();
    expect(resetUser).toHaveBeenCalledOnce();
  });
});
