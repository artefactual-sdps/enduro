<script setup lang="ts">
import Uppy from "@uppy/core";
import "@uppy/core/css/style.css";
import "@uppy/dashboard/css/style.css";
import Dashboard from "@uppy/vue/dashboard";
import XHR from "@uppy/xhr-upload";
import { computed, onBeforeUnmount, onMounted, watch } from "vue";
import { useRouter } from "vue-router";

import { getPath } from "@/client";
import { useTheme } from "@/composables/useTheme";
import { useAboutStore } from "@/stores/about";
import { useAuthStore } from "@/stores/auth";

const authStore = useAuthStore();
const aboutStore = useAboutStore();
const router = useRouter();
const { theme } = useTheme();

const gib = 1024 ** 3; // 1 GiB in bytes
const uploadMaxDefault = 4 * gib;
const dashboardOptions = computed(() => ({
  theme: theme.value,
  width: "100%",
}));

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

watch(theme, (value) => {
  uppy.getPlugin("Dashboard")?.setOptions({ theme: value });
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
  <Dashboard :uppy="uppy" :props="dashboardOptions" />
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
  background-color: var(--bs-tertiary-bg);
}

:deep(.uppy-Dashboard-inner) {
  border: var(--bs-border-width) solid var(--bs-border-color);
}

:deep(.uppy-Dashboard-AddFiles),
:deep(.uppy-DashboardContent-bar),
:deep(.uppy-StatusBar:not([aria-hidden="true"]).is-waiting),
:deep(.uppy-Dashboard-Item) {
  border-color: var(--bs-border-color);
}

:deep(.uppy-StatusBar.is-waiting .uppy-StatusBar-actions),
:deep(.uppy-Dashboard-dropFilesHereHint) {
  background-color: var(--bs-tertiary-bg);
}

:deep(.uppy-Dashboard-AddFilesPanel) {
  background: linear-gradient(
    0deg,
    var(--bs-tertiary-bg) 35%,
    rgba(var(--bs-tertiary-bg-rgb), 0.85) 100%
  );
}

:deep(.uppy-StatusBar.is-waiting .uppy-StatusBar-actionBtn--upload) {
  color: var(--bs-white);
  background-color: var(--bs-primary);
  border-radius: var(--bs-border-radius);
}

:deep(
  .uppy-StatusBar.is-waiting
    .uppy-StatusBar-actionBtn--upload:not(:disabled):hover
),
:deep(
  .uppy-StatusBar.is-waiting .uppy-StatusBar-actionBtn--upload:focus-visible
) {
  color: var(--bs-white);
  background-color: color-mix(in srgb, var(--bs-primary) 85%, black);
}

:deep(.uppy-StatusBar.is-waiting .uppy-StatusBar-actionBtn--upload:focus) {
  outline: 0;
  box-shadow: 0 0 0 var(--bs-focus-ring-width)
    color-mix(
      in srgb,
      color-mix(in srgb, var(--bs-primary) 85%, white) 50%,
      transparent
    );
}

:deep(
  .uppy-StatusBar.is-waiting .uppy-StatusBar-actionBtn--upload:disabled:hover
) {
  color: var(--bs-white);
  background-color: var(--bs-primary);
}
</style>
