import { createTemplatePromise } from "@vueuse/core";
import type { Component } from "vue";

type NoDialogProps = Record<never, never>;
type DialogComponent = new (...args: never[]) => unknown;

export interface DialogRequest {
  component: Component;
  props?: object;
}

// Vue exposes a component's typed props and emits on its public instance.
// Derive the request arguments and result so callers do not repeat those types.
type DialogProps<Dialog extends DialogComponent> =
  InstanceType<Dialog> extends { $props: infer Props extends object }
    ? Props
    : never;

type DialogResult<Dialog extends DialogComponent> =
  InstanceType<Dialog> extends {
    $emit: (event: "resolve", ...args: infer Arguments) => void;
  }
    ? Arguments[0]
    : never;

type DialogArguments<Props extends object> = NoDialogProps extends Props
  ? [props?: Props]
  : [props: Props];

export class DialogAlreadyOpenError extends Error {
  constructor() {
    super("Cannot open a dialog while another dialog is active.");
    this.name = "DialogAlreadyOpenError";
  }
}

// DialogHost renders this template once at the application root. Keeping it in
// this module lets callers start dialogs without coupling them to the host.
export const dialogTemplate = createTemplatePromise<
  unknown,
  [request: DialogRequest]
>();

// Overlapping dialogs are rejected explicitly instead of sharing the first
// promise or rendering multiple Bootstrap modals and backdrops.
let active = false;

/**
 * Opens a Vue component as a dialog and resolves with its `resolve` event
 * payload.
 *
 * Pass the concrete dialog component directly; there is no dialog registry or
 * opener wrapper. Its `defineProps` type determines whether the second argument
 * is required and which properties it accepts. Its typed `resolve` event
 * determines the promise result.
 *
 * Normal user cancellation should be represented by the component's result,
 * such as `false` or `null`. The promise instead resolves with `undefined` if
 * the application removes the dialog host before the component produces a
 * result, such as when a session expires.
 *
 * Only one dialog can be active at a time.
 *
 * @throws `DialogAlreadyOpenError` when another dialog is already active.
 *
 * @example
 * ```ts
 * const confirmed = await openDialog(ConfirmDialog, {
 *   heading: "Continue?",
 * });
 * ```
 */
export async function openDialog<Dialog extends DialogComponent>(
  component: Dialog,
  ...args: DialogArguments<DialogProps<Dialog>>
): Promise<DialogResult<Dialog> | undefined> {
  if (active) {
    throw new DialogAlreadyOpenError();
  }

  active = true;
  try {
    return (await dialogTemplate.start({
      component: component as Component,
      props: args[0],
    })) as DialogResult<Dialog> | undefined;
  } finally {
    active = false;
  }
}
