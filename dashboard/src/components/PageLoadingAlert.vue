<script setup lang="ts">
import { computed } from "vue";

import type { runtime } from "@/client";

interface Props {
  title?: string;
  error?: unknown;
  execute?: (delay?: number, ...args: any[]) => Promise<unknown>;
}

const { title = "Page loading error", error, execute } = defineProps<Props>();

const retry = () => {
  if (execute) execute();
};

const is404 = computed(() => {
  let nf = false;
  try {
    const err = error as runtime.ResponseError;
    nf = err.response.status === 404;
  } catch (err) {}
  return nf;
});
</script>

<template>
  <!-- Not found. -->
  <div class="alert alert-warning" role="alert" v-if="error && is404">
    <h4 class="alert-heading">Page not found!</h4>
    <p>We can't find the page you're looking for.</p>
    <hr />
    <router-link class="btn btn-warning" :to="{ name: '/' }"
      >Take me home</router-link
    >
  </div>

  <!-- Other errors. -->
  <div class="alert alert-danger" role="alert" v-if="error && !is404">
    <h4 class="alert-heading">{{ title }}</h4>
    <slot>
      <p>It was not possible to load this page.</p>
    </slot>
    <pre v-if="error" class="mb-0 p-2 rounded bg-light">{{ error }}</pre>
    <template v-if="execute">
      <hr />
      <button class="btn btn-danger" @click="retry">Retry</button>
    </template>
  </div>
</template>
