import { enableAutoUnmount } from "@vue/test-utils";
import { afterEach } from "vitest";

// Set a fixed local timezone for stable test output.
// America/Regina stays at UTC-06:00 year-round, avoiding DST-related changes.
process.env.TZ = "America/Regina";

enableAutoUnmount(afterEach);
