import { ResponseError } from "@/openapi-generator";

export function logError(e: unknown, msg: string) {
  if (e instanceof ResponseError) {
    msg = msg + ":";
    console.error(msg, e.response.status, e.response.statusText);
  } else if (e instanceof Error) {
    // Unknown error type.
    console.error(msg, e.message);
  } else {
    console.error(msg, "Unknown error");
  }
}
