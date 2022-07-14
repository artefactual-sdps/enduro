import { usePackageStore } from "../../src/stores/package";
import { setActivePinia, createPinia } from "pinia";
import { expect, describe, it, beforeEach } from "vitest";

describe("usePackageStore", () => {
  beforeEach(() => {
    setActivePinia(createPinia());
  });

  it("getActionById finds actions", () => {
    const packageStore = usePackageStore();
    packageStore.$patch((state) => {
      state.current_preservation_actions = {
        actions: [
          { id: 1, type: "create-aip" },
          { id: 2, type: "move-package" },
        ],
      };
    });

    expect(packageStore.getActionById(1)).toEqual({
      id: 1,
      type: "create-aip",
    });
    expect(packageStore.getActionById(2)).toEqual({
      id: 2,
      type: "move-package",
    });
    expect(packageStore.getActionById(3)).toBeUndefined();
  });

  it("getTaskById finds tasks", () => {
    const packageStore = usePackageStore();
    packageStore.$patch((state) => {
      state.current_preservation_actions = {
        actions: [
          { id: 1, type: "create-aip", tasks: [{ id: 1 }, { id: 2 }] },
          { id: 2, type: "move-package", tasks: [{ id: 3 }, { id: 4 }] },
        ],
      };
    });

    expect(packageStore.getTaskById(1, 1)).toEqual({ id: 1 });
    expect(packageStore.getTaskById(1, 2)).toEqual({ id: 2 });
    expect(packageStore.getTaskById(1, 3)).toBeUndefined();

    expect(packageStore.getTaskById(2, 3)).toEqual({ id: 3 });
    expect(packageStore.getTaskById(2, 4)).toEqual({ id: 4 });
    expect(packageStore.getTaskById(2, 5)).toBeUndefined();
  });
});
