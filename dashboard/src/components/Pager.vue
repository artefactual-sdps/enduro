<script setup lang="ts">
import { computed } from "vue";

import IconCaretLeft from "~icons/bi/caret-left-fill";
import IconCaretRight from "~icons/bi/caret-right-fill";
import IconSkipEnd from "~icons/bi/skip-end-fill";
import IconSkipStart from "~icons/bi/skip-start-fill";

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
      <!-- First page -->
      <li class="page-item" :class="{ disabled: currentPage === 1 }">
        <a
          id="first-page"
          class="page-link"
          href="#"
          @click.prevent="goToPage(1)"
          aria-label="Go to first page"
          title="First page"
        >
          <IconSkipStart />
        </a>
      </li>

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
          <IconCaretLeft />
        </a>
      </li>

      <!-- Ellipses -->
      <li
        v-if="visiblePages[0] > 1"
        class="d-none d-sm-block"
        aria-hidden="true"
      >
        <a href="#" class="page-link disabled">…</a>
      </li>

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

      <!-- Ellipses -->
      <li
        v-if="visiblePages[visiblePages.length - 1] < totalPages"
        class="d-none d-sm-block"
        aria-hidden="true"
      >
        <a href="#" class="page-link disabled">…</a>
      </li>

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
          <IconCaretRight />
        </a>
      </li>

      <!-- Last page -->
      <li class="page-item" :class="{ disabled: currentPage === totalPages }">
        <a
          id="last-page"
          class="page-link"
          href="#"
          @click.prevent="goToPage(totalPages)"
          aria-label="Go to last page"
          title="Last page"
        >
          <IconSkipEnd />
        </a>
      </li>
    </ul>
  </nav>
</template>
