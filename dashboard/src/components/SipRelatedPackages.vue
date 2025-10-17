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
    v-if="
      sipStore.current?.aipUuid ||
      (sipStore.current?.failedAs && sipStore.current?.failedKey)
    "
    class="card mb-3"
  >
    <div class="card-body">
      <h4 class="card-title">Related Packages</h4>
      <template v-if="sipStore.current?.aipUuid">
        <p class="card-text">
          <strong>AIP</strong>
          <UUID :id="sipStore.current.aipUuid" />
        </p>
        <RouterLink
          v-if="authStore.checkAttributes(['storage:aips:read'])"
          class="btn btn-primary btn-sm"
          :to="{
            name: '/storage/aips/[id]/',
            params: { id: sipStore.current.aipUuid },
          }"
        >
          View
        </RouterLink>
      </template>
      <template
        v-else-if="sipStore.current?.failedAs && sipStore.current?.failedKey"
      >
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
          v-if="authStore.checkAttributes(['ingest:sips:download'])"
          mode="out-in"
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
      </template>
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
