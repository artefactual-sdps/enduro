import { nextTick } from "vue";

import { themeController } from "@/theme";

export async function setTheme(theme: "light" | "dark") {
  if (themeController.theme.value === theme) return;
  themeController.toggle();
  await nextTick();
}
