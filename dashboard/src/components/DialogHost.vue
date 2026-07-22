<script setup lang="ts">
import DialogInstance from "@/components/DialogInstance.vue";
import { dialogTemplate } from "@/dialogs/dialog";
</script>

<template>
  <Component :is="dialogTemplate" v-slot="{ args, resolve }">
    <DialogInstance :request="args[0]" :resolve="resolve" />
  </Component>
</template>
