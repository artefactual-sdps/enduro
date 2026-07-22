import { createTemplatePromise } from "@vueuse/core";
import type { Component } from "vue";

type NoDialogProps = Record<never, never>;

export interface DialogRequest {
  component: Component;
  props?: object;
  // Result returned when the host disappears before the dialog resolves.
  cancelValue: unknown;
}

type DialogArguments<Props extends object> = NoDialogProps extends Props
  ? [props?: Props]
  : [props: Props];

export type DialogOpener<Result, Props extends object> = (
  ...args: DialogArguments<Props>
) => Promise<Result>;

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

async function openDialog<Result>(request: DialogRequest): Promise<Result> {
  if (active) {
    throw new DialogAlreadyOpenError();
  }

  active = true;
  try {
    return (await dialogTemplate.start(request)) as Result;
  } finally {
    active = false;
  }
}

export function defineDialog<Result, Props extends object = NoDialogProps>(
  component: Component,
  cancelValue: Result,
): DialogOpener<Result, Props> {
  const open = (props?: Props) =>
    openDialog<Result>({ component, props, cancelValue });
  // DialogOpener requires props only when Props contains required fields. The
  // runtime implementation accepts undefined for dialogs without props.
  return open as DialogOpener<Result, Props>;
}
