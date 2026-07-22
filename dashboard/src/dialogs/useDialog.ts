import Modal from "bootstrap/js/dist/modal";
import { onBeforeUnmount, onMounted, shallowRef } from "vue";

import useEventListener from "@/composables/useEventListener";

type ResolveArguments = [] | [result: unknown];

/**
 * Connects a dialog component's typed `resolve` event to its Bootstrap modal.
 *
 * Pass the component's typed emitter and its normal cancellation result, such
 * as `false` or `null`. If Bootstrap dismisses the modal directly, the
 * arguments passed after `emit` are emitted as that default result. For a
 * resultless dialog, declare `resolve: []` and omit the default.
 *
 * Bind the returned `element` ref to the modal root. Calling `close(result)`
 * records a result and starts hiding the modal. The composable emits `resolve`
 * only after Bootstrap finishes hiding, so action handlers must call `close`
 * rather than emit `resolve` directly.
 *
 * The composable owns the Bootstrap instance and hides and disposes it when the
 * component unmounts. The modal root must omit Bootstrap's `.fade` class so
 * forced teardown remains synchronous.
 *
 * @example
 * ```ts
 * const emit = defineEmits<{
 *   resolve: [confirmed: boolean];
 * }>();
 *
 * const { element, close } = useDialog(emit, false);
 * ```
 */
export default function useDialog<Arguments extends ResolveArguments>(
  emit: (event: "resolve", ...args: Arguments) => void,
  ...defaultArguments: NoInfer<Arguments>
) {
  const element = shallowRef<HTMLElement | null>(null);
  const modal = shallowRef<Modal | null>(null);
  let resultArguments = defaultArguments;

  useEventListener(element, "hidden.bs.modal", () => {
    emit("resolve", ...resultArguments);
    resultArguments = defaultArguments;
  });

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
    close: (...args: Arguments) => {
      resultArguments = args;
      modal.value?.hide();
    },
  };
}
