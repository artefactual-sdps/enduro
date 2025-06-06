<script setup lang="ts">
import Modal from "bootstrap/js/dist/modal";
import { onMounted, ref } from "vue";
import { closeDialog } from "vue3-promise-dialog";

import useEventListener from "@/composables/useEventListener";
import { useAboutStore } from "@/stores/about";
import IconDocumentation from "~icons/lucide/book-text";
import IconLicense from "~icons/lucide/file-text";
import IconContributing from "~icons/lucide/git-merge";

const aboutStore = useAboutStore();

const el = ref<HTMLElement | null>(null);
const modal = ref<Modal | null>(null);

onMounted(() => {
  if (!el.value) return;
  modal.value = new Modal(el.value);
  modal.value.show();

  aboutStore.load();
});

useEventListener(el, "hidden.bs.modal", () => closeDialog(null));
</script>

<template>
  <div class="modal" tabindex="-1" ref="el">
    <div class="modal-dialog">
      <div class="modal-content">
        <div class="modal-header">
          <h5 class="modal-title fw-bold d-flex align-items-center">
            <img src="/logo.png" alt="" height="30" class="me-2" />Enduro
          </h5>
          <button
            type="button"
            class="btn-close"
            data-bs-dismiss="modal"
            aria-label="Close"
          ></button>
        </div>
        <div class="modal-body">
          <div class="mb-3">
            <div class="row">
              <div
                class="col-12 col-sm-6 text-primary fw-bold text-sm-end text-truncate"
              >
                Application version:
              </div>
              <div class="col-12 col-sm-6 text-truncate">
                {{ aboutStore.formattedVersion }}
              </div>
            </div>
            <div class="row">
              <div
                class="col-12 col-sm-6 text-primary fw-bold text-sm-end text-truncate"
              >
                Preservation system:
              </div>
              <div class="col-12 col-sm-6 text-truncate">
                {{ aboutStore.preservationSystem }}
              </div>
            </div>
            <div class="row" v-if="aboutStore.preprocessing.enabled">
              <div
                class="col-12 col-sm-6 text-primary fw-bold text-sm-end text-truncate"
              >
                Preprocessing workflow:
              </div>
              <div class="col-12 col-sm-6 text-truncate">
                {{ aboutStore.preprocessing.workflowName }}
              </div>
            </div>
            <div
              class="row"
              v-if="aboutStore.poststorage && aboutStore.poststorage.length > 0"
            >
              <div
                class="col-12 col-sm-6 text-primary fw-bold text-sm-end text-truncate"
              >
                Poststorage workflows:
              </div>
              <div class="col-12 col-sm-6 d-flex flex-column text-truncate">
                <span
                  v-for="ps in aboutStore.poststorage"
                  :key="ps.taskQueue + '-' + ps.workflowName"
                  >{{ ps.workflowName }}</span
                >
              </div>
            </div>
          </div>
          <div class="small">
            Enduro is a new application under development by
            <a href="https://www.artefactual.com/" target="_blank"
              >Artefactual Systems</a
            >. Originally created as a more stable replacement for
            Archivematica's
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
          </div>
        </div>
        <div class="modal-footer">
          <a
            class="btn btn-primary d-flex align-items-center gap-2"
            href="https://enduro.readthedocs.io/"
            target="_blank"
          >
            <IconDocumentation aria-hidden="true" />
            Documentation
          </a>
          <a
            class="btn btn-primary d-flex align-items-center gap-2"
            href="https://github.com/artefactual-sdps/enduro/blob/main/LICENSE"
            target="_blank"
          >
            <IconLicense aria-hidden="true" />
            License
          </a>
          <a
            class="btn btn-primary d-flex align-items-center gap-2"
            href="https://github.com/artefactual-sdps/enduro/blob/main/CONTRIBUTING.md"
            target="_blank"
          >
            <IconContributing aria-hidden="true" />
            Contributing
          </a>
        </div>
      </div>
    </div>
  </div>
</template>
