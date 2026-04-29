import { createTestingPinia } from "@pinia/testing";
import { cleanup, fireEvent, render } from "@testing-library/vue";
import { afterEach, describe, expect, it, vi } from "vitest";

import SipDecisionAlert from "@/components/SipDecisionAlert.vue";
import { useSipStore } from "@/stores/sip";

describe("SipDecisionAlert.vue", () => {
  afterEach(() => cleanup());

  it("renders nothing without a current decision", () => {
    const { queryByRole } = render(SipDecisionAlert, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: { currentDecision: null },
            },
          }),
        ],
      },
    });

    expect(queryByRole("alert")).toBeNull();
  });

  it("shows the decision message and options", () => {
    const { getByLabelText, getByRole, getByText } = render(SipDecisionAlert, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: {
                currentDecision: {
                  message: "Choose how to continue",
                  options: ["Continue", "Cancel"],
                },
              },
            },
          }),
        ],
      },
    });

    getByRole("alert");
    getByText("Task: User decision required");
    getByText("Choose how to continue");
    expect((getByLabelText("Continue") as HTMLInputElement).checked).toBe(
      false,
    );
    expect((getByLabelText("Cancel") as HTMLInputElement).checked).toBe(false);
    expect(
      (getByRole("button", { name: "Submit" }) as HTMLButtonElement).disabled,
    ).toBe(true);
  });

  it("submits the selected option", async () => {
    const { getByLabelText, getByRole } = render(SipDecisionAlert, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: {
                currentDecision: {
                  message: "Choose how to continue",
                  options: ["Continue", "Cancel"],
                },
              },
            },
          }),
        ],
      },
    });
    const sipStore = useSipStore();

    await fireEvent.click(getByLabelText("Cancel"));
    await fireEvent.click(getByRole("button", { name: "Submit" }));

    expect(sipStore.submitDecision).toHaveBeenCalledWith("Cancel");
  });

  it("shows a submission error", () => {
    const { getByRole, getByText } = render(SipDecisionAlert, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: {
                currentDecision: {
                  message: "Choose how to continue",
                  options: ["Continue", "Cancel"],
                },
                currentDecisionError: "Decision expired",
              },
            },
          }),
        ],
      },
    });

    getByRole("button", { name: "Submit" });
    getByText("Decision expired");
  });

  it("disables submit while submission is in progress", () => {
    const { getByRole } = render(SipDecisionAlert, {
      global: {
        plugins: [
          createTestingPinia({
            createSpy: vi.fn,
            initialState: {
              sip: {
                currentDecision: {
                  message: "Choose how to continue",
                  options: ["Continue", "Cancel"],
                },
                submittingDecision: true,
              },
            },
          }),
        ],
      },
    });

    expect(
      (getByRole("button", { name: "Submitting..." }) as HTMLButtonElement)
        .disabled,
    ).toBe(true);
  });
});
