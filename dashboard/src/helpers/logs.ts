import { ResponseError } from "@/openapi-generator";

export function logError(e: unknown, msg: string) {
  msg = msg + ":";
  if (e instanceof ResponseError) {
    console.error(msg, e.response.status, e.response.statusText);
  } else if (e instanceof Error) {
    // Unknown error type.
    console.error(msg, e.message);
  } else {
    console.error(msg, e ?? "Unknown error");
  }
}
