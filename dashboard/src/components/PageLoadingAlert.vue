<script setup lang="ts">
import { computed } from "vue";

import type { runtime } from "@/client";

interface Props {
  title?: string;
  error?: unknown;
  execute?: ((delay?: number, ...args: unknown[]) => Promise<unknown>) | null;
}

const {
  title = "Page loading error",
  error = undefined,
  execute = null,
} = defineProps<Props>();

const retry = () => {
  if (execute) execute();
};

const is404 = computed(() => {
  let nf = false;
  try {
    const err = error as runtime.ResponseError;
    nf = err.response.status === 404;
  } catch {
    return false;
  }
  return nf;
});
</script>

<template>
  <!-- Not found. -->
  <div v-if="error && is404" class="alert alert-warning" role="alert">
    <h4 class="alert-heading">Page not found!</h4>
    <p>We can't find the page you're looking for.</p>
    <hr />
    <RouterLink class="btn btn-warning" :to="{ name: '/' }">
      Take me home
    </RouterLink>
  </div>

  <!-- Other errors. -->
  <div v-if="error && !is404" class="alert alert-danger" role="alert">
    <h4 class="alert-heading">
      {{ title }}
    </h4>
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
