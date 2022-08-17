<script setup lang="ts">
import { useClipboard } from "@vueuse/core";
import Tooltip from "bootstrap/js/dist/tooltip";
import { toRef, watch } from "vue";
import IconCheck from "~icons/akar-icons/check";
import IconCopy from "~icons/akar-icons/copy";

const props = defineProps<{ id?: string }>();

// $toRef can't be used because of https://github.com/vuejs/core/issues/6349.
const source = toRef(props, "id", "");

const { copy, copied, isSupported } = useClipboard({ source });

const el = $ref<HTMLElement | null>(null);
let tooltip: Tooltip | null = null;

watch($$(el), () => {
  if (el) tooltip = new Tooltip(el);
});

watch(copied, (val) => {
  if (tooltip) {
    tooltip.setContent({
      ".tooltip-inner": val ? "Copied!" : "Copy to clipboard",
    });
    if (!val) tooltip.hide();
  }
});
</script>

<template>
  <div v-if="id" class="d-flex align-items-start gap-2">
    <span class="font-monospace">{{ id }}</span>

    <template v-if="isSupported">
      <button
        ref="el"
        @click="copy()"
        class="btn btn-sm btn-link link-secondary p-0"
        data-bs-toggle="tooltip"
        data-bs-title="Copy to clipboard"
      >
        <!-- Copied visual hint. -->
        <span v-if="copied">
          <IconCheck aria-hidden="true" class="text-success" />
          <span class="visually-hidden">Copied!</span>
        </span>
        <!-- Copy icon. -->
        <span v-else>
          <IconCopy aria-hidden="true" />
          <span class="visually-hidden">Copy to clipboard</span>
        </span>
      </button>
    </template>
  </div>
</template>
