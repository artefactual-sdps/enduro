<script setup lang="ts">
interface Props {
  title?: string;
  error?: unknown;
  execute?: (delay?: number, ...args: any[]) => Promise<unknown>;
}

const { title = "Page loading error", error, execute } = defineProps<Props>();

const retry = () => {
  if (execute) execute();
};
</script>

<template>
  <div class="alert alert-danger" role="alert" v-if="error">
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
