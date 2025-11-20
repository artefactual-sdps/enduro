<script setup lang="ts">
defineProps<{
  logo: string;
  name: string;
  url?: string;
}>();
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
