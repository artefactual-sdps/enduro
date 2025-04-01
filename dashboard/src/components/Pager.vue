<script setup lang="ts">
import { computed } from "vue";

import IconPrev from "~icons/bi/caret-left-fill";
import IconNext from "~icons/bi/caret-right-fill";

const props = defineProps({
  currentPage: {
    type: Number,
    required: true,
  },
  totalPages: {
    type: Number,
    required: true,
  },
  maxVisiblePages: {
    type: Number,
    default: 7, // An odd number of pages makes the pager symmetrical.
  },
});

const emit = defineEmits<{
  (e: "page-change", page: number): void;
}>();

const visiblePages = computed(() => {
  if (props.totalPages <= props.maxVisiblePages) {
    // If total pages is less than max visible, show all pages
    return Array.from({ length: props.totalPages }, (_, i) => i + 1);
  }

  // Calculate range of pages to show
  const halfVisible = Math.floor(props.maxVisiblePages / 2);
  let startPage = Math.max(props.currentPage - halfVisible, 1);
  let endPage = startPage + props.maxVisiblePages - 1;

  // Adjust if end page exceeds total pages
  if (endPage > props.totalPages) {
    endPage = props.totalPages;
    startPage = Math.max(endPage - props.maxVisiblePages + 1, 1);
  }

  return Array.from(
    { length: endPage - startPage + 1 },
    (_, i) => startPage + i,
  );
});

const goToPage = (page: number) => {
  if (page < 1 || page > props.totalPages || page === props.currentPage) {
    return;
  }
  emit("page-change", page);
};
</script>

<template>
  <nav role="navigation" aria-label="Pagination navigation">
    <ul class="pagination justify-content-center">
      <!-- Previous page -->
      <li class="page-item" :class="{ disabled: currentPage === 1 }">
        <a
          id="prev-page"
          class="page-link"
          href="#"
          @click.prevent="goToPage(currentPage - 1)"
          aria-label="Go to previous page"
          title="Previous page"
        >
          <IconPrev />
        </a>
      </li>

      <!-- First page and ellipses-->
      <template v-if="visiblePages[0] > 1">
        <li class="page-item" :class="{ disabled: currentPage === 1 }">
          <a
            id="first-page"
            class="page-link"
            href="#"
            @click.prevent="goToPage(1)"
            >1</a
          >
        </li>
        <li class="d-none d-sm-block" aria-hidden="true">
          <a href="#" class="page-link disabled">…</a>
        </li>
      </template>

      <!-- Page numbers -->
      <li
        v-for="page in visiblePages"
        :key="page"
        class="page-item"
        :class="{ active: currentPage === page }"
      >
        <a
          :id="`page-${page}`"
          class="page-link"
          href="#"
          @click.prevent="goToPage(page)"
          >{{ page }}</a
        >
      </li>

      <!-- Last page and ellipses-->
      <template v-if="visiblePages[visiblePages.length - 1] < totalPages">
        <li class="d-none d-sm-block" aria-hidden="true">
          <a href="#" class="page-link disabled">…</a>
        </li>
        <li class="page-item" :class="{ disabled: currentPage === totalPages }">
          <a
            id="last-page"
            class="page-link"
            href="#"
            @click.prevent="goToPage(totalPages)"
            >{{ totalPages }}</a
          >
        </li>
      </template>

      <!-- Next page -->
      <li class="page-item" :class="{ disabled: currentPage === totalPages }">
        <a
          id="next-page"
          class="page-link"
          href="#"
          @click.prevent="goToPage(currentPage + 1)"
          aria-label="Go to next page"
          title="Next page"
        >
          <IconNext />
        </a>
      </li>
    </ul>
  </nav>
</template>
