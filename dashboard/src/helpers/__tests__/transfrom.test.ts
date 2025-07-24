import { describe, expect, it } from "vitest";

import { transformKeys } from "@/helpers/transform";

describe("transformKeys", () => {
  it("handles primitive values", () => {
    expect(transformKeys('{"stringTest": "test"}')).toEqual({
      string_test: "test",
    });
    expect(transformKeys('{"numberTest": 42}')).toEqual({ number_test: 42 });
    expect(transformKeys('{"booleanTest": true}')).toEqual({
      boolean_test: true,
    });
    expect(transformKeys('{"nullTest": null}')).toEqual({ null_test: null });
  });

  it("transforms camelCase keys to snake_case", () => {
    const input = '{"firstName": "John", "lastName": "Doe"}';
    const expected = { first_name: "John", last_name: "Doe" };
    expect(transformKeys(input)).toEqual(expected);
  });

  it("transforms PascalCase keys to snake_case", () => {
    const input = '{"FirstName": "John", "LastName": "Doe"}';
    const expected = { first_name: "John", last_name: "Doe" };
    expect(transformKeys(input)).toEqual(expected);
  });

  it("transforms kebab-case keys to snake_case", () => {
    const input = '{"first-name": "John", "last-name": "Doe"}';
    const expected = { first_name: "John", last_name: "Doe" };
    expect(transformKeys(input)).toEqual(expected);
  });

  it("handles keys that are already snake_case", () => {
    const input = '{"first_name": "John", "last_name": "Doe"}';
    const expected = { first_name: "John", last_name: "Doe" };
    expect(transformKeys(input)).toEqual(expected);
  });

  it("handles nested objects recursively", () => {
    const input = JSON.stringify({
      userName: "john",
      userProfile: {
        firstName: "John",
        contactInfo: {
          emailAddress: "john@example.com",
          phoneNumber: "123-456-7890",
        },
      },
    });

    const expected = {
      user_name: "john",
      user_profile: {
        first_name: "John",
        contact_info: {
          email_address: "john@example.com",
          phone_number: "123-456-7890",
        },
      },
    };

    expect(transformKeys(input)).toEqual(expected);
  });

  it("handles arrays of primitives", () => {
    const input = '{"itemList": [1, 2, 3]}';
    const expected = { item_list: [1, 2, 3] };
    expect(transformKeys(input)).toEqual(expected);
  });

  it("handles arrays of objects", () => {
    const input = JSON.stringify({
      userList: [
        { firstName: "John", lastName: "Doe" },
        { firstName: "Jane", lastName: "Smith" },
      ],
    });

    const expected = {
      user_list: [
        { first_name: "John", last_name: "Doe" },
        { first_name: "Jane", last_name: "Smith" },
      ],
    };

    expect(transformKeys(input)).toEqual(expected);
  });

  it("handles nested arrays", () => {
    const input = JSON.stringify({
      matrixData: [
        [1, 2],
        [3, 4],
      ],
      complexArray: [{ itemName: "test", subItems: [{ subName: "sub1" }] }],
    });

    const expected = {
      matrix_data: [
        [1, 2],
        [3, 4],
      ],
      complex_array: [{ item_name: "test", sub_items: [{ sub_name: "sub1" }] }],
    };

    expect(transformKeys(input)).toEqual(expected);
  });

  it("handles empty objects and arrays", () => {
    expect(transformKeys('{"emptyObject": {}, "emptyArray": []}')).toEqual({
      empty_object: {},
      empty_array: [],
    });
  });

  it("handles mixed data types", () => {
    const input = JSON.stringify({
      stringValue: "test",
      numberValue: 42,
      booleanValue: true,
      nullValue: null,
      arrayValue: [1, "two", true, null],
      objectValue: { nestedKey: "nested" },
    });

    const expected = {
      string_value: "test",
      number_value: 42,
      boolean_value: true,
      null_value: null,
      array_value: [1, "two", true, null],
      object_value: { nested_key: "nested" },
    };

    expect(transformKeys(input)).toEqual(expected);
  });

  it("handles deeply nested structures", () => {
    const input = JSON.stringify({
      level1: {
        level2: {
          level3: {
            deepValue: "found",
            deepArray: [{ veryDeep: "value" }],
          },
        },
      },
    });

    const expected = {
      level_1: {
        level_2: {
          level_3: {
            deep_value: "found",
            deep_array: [{ very_deep: "value" }],
          },
        },
      },
    };

    expect(transformKeys(input)).toEqual(expected);
  });

  it("throws error for invalid JSON", () => {
    expect(() => transformKeys("invalid json")).toThrow();
  });
});
