import Modal from "bootstrap/js/dist/modal";
import { onBeforeUnmount, onMounted, shallowRef } from "vue";

import useEventListener from "@/composables/useEventListener";

// Own the Bootstrap instance for a dialog and report completion after
// Bootstrap has finished hiding it.
export default function useBootstrapModal(onHidden: () => void) {
  const element = shallowRef<HTMLElement | null>(null);
  const modal = shallowRef<Modal | null>(null);

  useEventListener(element, "hidden.bs.modal", onHidden);

  onMounted(() => {
    if (!element.value) return;

    modal.value = new Modal(element.value);
    modal.value.show();
  });

  onBeforeUnmount(() => {
    // dispose() alone does not undo the body scroll lock installed by show().
    // Dialogs omit Bootstrap's .fade class so hide() completes before disposal.
    modal.value?.hide();
    modal.value?.dispose();
  });

  return {
    element,
    hide: () => modal.value?.hide(),
  };
}
