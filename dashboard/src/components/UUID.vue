<script setup lang="ts">
import { useClipboard } from "@vueuse/core";
import { toRef } from "vue";
import IconCheck from "~icons/akar-icons/check";
import IconCopy from "~icons/akar-icons/copy";

const props = defineProps<{ id?: string }>();

// $toRef can't be used because of https://github.com/vuejs/core/issues/6349.
const source = toRef(props, "id", "");

const { copy, copied, isSupported } = useClipboard({ source });
</script>

<template>
  <div v-if="id">
    <span class="font-monospace">{{ id }}</span>

    <template v-if="isSupported">
      <button
        @click="copy()"
        class="btn btn-sm btn-link link-secondary p-0 ms-2"
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
