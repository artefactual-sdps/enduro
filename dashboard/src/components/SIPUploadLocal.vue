<script setup lang="ts">
import Uppy from "@uppy/core";
import { Dashboard } from "@uppy/vue";
import XHR from "@uppy/xhr-upload";
import { onMounted } from "vue";
import { useRouter } from "vue-router/auto";

import { getPath } from "@/client";
import { useAboutStore } from "@/stores/about";
import { useAuthStore } from "@/stores/auth";

import "@uppy/core/dist/style.css";
import "@uppy/dashboard/dist/style.css";
import "@uppy/progress-bar/dist/style.css";

const authStore = useAuthStore();
const aboutStore = useAboutStore();
const router = useRouter();

const GiB = 1024 ** 3; // 1 GiB in bytes
const uploadMaxDefault = 4 * GiB;

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
</script>

<template>
  <div class="text-muted mb-3">
    SIPs <strong>must</strong> be zipped. No SIPs larger than
    {{ aboutStore.formattedUploadMaxSize }}. Ingest will start automatically.
  </div>
  <Dashboard :uppy="uppy" />
</template>
