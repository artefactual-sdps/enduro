import { useColorMode } from "@vueuse/core";

const colorMode = useColorMode({
  attribute: "data-bs-theme",
  storageKey: "enduro-theme",
});

export const themeController = {
  theme: colorMode.state,
  toggle: () => {
    colorMode.value = colorMode.state.value === "light" ? "dark" : "light";
  },
};
