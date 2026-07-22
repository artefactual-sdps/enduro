# Promise dialogs

The dashboard treats a dialog as an asynchronous question:

```ts
const confirmed = await openDialog(ConfirmDialog, {
  heading: "Continue?",
});

if (!confirmed) return;
```

This keeps the business flow in the component that needs the answer without
making that component own modal visibility, global placement, or Bootstrap
cleanup.

Concrete dialogs remain regular application components in
[`components`](../components). The files in this directory provide only the
shared machinery needed to open those components as promises.

There is no dialog registry and there are no named opener wrappers. Callers
pass the concrete component directly to `openDialog`. Vue's generated component
types let `openDialog` infer the component's props and the value carried by its
typed `resolve` event.

## Architecture

- [`dialog.ts`](dialog.ts) exposes `openDialog`, holds the pending request, and
  rejects attempts to open overlapping dialogs.
- [`DialogHost.vue`](DialogHost.vue) renders a pending request at the
  application root.
- [`DialogInstance.vue`](DialogInstance.vue) connects the component's `resolve`
  event to the pending promise. If the host is removed, it resolves with
  `undefined`.
- [`useDialog.ts`](useDialog.ts) carries the chosen result through Bootstrap's
  hide lifecycle and owns the modal instance.

[`App.vue`](../App.vue) mounts one `DialogHost` while the user session is valid.
When a session expires, removing the host cancels the pending request and
unmounts its component. The component's Bootstrap lifecycle hook hides the
modal before disposal, removing its backdrop and restoring document scrolling.

The request flow is:

1. A caller passes a component and optional props to `openDialog`.
2. `DialogHost` renders that component through `DialogInstance`.
3. The component asks Bootstrap to hide when the user finishes interacting.
4. Bootstrap emits `hidden.bs.modal` after restoring the document.
5. The component emits `resolve`, and `DialogInstance` settles the promise.

Only one dialog may be active. Opening another before the first settles rejects
with `DialogAlreadyOpenError`. This avoids callers sharing results or stacking
Bootstrap modals and backdrops.

## Opening a dialog

Import the component and `openDialog`:

```ts
import AboutDialog from "@/components/AboutDialog.vue";
import { openDialog } from "@/dialogs/dialog";

await openDialog(AboutDialog);
```

Pass component props as the second argument:

```ts
import LocationDialog from "@/components/LocationDialog.vue";
import { openDialog } from "@/dialogs/dialog";

const locationId = await openDialog(LocationDialog, {
  currentLocationId: aipStore.current?.locationUuid,
});

if (!locationId) return;
```

No result or props types need to be repeated at the call site:

- `defineProps` determines whether the second argument is required and which
  properties it accepts.
- The payload of the component's `resolve` event determines the promise result.
- A forced teardown, such as session expiration, adds `undefined` to the
  inferred result type.

For example, a dialog emitting `boolean` returns
`Promise<boolean | undefined>`:

```ts
const confirmed = await openDialog(BatchReviewConfirmDialog, {
  heading: "Cancel batch",
  bodyHtml: "<p>Are you sure?</p>",
  confirmClass: "btn-danger",
});
```

## Creating a dialog

Create a normal component under `src/components`. Give its props and `resolve`
event explicit TypeScript types, and pass the typed emitter to `useDialog`:

```vue
<script setup lang="ts">
import useDialog from "@/dialogs/useDialog";

defineProps<{
  heading: string;
}>();

const emit = defineEmits<{
  resolve: [confirmed: boolean];
}>();

const { element, close } = useDialog(emit, false);
</script>

<template>
  <div ref="element" class="modal" tabindex="-1">
    <div class="modal-dialog">
      <div class="modal-content">
        <h1>{{ heading }}</h1>
        <button type="button" @click="close(true)">Confirm</button>
        <button type="button" data-bs-dismiss="modal">Cancel</button>
      </div>
    </div>
  </div>
</template>
```

The result should represent normal user cancellation as well as confirmation.
Common choices are `false` for confirmation dialogs and `null` for selection or
text-entry dialogs. Forced host teardown is distinct from user cancellation and
resolves with `undefined`.

The arguments after `emit` are the default result used when Bootstrap dismisses
the modal directly. `close(result)` overrides that default and hides the modal.
For a resultless dialog, declare `resolve: []` and call `useDialog(emit)` with no
default result.

After creating the component, callers can import and open it directly. There is
no registration step or companion opener file.

## Lifecycle rules

- Do not emit `resolve` directly from an action button. Call `close(result)` and
  let `useDialog` emit after Bootstrap finishes hiding. Otherwise Vue can remove
  the component before Bootstrap restores body scrolling and removes its
  backdrop.
- Dialogs currently omit Bootstrap's `.fade` class. Forced teardown calls
  `hide()` immediately followed by `dispose()`, so hiding must remain
  synchronous. If animation is added, disposal must wait for
  `hidden.bs.modal`.
- Treat `undefined` as forced cancellation when a caller needs to distinguish
  it from the dialog's normal user-cancellation result.
- Keep `DialogHost` mounted exactly once. The application owns it; individual
  pages and components should only call `openDialog`.
- Do not deliberately overlap dialogs. `DialogAlreadyOpenError` indicates a
  control-flow problem that should be resolved by the caller.

## Testing

Dialog component tests may mock Bootstrap's `show`, `hide`, and `dispose`
methods. Dispatch `hidden.bs.modal` explicitly to verify the emitted result.

Caller tests can mock `openDialog` and should verify the component and relevant
props:

```ts
expect(openDialogMock).toHaveBeenCalledWith(ExampleDialog, {
  heading: "Continue?",
});
```

Keep lifecycle behavior covered separately with real Bootstrap:

- [`useDialog.spec.ts`](useDialog.spec.ts) verifies that forced unmount restores
  body styles and removes the backdrop.
- [`DialogHost.spec.ts`](DialogHost.spec.ts) verifies results, cancellation, and
  overlapping-request rejection.
- [`App.spec.ts`](../App.spec.ts) covers the complete session-expiration path.

When adding a dialog, test both its successful result and its user-cancellation
result. Add or update a caller test so the expected component and props are part
of the tested contract.
