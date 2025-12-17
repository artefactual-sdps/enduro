import { createTestingPinia } from "@pinia/testing";
import { flushPromises, mount } from "@vue/test-utils";
import { afterEach, describe, expect, it, vi } from "vitest";

import SIPUploadSource from "@/components/SIPUploadSource.vue";

// Use hoisted mocks so vi.mock factories can access them.
const ingestListSipSourceObjects = vi.hoisted(() => vi.fn());
const ingestAddBatch = vi.hoisted(() => vi.fn());
const ingestAddSip = vi.hoisted(() => vi.fn());
const push = vi.hoisted(() => vi.fn());

vi.mock("@/client", () => ({
  api: {},
  client: {
    ingest: {
      ingestListSipSourceObjects,
      ingestAddBatch,
      ingestAddSip,
    },
  },
}));

vi.mock("vue-router/auto", () => ({
  useRouter: () => ({ push }),
}));

const mountOptions = (
  attributes: string[] = ["ingest:sips:create", "ingest:batches:create"],
) => ({
  global: {
    plugins: [
      createTestingPinia({
        createSpy: vi.fn,
        initialState: {
          auth: {
            config: { enabled: true, abac: { enabled: true } },
            attributes: attributes,
          },
        },
      }),
    ],
    config: {
      globalProperties: {
        $filters: { formatDateTime: vi.fn() },
      },
    },
  },
});

describe("SIPUploadSource.vue", () => {
  afterEach(() => {
    vi.clearAllMocks();
  });

  it("defaults to SIP upload and disables the switch when user lacks batch permission", async () => {
    ingestListSipSourceObjects.mockResolvedValueOnce({});

    const wrapper = mount(
      SIPUploadSource,
      mountOptions(["ingest:sips:create"]),
    );
    await flushPromises();

    const toggle = wrapper.get("#batch-switch");
    expect((toggle.element as HTMLInputElement).checked).toBe(false);
    expect((toggle.element as HTMLInputElement).disabled).toBe(true);
  });

  it("defaults to batch upload and disables the switch when user lacks SIP permission", async () => {
    ingestListSipSourceObjects.mockResolvedValueOnce({});

    const wrapper = mount(
      SIPUploadSource,
      mountOptions(["ingest:batches:create"]),
    );
    await flushPromises();

    const toggle = wrapper.get("#batch-switch");
    expect((toggle.element as HTMLInputElement).checked).toBe(true);
    expect((toggle.element as HTMLInputElement).disabled).toBe(true);
  });

  it("defaults to SIP upload and enables the switch when user has both permissions", async () => {
    ingestListSipSourceObjects.mockResolvedValueOnce({});

    const wrapper = mount(SIPUploadSource, mountOptions());
    await flushPromises();

    const toggle = wrapper.get("#batch-switch");
    expect((toggle.element as HTMLInputElement).checked).toBe(false);
    expect((toggle.element as HTMLInputElement).disabled).toBe(false);
  });

  it("uploads individual SIPs", async () => {
    ingestListSipSourceObjects.mockResolvedValueOnce({
      objects: [
        { key: "sip-1", size: 123, modTime: "2024-01-01T00:00:00Z" },
        { key: "sip-2", size: 456, modTime: "2024-01-02T00:00:00Z" },
      ],
    });
    ingestAddSip.mockResolvedValue({});

    const wrapper = mount(SIPUploadSource, mountOptions());
    await flushPromises();

    await wrapper.get("#cb-sip-1").setValue(true);
    await wrapper.get("#cb-sip-2").setValue(true);
    await wrapper.get("button.btn-primary").trigger("click");

    expect(ingestAddSip).toHaveBeenCalledWith({
      key: "sip-1",
      sourceId: "e6ddb29a-66d1-480e-82eb-fcfef1c825c5",
    });
    expect(ingestAddSip).toHaveBeenCalledWith({
      key: "sip-2",
      sourceId: "e6ddb29a-66d1-480e-82eb-fcfef1c825c5",
    });
    expect(ingestAddBatch).not.toHaveBeenCalled();
    expect(push).toHaveBeenCalledWith({ path: "/ingest/sips" });
  });

  it("uploads a batch with identifier", async () => {
    ingestListSipSourceObjects.mockResolvedValueOnce({
      objects: [
        { key: "sip-1", size: 123, modTime: "2024-01-01T00:00:00Z" },
        { key: "sip-2", size: 456, modTime: "2024-01-02T00:00:00Z" },
      ],
    });
    ingestAddBatch.mockResolvedValueOnce({});

    const wrapper = mount(SIPUploadSource, mountOptions());
    await flushPromises();

    await wrapper.get("#cb-sip-1").setValue(true);
    await wrapper.get("#cb-sip-2").setValue(true);
    await wrapper.get("#batch-switch").setValue(true);
    await wrapper.get("#batch-id").setValue(" custom-id ");
    await wrapper.get("button.btn-primary").trigger("click");

    expect(ingestAddBatch).toHaveBeenCalledWith({
      addBatchRequestBody: {
        keys: ["sip-1", "sip-2"],
        sourceId: "e6ddb29a-66d1-480e-82eb-fcfef1c825c5",
        identifier: "custom-id",
      },
    });
    expect(ingestAddSip).not.toHaveBeenCalled();
    expect(push).toHaveBeenCalledWith({ path: "/ingest/sips" });
  });

  it("uploads a batch without identifier", async () => {
    ingestListSipSourceObjects.mockResolvedValueOnce({
      objects: [{ key: "sip-1", size: 123, modTime: "2024-01-01T00:00:00Z" }],
    });
    ingestAddBatch.mockResolvedValueOnce({});

    const wrapper = mount(SIPUploadSource, mountOptions());
    await flushPromises();

    await wrapper.get("#cb-sip-1").setValue(true);
    await wrapper.get("#batch-switch").setValue(true);
    await wrapper.get("#batch-id").setValue("   ");
    await wrapper.get("button.btn-primary").trigger("click");

    expect(ingestAddBatch).toHaveBeenCalledWith({
      addBatchRequestBody: {
        keys: ["sip-1"],
        sourceId: "e6ddb29a-66d1-480e-82eb-fcfef1c825c5",
      },
    });
    expect(ingestAddSip).not.toHaveBeenCalled();
    expect(push).toHaveBeenCalledWith({ path: "/ingest/sips" });
  });

  it("shows an error when ingest fails", async () => {
    ingestListSipSourceObjects.mockResolvedValueOnce({
      objects: [{ key: "sip-1", size: 123, modTime: "2024-01-01T00:00:00Z" }],
    });
    ingestAddSip.mockRejectedValueOnce(new Error("API error"));

    const wrapper = mount(SIPUploadSource, mountOptions());
    await flushPromises();

    await wrapper.get("#cb-sip-1").setValue(true);
    await wrapper.get("button.btn-primary").trigger("click");

    expect(wrapper.text()).toContain("Failed to start ingest.");
    expect(push).not.toHaveBeenCalled();
  });
});
