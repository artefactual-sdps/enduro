<script setup lang="ts">
import { api } from "@/client";
import UUID from "@/components/UUID.vue";
import { useAuthStore } from "@/stores/auth";
import { useSipStore } from "@/stores/sip";

const authStore = useAuthStore();
const sipStore = useSipStore();
</script>

<template>
  <div
    class="card mb-3"
    v-if="
      sipStore.current?.aipId ||
      (sipStore.current?.failedAs && sipStore.current?.failedKey)
    "
  >
    <div class="card-body">
      <h4 class="card-title">Related Packages</h4>
      <template v-if="sipStore.current?.aipId">
        <p class="card-text">
          <strong>AIP</strong>
          <UUID :id="sipStore.current.aipId" />
        </p>
        <router-link
          v-if="authStore.checkAttributes(['storage:aips:read'])"
          class="btn btn-primary btn-sm"
          :to="{
            name: '/storage/aips/[id]/',
            params: { id: sipStore.current.aipId },
          }"
          >View</router-link
        >
      </template>
      <p class="card-text">
        <strong
          v-if="
            sipStore.current.failedAs == api.EnduroIngestSipFailedAsEnum.Sip
          "
          >Failed SIP</strong
        >
        <strong
          v-if="
            sipStore.current.failedAs == api.EnduroIngestSipFailedAsEnum.Pip
          "
          >Failed PIP</strong
        >
        <br />
        {{ sipStore.current.failedKey }}
      </p>
      <Transition
        mode="out-in"
        v-if="authStore.checkAttributes(['ingest:sips:download'])"
      >
        <div
          v-if="sipStore.downloadError"
          class="alert alert-danger text-center mb-0"
          role="alert"
        >
          {{ sipStore.downloadError }}
        </div>
        <button
          v-else
          type="button"
          class="btn btn-primary btn-sm"
          @click="sipStore.download()"
        >
          Download
        </button>
      </Transition>
    </div>
  </div>
</template>

<style scoped>
.v-enter-active,
.v-leave-active {
  transition: opacity 0.3s;
}
.v-enter-from,
.v-leave-to {
  opacity: 0;
}
.v-enter-to,
.v-leave-from {
  opacity: 1;
}
</style>
