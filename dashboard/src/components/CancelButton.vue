<script setup lang="ts">
const cancel = () => {
  alert("YOU ARE DONE!");
};
</script>

<template>
  <div class="action">
    <button class="btn btn-warning" @click="cancel()">
      <slot></slot>
    </button>
  </div>
</template>
