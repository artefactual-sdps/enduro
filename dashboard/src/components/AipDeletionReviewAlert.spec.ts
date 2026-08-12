import { createTestingPinia } from "@pinia/testing";
import { flushPromises, mount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api } from "@/client";
import AipDeletionReviewAlert from "@/components/AipDeletionReviewAlert.vue";
import { useAipStore } from "@/stores/aip";

const mountAlert = async (canCancel: boolean, attributes: string[]) => {
  const pinia = createTestingPinia({
    createSpy: vi.fn,
    initialState: {
      aip: {
        current: {
          status: api.EnduroStorageAipStatusEnum.Pending,
          uuid: "aip-uuid",
        },
      },
      auth: {
        config: { enabled: true, abac: { enabled: true } },
        attributes,
      },
    },
  });
  const aipStore = useAipStore(pinia);
  vi.mocked(aipStore.canCancelDeletion).mockResolvedValue(canCancel);

  const wrapper = mount(AipDeletionReviewAlert, {
    props: { note: "Deletion requested by user@example.com" },
    global: { plugins: [pinia] },
  });
  await flushPromises();

  return wrapper;
};

describe("AipDeletionReviewAlert.vue", () => {
  afterEach(() => vi.clearAllMocks());

  it("shows no actions when none are available", async () => {
    const wrapper = await mountAlert(false, []);

    expect(wrapper.get("p").classes()).toContain("mb-0");
    expect(wrapper.find(".d-flex.flex-wrap.gap-2").exists()).toBe(false);
  });

  it("shows the cancel action when cancellation is available", async () => {
    const wrapper = await mountAlert(true, []);

    expect(wrapper.get("p").classes()).not.toContain("mb-0");
    expect(wrapper.get(".d-flex.flex-wrap.gap-2").text()).toBe("Cancel");
  });

  it("shows the review actions when review is available", async () => {
    const wrapper = await mountAlert(false, ["storage:aips:deletion:review"]);

    expect(wrapper.get("p").classes()).not.toContain("mb-0");
    expect(wrapper.get(".d-flex.flex-wrap.gap-2").text()).toContain("Approve");
    expect(wrapper.get(".d-flex.flex-wrap.gap-2").text()).toContain("Reject");
  });
});
