import { describe, expect, it, vi } from "vitest";

import { logError } from "@/helpers/logs";
import { ResponseError } from "@/openapi-generator";

describe("logError", () => {
  it("logs ResponseError with status and statusText", () => {
    const consoleErrorSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});
    const error = new ResponseError(
      new Response("Not Found", { status: 404, statusText: "Not Found" }),
    );
    logError(error, "Test error");
    expect(consoleErrorSpy).toHaveBeenCalledWith(
      "Test error:",
      404,
      "Not Found",
    );
    consoleErrorSpy.mockRestore();
  });

  it("logs Error with message", () => {
    const consoleErrorSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});
    const error = new Error("Something went wrong");
    logError(error, "Test error");
    expect(consoleErrorSpy).toHaveBeenCalledWith(
      "Test error",
      "Something went wrong",
    );
    consoleErrorSpy.mockRestore();
  });

  it("logs unknown error type", () => {
    const consoleErrorSpy = vi
      .spyOn(console, "error")
      .mockImplementation(() => {});
    logError("Unknown error", "Test error");
    expect(consoleErrorSpy).toHaveBeenCalledWith("Test error", "Unknown error");
    consoleErrorSpy.mockRestore();
  });
});
