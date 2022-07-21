/* eslint-env node */
require("@rushstack/eslint-patch/modern-module-resolution");

module.exports = {
  root: true,
  globals: {
    $ref: "readonly",
    $computed: "readonly",
    $shallowRef: "readonly",
    $customRef: "readonly",
    $toRef: "readonly",
    $$: "readonly",
    window: "readonly",
  },
  extends: [
    // "plugin:vue/vue3-essential",
    "eslint:recommended",
    "@vue/eslint-config-typescript/recommended",
    "@vue/eslint-config-prettier",
  ],
  ignorePatterns: [
    "coverage/**",
    "dist/**",
    "node_modules/**",
    "src/openapi-generator/**",
  ],
  rules: {
    "@typescript-eslint/no-explicit-any": "off",
    "sort-imports": [
      "error",
      {
        ignoreCase: false,
        ignoreDeclarationSort: false,
        ignoreMemberSort: false,
        memberSyntaxSortOrder: ["none", "all", "multiple", "single"],
        allowSeparatedGroups: false,
      },
    ],
  },
};
