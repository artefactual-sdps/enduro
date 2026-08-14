import { createGlobalState, useColorMode } from "@vueuse/core";
import { computed } from "vue";

// Create the theme state once and share it across all consumers.
export const useTheme = createGlobalState(() => {
  const colorMode = useColorMode({
    attribute: "data-bs-theme",
    storageKey: "enduro-theme",
  });
  const isDark = computed(() => colorMode.state.value === "dark");

  const toggle = () => {
    colorMode.value = isDark.value ? "light" : "dark";
  };

  return {
    isDark,
    theme: colorMode.state,
    toggle,
  };
});
