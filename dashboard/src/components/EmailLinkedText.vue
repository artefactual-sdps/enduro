<script setup lang="ts">
const props = defineProps<{
  text: string;
}>();

const emailPattern = "[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\\.[a-zA-Z]{2,}";
const splitRegex = new RegExp(`(${emailPattern})`, "g");
const testRegex = new RegExp(`^${emailPattern}$`);

// Split text into alternating parts: text / email / text / email...
const parts = props.text.split(splitRegex);

function isEmail(part: string): boolean {
  return testRegex.test(part);
}
</script>

<template>
  <span>
    <template v-for="(part, index) in parts" :key="index">
      <template v-if="isEmail(part)">
        <a :href="'mailto:' + part">{{ part }}</a>
      </template>
      <template v-else>
        {{ part }}
      </template>
    </template>
  </span>
</template>
