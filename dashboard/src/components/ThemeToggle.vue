<script setup lang="ts">
import { computed } from "vue";

import IconMoon from "~icons/clarity/moon-line";
import IconSun from "~icons/clarity/sun-line";

import { useTheme } from "@/composables/useTheme";

const props = withDefaults(defineProps<{ compact?: boolean }>(), {
  compact: false,
});
const { isDark, toggle } = useTheme();

const actionLabel = computed(() =>
  isDark.value ? "Light theme" : "Dark theme",
);
</script>

<template>
  <button
    type="button"
    :class="
      props.compact
        ? 'theme-toggle-compact btn d-inline-flex rounded-circle p-3'
        : 'dropdown-item d-flex align-items-center gap-3'
    "
    :aria-label="actionLabel"
    @click="toggle"
  >
    <IconMoon v-if="!isDark" aria-hidden="true" />
    <IconSun v-else aria-hidden="true" />
    <span v-if="!props.compact">{{ actionLabel }}</span>
  </button>
</template>

<style scoped>
.theme-toggle-compact {
  --bs-btn-hover-color: var(--bs-emphasis-color);
  --bs-btn-hover-bg: var(--bs-tertiary-bg);
  --bs-btn-active-color: var(--bs-emphasis-color);
  --bs-btn-active-bg: var(--bs-tertiary-bg);
  --bs-btn-focus-shadow-rgb: var(--bs-primary-rgb);
}
</style>
