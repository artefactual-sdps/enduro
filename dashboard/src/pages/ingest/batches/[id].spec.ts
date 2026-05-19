import { createTestingPinia } from "@pinia/testing";
import { shallowMount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";
import { createMemoryHistory, createRouter } from "vue-router";

import Page from "@/pages/ingest/batches/[id].vue";
import { useBatchStore } from "@/stores/batch";

describe("ingest/batches/[id].vue", () => {
  afterEach(() => vi.clearAllMocks());

  it("resets the batch store on unmount", async () => {
    const router = createRouter({
      history: createMemoryHistory(),
      routes: [
        {
          name: "/ingest/batches/[id]/",
          path: "/ingest/batches/:id/",
          component: {},
        },
      ],
    });
    await router.push("/ingest/batches/batch-uuid/");
    const pinia = createTestingPinia({ createSpy: vi.fn });
    vi.mocked(useBatchStore(pinia).fetchCurrent).mockResolvedValue(undefined);
    const wrapper = shallowMount(Page, {
      global: { plugins: [pinia, router] },
    });

    const reset = vi.spyOn(useBatchStore(), "$reset");
    wrapper.unmount();
    expect(reset).toHaveBeenCalledOnce();
  });
});
