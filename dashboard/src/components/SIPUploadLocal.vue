<script setup lang="ts">
import Uppy from "@uppy/core";
import "@uppy/core/css/style.css";
import "@uppy/dashboard/css/style.css";
import Dashboard from "@uppy/vue/dashboard";
import XHR from "@uppy/xhr-upload";
import { onBeforeUnmount, onMounted } from "vue";
import { useRouter } from "vue-router";

import { getPath } from "@/client";
import { useAboutStore } from "@/stores/about";
import { useAuthStore } from "@/stores/auth";

const authStore = useAuthStore();
const aboutStore = useAboutStore();
const router = useRouter();

const gib = 1024 ** 3; // 1 GiB in bytes
const uploadMaxDefault = 4 * gib;

aboutStore.$subscribe((_, state) => {
  uppy.setOptions({
    restrictions: { maxFileSize: state.uploadMaxSize },
  });
});

onMounted(() => {
  aboutStore.load();
});

const uppy = new Uppy({
  restrictions: { maxFileSize: uploadMaxDefault },
}).use(XHR, {
  endpoint: getPath() + "/ingest/sips/upload",
  allowedMetaFields: false,
  // Called again for every retry too.
  async onBeforeRequest(xhr) {
    if (!authStore.isUserValid) {
      await authStore.signinSilent();
    }
    xhr.setRequestHeader(
      "Authorization",
      `Bearer ${authStore.getUserAccessToken}`,
    );
  },
  async onAfterResponse(xhr) {
    switch (xhr.status) {
      // "202 Accepted" is returned on successful upload.
      case 202:
        setTimeout(() => {
          router.push({
            path: "/ingest/sips",
          });
        }, 500);
        break;
      // "401 Unauthorized" is returned when the auth token has expired.
      case 401:
        await authStore.signinSilent();
        break;
    }
  },
  getResponseData: () => {
    return { url: "" };
  },
});

onBeforeUnmount(() => {
  uppy.destroy();
});
</script>

<template>
  <div class="text-muted mb-3">
    SIPs <strong>must</strong> be zipped. No SIPs larger than
    {{ aboutStore.formattedUploadMaxSize }}. Ingest will start automatically.
  </div>
  <Dashboard :uppy="uppy" :props="{ width: '100%' }" />
</template>

<style scoped>
/* Uppy renders its internals inside a child component, so :deep() is required
 * to reach them from this scoped style block. */
:deep(.uppy-Root),
:deep(.uppy-u-reset) {
  font-family: var(--bs-body-font-family);
}

:deep(.uppy-Dashboard-inner),
:deep(.uppy-Dashboard-innerWrap),
:deep(.uppy-Dashboard-AddFiles) {
  border-radius: var(--bs-border-radius);
}

:deep(.uppy-Dashboard-inner),
:deep(.uppy-DashboardContent-bar),
:deep(.uppy-StatusBar) {
  background-color: var(--bs-light);
}

:deep(.uppy-StatusBar.is-waiting .uppy-StatusBar-actionBtn--upload) {
  background-color: var(--bs-primary);
  border-radius: var(--bs-border-radius);
}
</style>
