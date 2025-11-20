<script setup lang="ts">
import { onMounted, ref } from "vue";

import { useAuthStore } from "@/stores/auth";
import { useLayoutStore } from "@/stores/layout";
import IconHome from "~icons/clarity/home-line";

const authStore = useAuthStore();
const layoutStore = useLayoutStore();
layoutStore.updateBreadcrumb([]);

const url = import.meta.env.VITE_CUSTOM_HOME_URL as string | undefined;
const content = ref<string | null>(null);
const error = ref<string | null>(null);
const loading = ref(false);

onMounted(async () => {
  if (!url) return;
  loading.value = true;
  try {
    const response = await fetch(url);
    if (!response.ok) throw new Error(`Response status: ${response.status}`);
    const data = await response.text();
    if (!data) throw new Error("Content is empty.");
    content.value = data;
  } catch (err) {
    console.error("Error loading custom home HTML:", err);
    error.value = "Failed to load custom home content.";
  }
  loading.value = false;
});
</script>

<template>
  <div class="container-xxl">
    <!-- Custom HTML content -->
    <template v-if="url">
      <div v-if="loading" class="text-center p-3">
        <div class="spinner-border text-muted" role="status">
          <span class="visually-hidden">Loading...</span>
        </div>
      </div>
      <div v-else-if="error" class="alert alert-warning" role="alert">
        {{ error }}
      </div>
      <SafeHtml v-else-if="content" :html="content" />
    </template>

    <!-- Default content -->
    <div v-if="!url || error">
      <h1 class="d-flex mb-3">
        <IconHome class="me-3 text-dark" />Welcome<span
          v-if="authStore.isEnabled"
          >, {{ authStore.getUserDisplayName }}</span
        >!
      </h1>
      <p>
        Enduro is a new application under development by
        <a href="https://www.artefactual.com/" target="_blank"
          >Artefactual Systems</a
        >. Originally created as a more stable replacement for Archivematica's
        <a
          href="https://github.com/artefactual/automation-tools"
          target="_blank"
          >automation-tools</a
        >
        library of scripts, it has since evolved into a flexible tool to be
        paired with preservation applications like
        <a href="https://www.archivematica.org/" target="_blank"
          >Archivematica</a
        >
        and
        <a href="https://github.com/artefactual-labs/a3m" target="_blank"
          >a3m</a
        >
        to provide initial ingest activities such as content and structure
        validation, packaging, and more.
      </p>
      <p>
        See the
        <a href="https://enduro.readthedocs.io/" target="_blank"
          >Enduro documentation</a
        >
        for more information!
      </p>
    </div>
  </div>
</template>
