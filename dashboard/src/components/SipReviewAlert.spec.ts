import { createTestingPinia } from "@pinia/testing";
import { flushPromises, mount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api } from "@/client";
import LocationDialog from "@/components/LocationDialog.vue";
import SipReviewAlert from "@/components/SipReviewAlert.vue";
import { useAipStore } from "@/stores/aip";
import { useSipStore } from "@/stores/sip";

const openDialogMock = vi.hoisted(() => vi.fn());

vi.mock("@/dialogs/dialog", () => ({
  openDialog: openDialogMock,
}));

const mountPendingAlert = () =>
  mount(SipReviewAlert, {
    global: {
      plugins: [
        createTestingPinia({
          createSpy: vi.fn,
          initialState: {
            sip: {
              current: {
                status: api.EnduroIngestSipStatusEnum.Pending,
                uuid: "sip-1",
              },
            },
          },
        }),
      ],
    },
  });

describe("SipReviewAlert.vue", () => {
  afterEach(() => vi.clearAllMocks());

  it("resets the AIP store on unmount", () => {
    const wrapper = mount(SipReviewAlert, {
      global: { plugins: [createTestingPinia({ createSpy: vi.fn })] },
    });

    const reset = vi.spyOn(useAipStore(), "$reset");
    wrapper.unmount();
    expect(reset).toHaveBeenCalledOnce();
  });

  it("does not offer the obsolete task expansion action", () => {
    const wrapper = mountPendingAlert();

    expect(wrapper.text()).toContain("Task: Review AIP");
    expect(wrapper.text()).not.toContain("Expand");
  });

  it("confirms the SIP with the selected location", async () => {
    openDialogMock.mockResolvedValue("location-1");
    const wrapper = mountPendingAlert();
    const sipStore = useSipStore();

    await wrapper.get("button.btn-success").trigger("click");
    await flushPromises();

    expect(openDialogMock).toHaveBeenCalledWith(LocationDialog);
    expect(sipStore.confirm).toHaveBeenCalledWith("location-1");
  });

  it("does not confirm the SIP when location selection is cancelled", async () => {
    openDialogMock.mockResolvedValue(null);
    const wrapper = mountPendingAlert();
    const sipStore = useSipStore();

    await wrapper.get("button.btn-success").trigger("click");
    await flushPromises();

    expect(openDialogMock).toHaveBeenCalledWith(LocationDialog);
    expect(sipStore.confirm).not.toHaveBeenCalled();
  });
});
