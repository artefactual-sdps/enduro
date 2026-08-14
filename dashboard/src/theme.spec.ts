import { afterEach, describe, expect, it, vi } from "vitest";
import { nextTick } from "vue";

class FakeColorScheme {
  matches: boolean;
  readonly media = "(prefers-color-scheme: dark)";
  readonly addEventListener = vi.fn(
    (type: string, listener: (event: MediaQueryListEvent) => void) => {
      if (type === "change") this.listeners.add(listener);
    },
  );
  private listeners = new Set<(event: MediaQueryListEvent) => void>();

  constructor(matches: boolean) {
    this.matches = matches;
  }

  setMatches(matches: boolean) {
    this.matches = matches;
    const event = { matches } as MediaQueryListEvent;
    this.listeners.forEach((listener) => listener(event));
  }
}

function createMemoryStorage(initialTheme?: string): Storage {
  const values = new Map<string, string>();
  if (initialTheme !== undefined) values.set("enduro-theme", initialTheme);

  const storage: Storage = {
    get length() {
      return values.size;
    },
    clear: () => {
      values.clear();
    },
    getItem: (key) => {
      return values.get(key) ?? null;
    },
    key: (index) => {
      return Array.from(values.keys())[index] ?? null;
    },
    removeItem: (key) => {
      values.delete(key);
    },
    setItem: (key, value) => {
      values.set(key, value);
    },
  };

  return Object.setPrototypeOf(storage, Storage.prototype);
}

async function loadTheme({
  storedTheme,
  systemDark = false,
}: {
  storedTheme?: string;
  systemDark?: boolean;
} = {}) {
  vi.resetModules();

  const colorScheme = new FakeColorScheme(systemDark);
  const storage = createMemoryStorage(storedTheme);
  vi.stubGlobal(
    "matchMedia",
    vi.fn(() => colorScheme as unknown as MediaQueryList),
  );
  vi.stubGlobal("localStorage", storage);

  const themeModule = await import("@/theme");

  return { colorScheme, storage, themeModule };
}

describe("theme controller", () => {
  afterEach(() => {
    delete document.documentElement.dataset.bsTheme;
    vi.unstubAllGlobals();
    vi.resetModules();
  });

  it("stores and uses the automatic theme until the user makes a choice", async () => {
    const { colorScheme, storage, themeModule } = await loadTheme({
      systemDark: true,
    });

    expect(themeModule.themeController.theme.value).toBe("dark");
    expect(document.documentElement.dataset.bsTheme).toBe("dark");
    expect(storage.getItem("enduro-theme")).toBe("auto");

    colorScheme.setMatches(false);
    await nextTick();

    expect(themeModule.themeController.theme.value).toBe("light");
    expect(document.documentElement.dataset.bsTheme).toBe("light");
  });

  it("persists a toggle and ignores later system changes", async () => {
    const { colorScheme, storage, themeModule } = await loadTheme();

    themeModule.themeController.toggle();
    await nextTick();

    expect(themeModule.themeController.theme.value).toBe("dark");
    expect(document.documentElement.dataset.bsTheme).toBe("dark");
    expect(storage.getItem("enduro-theme")).toBe("dark");

    colorScheme.setMatches(true);
    colorScheme.setMatches(false);
    await nextTick();

    expect(themeModule.themeController.theme.value).toBe("dark");
    expect(document.documentElement.dataset.bsTheme).toBe("dark");
  });

  it.each(["light", "dark"] as const)(
    "restores the stored %s theme instead of the system theme",
    async (storedTheme) => {
      const { themeModule } = await loadTheme({
        storedTheme,
        systemDark: storedTheme !== "dark",
      });

      expect(themeModule.themeController.theme.value).toBe(storedTheme);
      expect(document.documentElement.dataset.bsTheme).toBe(storedTheme);
    },
  );

  it("syncs theme changes from another tab", async () => {
    const { storage, themeModule } = await loadTheme();

    storage.setItem("enduro-theme", "dark");
    window.dispatchEvent(
      new StorageEvent("storage", {
        key: "enduro-theme",
        newValue: "dark",
        storageArea: storage,
      }),
    );
    await nextTick();

    expect(themeModule.themeController.theme.value).toBe("dark");
    expect(document.documentElement.dataset.bsTheme).toBe("dark");
  });
});
