<script setup lang="ts">
import { useAipStore } from "@/stores/aip";

const aipStore = useAipStore();

const { note } = defineProps<{
  note: string;
}>();

const review = async (approved: boolean) => {
  // TODO: Add confirmation dialog.
  aipStore.reviewDeletion(approved);
};
</script>

<template>
  <div class="alert alert-info" role="alert" v-if="aipStore.isPending">
    <h4 class="alert-heading">Task: Review Delete AIP request</h4>
    <p class="line-break">{{ note }}</p>
    <hr />
    <div class="d-flex flex-wrap gap-2">
      <button class="btn btn-success" type="button" @click="review(true)">
        Approve
      </button>
      <button class="btn btn-danger" type="button" @click="review(false)">
        Reject
      </button>
    </div>
  </div>
</template>
