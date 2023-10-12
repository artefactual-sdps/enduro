import { onBeforeUnmount, onMounted } from "vue";
import type { Ref } from "vue";

function useEventListener(
  element: Ref<HTMLElement | null>,
  event: string,
  callback: EventListener,
): void {
  onMounted(() => {
    element?.value?.addEventListener(event, callback);
  });

  onBeforeUnmount(() => {
    element?.value?.removeEventListener(event, callback);
  });
}

export default useEventListener;
