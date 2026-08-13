import { readonly, ref } from "vue";

type Theme = "light" | "dark";

const colorScheme = window.matchMedia("(prefers-color-scheme: dark)");
const theme = ref<Theme>(colorScheme.matches ? "dark" : "light");
let hasExplicitPreference = false;
let isListening = false;

const applyTheme = (value: Theme) => {
  theme.value = value;
  document.documentElement.dataset.bsTheme = value;
};

const initialize = () => {
  let storedTheme: Theme | null = null;
  try {
    const value = localStorage.getItem("enduro-theme");
    if (value === "light" || value === "dark") storedTheme = value;
  } catch {
    // Use the system theme when storage is blocked.
  }

  hasExplicitPreference = storedTheme !== null;
  applyTheme(storedTheme ?? (colorScheme.matches ? "dark" : "light"));

  if (isListening) return;

  colorScheme.addEventListener("change", ({ matches }) => {
    if (!hasExplicitPreference) applyTheme(matches ? "dark" : "light");
  });
  isListening = true;
};

const toggle = () => {
  const nextTheme = theme.value === "light" ? "dark" : "light";
  hasExplicitPreference = true;

  try {
    localStorage.setItem("enduro-theme", nextTheme);
  } catch {
    // Keep the theme change for the current session when storage is blocked.
  }

  applyTheme(nextTheme);
};

export const themeController = {
  theme: readonly(theme),
  initialize,
  toggle,
};
