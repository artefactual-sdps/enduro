<script setup lang="ts">
/* eslint-disable vue/no-v-html -- Using DOMPurify. */
import DOMPurify from "dompurify";
import { computed } from "vue";

const props = withDefaults(defineProps<{ html?: string }>(), { html: "" });

const sanitized = computed(() => DOMPurify.sanitize(props.html));
</script>

<template>
  <div v-if="sanitized" v-html="sanitized" />
</template>
