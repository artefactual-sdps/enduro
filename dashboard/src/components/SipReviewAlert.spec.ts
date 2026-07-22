import { createTestingPinia } from "@pinia/testing";
import { flushPromises, mount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";

import { api } from "@/client";
import SipReviewAlert from "@/components/SipReviewAlert.vue";
import { useAipStore } from "@/stores/aip";
import { useSipStore } from "@/stores/sip";

const openLocationDialogMock = vi.hoisted(() => vi.fn());

vi.mock("@/dialogs/location", () => ({
  openLocationDialog: openLocationDialogMock,
}));

const mountPendingAlert = () =>
  mount(SipReviewAlert, {
    props: { expandCounter: 0 },
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
      props: { expandCounter: 0 },
      global: { plugins: [createTestingPinia({ createSpy: vi.fn })] },
    });

    const reset = vi.spyOn(useAipStore(), "$reset");
    wrapper.unmount();
    expect(reset).toHaveBeenCalledOnce();
  });

  it("confirms the SIP with the selected location", async () => {
    openLocationDialogMock.mockResolvedValue("location-1");
    const wrapper = mountPendingAlert();
    const sipStore = useSipStore();

    await wrapper.get("button.btn-success").trigger("click");
    await flushPromises();

    expect(openLocationDialogMock).toHaveBeenCalledOnce();
    expect(sipStore.confirm).toHaveBeenCalledWith("location-1");
  });

  it("does not confirm the SIP when location selection is cancelled", async () => {
    openLocationDialogMock.mockResolvedValue(null);
    const wrapper = mountPendingAlert();
    const sipStore = useSipStore();

    await wrapper.get("button.btn-success").trigger("click");
    await flushPromises();

    expect(openLocationDialogMock).toHaveBeenCalledOnce();
    expect(sipStore.confirm).not.toHaveBeenCalled();
  });
});
