<script setup lang="ts">
import { onMounted, ref } from "vue";

import EmailLinkedText from "@/components/EmailLinkedText.vue";
import { useAipStore } from "@/stores/aip";
import { useAuthStore } from "@/stores/auth";

const aipStore = useAipStore();
const authStore = useAuthStore();

const { note } = defineProps<{
  note: string;
}>();

const canCancel = ref(false);

const cancel = async () => {
  aipStore.cancelDeletionRequest();
};

const review = async (approved: boolean) => {
  // TODO: Add confirmation dialog.
  aipStore.reviewDeletion(approved);
};

onMounted(() => {
  aipStore.canCancelDeletion().then((result) => {
    canCancel.value = result;
  });
});
</script>

<template>
  <div v-if="aipStore.isPending" class="alert alert-info" role="alert">
    <h4 class="alert-heading">Task: Review AIP deletion request</h4>
    <p class="line-break">
      <EmailLinkedText :text="note" />
    </p>
    <div class="d-flex flex-wrap gap-2">
      <template v-if="canCancel">
        <hr />
        <button class="btn btn-info" type="button" @click="cancel()">
          Cancel
        </button>
      </template>
      <template
        v-else-if="authStore.checkAttributes(['storage:aips:deletion:review'])"
      >
        <hr />
        <button class="btn btn-success" type="button" @click="review(true)">
          Approve
        </button>
        <button class="btn btn-danger" type="button" @click="review(false)">
          Reject
        </button>
      </template>
    </div>
  </div>
</template>
