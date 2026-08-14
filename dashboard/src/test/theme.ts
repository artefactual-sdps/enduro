import { nextTick } from "vue";

import { useTheme } from "@/composables/useTheme";

export async function setTheme(theme: "light" | "dark") {
  const currentTheme = useTheme();
  if (currentTheme.theme.value === theme) return;
  currentTheme.toggle();
  await nextTick();
}
