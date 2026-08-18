import { enableAutoUnmount } from "@vue/test-utils";
import { Storage } from "happy-dom";
import { afterEach } from "vitest";

// Set a fixed local timezone for stable test output.
// America/Regina stays at UTC-06:00 year-round, avoiding DST-related changes.
process.env.TZ = "America/Regina";

// Node 26 defines Web Storage globals that are unavailable without a backing
// file, and Vitest does not replace them when it populates the happy-dom
// environment. Provide isolated browser storage for each test worker.
for (const name of ["localStorage", "sessionStorage"] as const) {
  Object.defineProperty(globalThis, name, {
    configurable: true,
    enumerable: true,
    value: new Storage(),
    writable: true,
  });
}

enableAutoUnmount(afterEach);
