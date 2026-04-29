<script setup lang="ts">
import { ref } from "vue";

import { useSipStore } from "@/stores/sip";

const sipStore = useSipStore();
const selectedOption = ref("");

const submitDecision = async () => {
  if (!selectedOption.value) return;
  await sipStore.submitDecision(selectedOption.value);
};
</script>

<template>
  <div v-if="sipStore.currentDecision" class="alert alert-info" role="alert">
    <h4 class="alert-heading">Task: User decision required</h4>
    <p>{{ sipStore.currentDecision.message }}</p>
    <hr />
    <form @submit.prevent="submitDecision">
      <div
        v-for="(option, index) in sipStore.currentDecision.options"
        :key="option"
        class="form-check"
      >
        <input
          :id="`sip-decision-option-${index}`"
          v-model="selectedOption"
          class="form-check-input"
          name="sip-decision-option"
          type="radio"
          :value="option"
        />
        <label class="form-check-label" :for="`sip-decision-option-${index}`">
          {{ option }}
        </label>
      </div>

      <button
        class="btn btn-primary mt-3"
        :disabled="!selectedOption || sipStore.submittingDecision"
        type="submit"
      >
        <span
          v-if="sipStore.submittingDecision"
          class="spinner-border spinner-border-sm me-2"
          role="status"
          aria-hidden="true"
        />
        {{ sipStore.submittingDecision ? "Submitting..." : "Submit" }}
      </button>
      <div
        v-if="sipStore.currentDecisionError"
        class="alert alert-danger mt-3 mb-0"
        role="alert"
      >
        {{ sipStore.currentDecisionError }}
      </div>
    </form>
  </div>
</template>
