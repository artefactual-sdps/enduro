import { snakeCase } from "lodash-es";

// Parses a JSON string and recursively transforms all object
// keys to snake_case, returnning the transformed object.
export function transformKeys(value: string): unknown {
  const json = JSON.parse(value);

  const transform = (data: unknown): unknown => {
    if (data === null || typeof data !== "object") {
      return data;
    }

    if (Array.isArray(data)) {
      return data.map(transform);
    }

    const transformed: Record<string, unknown> = {};
    for (const [key, value] of Object.entries(
      data as Record<string, unknown>,
    )) {
      transformed[snakeCase(key)] = transform(value);
    }

    return transformed;
  };

  return transform(json);
}
