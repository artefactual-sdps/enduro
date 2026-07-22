# Promise dialogs

The dashboard uses a small wrapper around VueUse's `createTemplatePromise` so
callers can await a dialog result without owning its rendering or Bootstrap
lifecycle.

## Architecture

- [`DialogHost.vue`](../components/DialogHost.vue) renders pending dialogs. It
  must be mounted exactly once near the application root.
- [`DialogInstance.vue`](../components/DialogInstance.vue) connects a dialog's
  `resolve` event to its promise and supplies a cancellation result during
  forced teardown.
- [`dialog.ts`](dialog.ts) creates typed, named opener functions and prevents
  more than one dialog from being active at a time.
- [`useBootstrapModal.ts`](../composables/useBootstrapModal.ts) owns Bootstrap's
  show, hide, and disposal lifecycle for dialog components.

`App.vue` mounts the host only while the user session is valid. If a session is
invalidated with a dialog open, unmounting the host resolves the promise with
that dialog's configured cancellation value.

## Adding a dialog

Create a component that accepts its normal props and emits a typed `resolve`
event. Use `useBootstrapModal` so the result is emitted after Bootstrap has
finished hiding the modal:

```vue
<script setup lang="ts">
import useBootstrapModal from "@/composables/useBootstrapModal";

const emit = defineEmits<{
  resolve: [confirmed: boolean];
}>();

let confirmed = false;
const { element, hide } = useBootstrapModal(() => {
  emit("resolve", confirmed);
});

const confirm = () => {
  confirmed = true;
  hide();
};
</script>

<template>
  <div ref="element" class="modal">
    <!-- Dialog markup. -->
  </div>
</template>
```

Define a named opener next to the other dialog definitions. The cancellation
value must be a valid result for the dialog:

```ts
import ExampleDialog from "@/components/ExampleDialog.vue";
import { defineDialog } from "@/dialogs/dialog";

export const openExampleDialog = defineDialog<boolean>(ExampleDialog, false);
```

Callers can then await the result without depending on the component:

```ts
const confirmed = await openExampleDialog();
if (!confirmed) return;
```

## Lifecycle constraints

- Only one dialog may be active. A second request rejects with
  `DialogAlreadyOpenError` rather than reusing another caller's promise.
- Resolve through the callback passed to `useBootstrapModal`. Resolving earlier
  removes the component before Bootstrap has completed its cleanup.
- Dialogs currently omit Bootstrap's `.fade` class. Forced teardown calls
  `hide()` immediately followed by `dispose()`, so hiding must be synchronous.
  If animation is added, update the composable to dispose after
  `hidden.bs.modal` instead.

## Testing

Test each dialog's confirmed and cancelled results, and test callers by mocking
the named opener. Keep the real-Bootstrap lifecycle test in
`useBootstrapModal.spec.ts`; mocked Bootstrap methods cannot detect leaked body
scroll locks or backdrops.
