export class FakeEventSource implements EventSource {
  static readonly CONNECTING = 0;
  static readonly OPEN = 1;
  static readonly CLOSED = 2;
  static instances: FakeEventSource[] = [];

  readonly CONNECTING = 0;
  readonly OPEN = 1;
  readonly CLOSED = 2;
  readonly url: string;
  readonly withCredentials = false;
  readyState = FakeEventSource.CONNECTING;
  onopen: ((this: EventSource, ev: Event) => unknown) | null = null;
  onmessage: ((this: EventSource, ev: MessageEvent) => unknown) | null = null;
  onerror: ((this: EventSource, ev: Event) => unknown) | null = null;

  constructor(url: string | URL) {
    this.url = String(url);
    FakeEventSource.instances.push(this);
  }

  static reset(): void {
    FakeEventSource.instances = [];
  }

  static latest(): FakeEventSource {
    const source =
      FakeEventSource.instances[FakeEventSource.instances.length - 1];
    if (!source) throw new Error("No EventSource instances");
    return source;
  }

  addEventListener(): void {}

  removeEventListener(): void {}

  dispatchEvent(): boolean {
    return true;
  }

  close(): void {
    this.readyState = FakeEventSource.CLOSED;
  }

  open(): void {
    this.readyState = FakeEventSource.OPEN;
    this.onopen?.call(this, new Event("open"));
  }

  message(data: unknown): void {
    this.onmessage?.call(
      this,
      new MessageEvent("message", { data: String(data) }),
    );
  }

  error(): void {
    this.onerror?.call(this, new Event("error"));
  }
}
