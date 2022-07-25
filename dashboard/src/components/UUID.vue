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
      <button @click="copy()" class="btn btn-sm btn-link link-secondary ms-2">
        <Transition name="slide-up" :duration="200">
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
        </Transition>
      </button>
    </template>
  </div>
</template>

<style scoped lang="scss">
button {
  position: relative;
  width: $spacer * 1.2;
  height: $spacer * 1.5;
  overflow: hidden;

  span {
    position: absolute;
    top: 0;
    left: 0;
  }
}

.slide-up-enter-active,
.slide-up-leave-active {
  transition: all 0.2s ease-out;
}

.slide-up-enter-from {
  opacity: 0;
  transform: translateY($spacer);
}

.slide-up-leave-to {
  opacity: 0;
  transform: translateY(-$spacer);
}
</style>
