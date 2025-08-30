<script setup lang="ts">
import { defineProps } from "vue";

const props = defineProps<{
  logo: string;
  name: string;
  url: string;
}>();

const imgTag = `<img
    id="institution-logo"
    src="${props.logo}"
    alt="${props.name}"
    height="30"
    
/>`;
</script>

<template>
  <div v-if="props.logo && props.url" class="d-none d-sm-block mx-3">
    <a :href="props.url" target="_blank" rel="external" v-html="imgTag"></a>
  </div>
  <div
    v-else-if="props.logo"
    v-html="imgTag"
    class="d-none d-sm-block mx-3"
  ></div>
</template>
