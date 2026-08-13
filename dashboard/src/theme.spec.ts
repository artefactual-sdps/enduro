import { afterEach, describe, expect, it, vi } from "vitest";

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

function createMemoryStorage(initialTheme?: string, available = true): Storage {
  const values = new Map<string, string>();
  if (initialTheme !== undefined) values.set("enduro-theme", initialTheme);

  const ensureAvailable = () => {
    if (!available) throw new DOMException("Storage unavailable");
  };

  return {
    get length() {
      ensureAvailable();
      return values.size;
    },
    clear: () => {
      ensureAvailable();
      values.clear();
    },
    getItem: (key) => {
      ensureAvailable();
      return values.get(key) ?? null;
    },
    key: (index) => {
      ensureAvailable();
      return Array.from(values.keys())[index] ?? null;
    },
    removeItem: (key) => {
      ensureAvailable();
      values.delete(key);
    },
    setItem: (key, value) => {
      ensureAvailable();
      values.set(key, value);
    },
  };
}

async function loadTheme({
  storedTheme,
  systemDark = false,
  storageAvailable = true,
}: {
  storedTheme?: string;
  systemDark?: boolean;
  storageAvailable?: boolean;
} = {}) {
  vi.resetModules();

  const colorScheme = new FakeColorScheme(systemDark);
  const storage = createMemoryStorage(storedTheme, storageAvailable);
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

  it("exports only the theme controller", async () => {
    const { themeModule } = await loadTheme();

    expect(Object.keys(themeModule)).toEqual(["themeController"]);
  });

  it("uses the system theme until the user makes a choice", async () => {
    const { colorScheme, storage, themeModule } = await loadTheme({
      systemDark: true,
    });

    themeModule.themeController.initialize();

    expect(themeModule.themeController.theme.value).toBe("dark");
    expect(document.documentElement.dataset.bsTheme).toBe("dark");
    expect(storage.getItem("enduro-theme")).toBeNull();

    colorScheme.setMatches(false);

    expect(themeModule.themeController.theme.value).toBe("light");
    expect(document.documentElement.dataset.bsTheme).toBe("light");
  });

  it("persists a toggle and ignores later system changes", async () => {
    const { colorScheme, storage, themeModule } = await loadTheme();

    themeModule.themeController.initialize();
    themeModule.themeController.toggle();

    expect(themeModule.themeController.theme.value).toBe("dark");
    expect(document.documentElement.dataset.bsTheme).toBe("dark");
    expect(storage.getItem("enduro-theme")).toBe("dark");

    colorScheme.setMatches(true);
    colorScheme.setMatches(false);

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

      themeModule.themeController.initialize();

      expect(themeModule.themeController.theme.value).toBe(storedTheme);
      expect(document.documentElement.dataset.bsTheme).toBe(storedTheme);
    },
  );

  it("ignores an invalid stored theme", async () => {
    const { themeModule } = await loadTheme({
      storedTheme: "sepia",
      systemDark: true,
    });

    themeModule.themeController.initialize();

    expect(themeModule.themeController.theme.value).toBe("dark");
    expect(document.documentElement.dataset.bsTheme).toBe("dark");
  });

  it("keeps the session choice when storage is unavailable", async () => {
    const { colorScheme, themeModule } = await loadTheme({
      storageAvailable: false,
    });

    themeModule.themeController.initialize();
    themeModule.themeController.toggle();
    colorScheme.setMatches(false);

    expect(themeModule.themeController.theme.value).toBe("dark");
    expect(document.documentElement.dataset.bsTheme).toBe("dark");
  });

  it("registers the system theme listener only once", async () => {
    const { colorScheme, themeModule } = await loadTheme();

    themeModule.themeController.initialize();
    themeModule.themeController.initialize();

    expect(colorScheme.addEventListener).toHaveBeenCalledOnce();
  });
});
