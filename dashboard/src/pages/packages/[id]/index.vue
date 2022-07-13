<script setup lang="ts">
import IconCircleChevronDown from "~icons/akar-icons/circle-chevron-down";
import IconCircleChevronUp from "~icons/akar-icons/circle-chevron-up";
import PackageDetailsCard from "@/components/PackageDetailsCard.vue";
import PackageLocationCard from "@/components/PackageLocationCard.vue";
import PackageReviewAlert from "@/components/PackageReviewAlert.vue";
import PackageStatusBadge from "@/components/PackageStatusBadge.vue";
import { ref, onMounted, watch } from "vue";
import useEventListener from "@/composables/useEventListener";
import { usePackageStore } from "@/stores/package";
import Collapse from "bootstrap/js/dist/collapse";

const packageStore = usePackageStore();

let shown = $ref<boolean>(false);

const el = ref<HTMLElement | null>(null);
useEventListener(el, "shown.bs.collapse", (e) => {
  shown = true;
  el.value?.scrollIntoView();
});
useEventListener(el, "hidden.bs.collapse", (e) => (shown = false));

let col = <Collapse | null>null;
onMounted(() => {
  if (!el.value) return;
  col = new Collapse(el.value, { toggle: false });
});

const expandAll = () => col?.show();
const collapseAll = () => col?.hide();

watch(packageStore.ui.expand, () => col?.show());
</script>

<template>
  <div v-if="packageStore.current">
    <div class="row">
      <div class="col-md-6">
        <h2>AIP creation details</h2>
        <dl>
          <dt>Name</dt>
          <dd>{{ packageStore.current.name }}</dd>
          <dt>AIP UUID</dt>
          <dd>{{ packageStore.current.aipId }}</dd>
          <dt>Workflow status</dt>
          <dd>
            <PackageStatusBadge
              :status="packageStore.current.status"
              :note="'Create and Review AIP'"
            />
          </dd>
          <dt>Started</dt>
          <dd>{{ $filters.formatDateTime(packageStore.current.startedAt) }}</dd>
          <span v-if="packageStore.current.completedAt">
            <dt>Completed</dt>
            <dd>
              {{ $filters.formatDateTime(packageStore.current.completedAt) }}
              <div class="pt-2">
                (took
                {{
                  $filters.formatDuration(
                    packageStore.current.startedAt,
                    packageStore.current.completedAt
                  )
                }})
              </div>
            </dd>
          </span>
        </dl>
      </div>
      <div class="col-md-6">
        <PackageLocationCard />
        <PackageDetailsCard />
      </div>
    </div>

    <div class="d-flex">
      <h2 class="mb-0">Preservation actions</h2>
      <div
        class="align-self-end ms-auto d-flex"
        v-if="packageStore.current_preservation_actions?.actions"
      >
        <button
          class="btn btn-sm btn-link text-decoration-none p-0"
          type="button"
          @click="expandAll()"
        >
          Expand all
        </button>
        <span class="px-1 link-primary">|</span>
        <button
          class="btn btn-sm btn-link text-decoration-none p-0"
          type="button"
          @click="collapseAll()"
        >
          Collapse all
        </button>
      </div>
    </div>

    <hr />

    <div class="mb-3">
      <div class="d-flex">
        <h3>
          Create and Review AIP
          <PackageStatusBadge :status="packageStore.current.status" />
        </h3>
        <button
          class="btn btn-sm btn-link text-decoration-none ms-auto"
          type="button"
          data-bs-toggle="collapse"
          data-bs-target="#preservation-actions-table-0"
          aria-expanded="false"
          aria-controls="preservation-actions-table-0"
          v-if="packageStore.current_preservation_actions?.actions"
        >
          <IconCircleChevronUp style="font-size: 2em" v-if="shown" />
          <IconCircleChevronDown style="font-size: 2em" v-else />
        </button>
      </div>
      <span v-if="packageStore.current.completedAt">
        Completed
        {{ $filters.formatDateTime(packageStore.current.completedAt) }}
        (took
        {{
          $filters.formatDuration(
            packageStore.current.startedAt,
            packageStore.current.completedAt
          )
        }})
      </span>
      <span v-else>
        Started {{ $filters.formatDateTime(packageStore.current.startedAt) }}
      </span>
    </div>

    <PackageReviewAlert />

    <div ref="el" id="preservation-actions-table-0" class="collapse">
      <table
        class="table table-bordered table-sm"
        v-if="packageStore.current_preservation_actions?.actions"
      >
        <thead>
          <tr>
            <th scope="col">Task #</th>
            <th scope="col">Name</th>
            <th scope="col">Outcome</th>
            <th scope="col">Notes</th>
          </tr>
        </thead>
        <tbody>
          <tr
            v-for="(action, idx) in packageStore.current_preservation_actions
              .actions"
            :key="action.id"
          >
            <td>{{ idx + 1 }}</td>
            <td>{{ action.name }}</td>
            <td>
              <span
                class="badge"
                :class="$filters.formatPreservationActionStatus(action.status)"
                >{{ action.status }}</span
              >
            </td>
            <td>TODO: note goes here</td>
          </tr>
        </tbody>
      </table>
    </div>
  </div>
</template>
