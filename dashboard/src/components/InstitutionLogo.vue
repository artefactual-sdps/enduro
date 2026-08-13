<script setup lang="ts">
import { computed } from "vue";

import { themeController } from "@/theme";

const logos = {
  light: import.meta.env.VITE_INSTITUTION_LOGO_LIGHT,
  dark: import.meta.env.VITE_INSTITUTION_LOGO_DARK,
  legacy: import.meta.env.VITE_INSTITUTION_LOGO,
};
const name = import.meta.env.VITE_INSTITUTION_NAME;
const url = import.meta.env.VITE_INSTITUTION_URL;
const logo = computed(() => logos[themeController.theme.value] || logos.legacy);
</script>

<template>
  <div v-if="logo" class="d-none d-sm-block mx-3">
    <a v-if="url" :href="url" target="_blank" rel="external">
      <img :src="logo" :alt="name" />
    </a>
    <img v-else :src="logo" :alt="name" />
  </div>
</template>

<style scoped>
img {
  max-height: 30px;
  height: auto;
  width: auto;
}
</style>
