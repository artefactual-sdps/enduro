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
      <div class="d-inline position-relative ms-2">
        <Transition name="slide-up" :duration="200">
          <!-- Copied visual hint. -->
          <span
            v-if="copied"
            class="btn btn-sm position-absolute border-0 text-success"
          >
            <IconCheck aria-hidden="true" />
            <span class="visually-hidden">Copied!</span>
          </span>

          <!-- Copy button. -->
          <button
            v-else
            @click="copy()"
            class="btn btn-sm position-absolute border-0 link-secondary"
          >
            <IconCopy aria-hidden="true" />
            <span class="visually-hidden">Copy to clipboard</span>
          </button>
        </Transition>
      </div>
    </template>
  </div>
</template>

<style scoped>
.btn-sm {
  padding: 0;
}

.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.2s ease-out;
}

.slide-up-enter-from {
  opacity: 0;
  transform: translateY(10px);
}

.slide-up-leave-to {
  opacity: 0;
  transform: translateY(-10px);
}
</style>
