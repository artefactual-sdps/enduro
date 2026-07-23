<script setup lang="ts">
import { onBeforeUnmount } from "vue";

import type { DialogRequest } from "@/dialogs/dialog";

const props = defineProps<{
  request: DialogRequest;
  resolve: (value: unknown) => void;
}>();

// Resolving removes this instance and runs its unmount hook too. The guard
// keeps a normal result from being replaced by the cancellation value.
let settled = false;

const settle = (value: unknown) => {
  if (settled) return;

  settled = true;
  props.resolve(value);
};

// Host teardown, such as an invalidated login, must not leave callers waiting
// on a promise that can no longer be completed by the dialog component.
onBeforeUnmount(() => {
  settle(props.request.cancelValue);
});
</script>

<template>
  <Component :is="request.component" v-bind="request.props" @resolve="settle" />
</template>
