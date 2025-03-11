import globals from "globals";
import pluginJs from "@eslint/js";
import tseslint from "typescript-eslint";
import pluginVue from "eslint-plugin-vue";
import pluginImport from "eslint-plugin-import";

/** @type {import('eslint').Linter.Config[]} */
export default [
  { files: ["**/*.{js,mjs,cjs,ts,vue}"] },
  {
    ignores: [
      "build",
      "coverage",
      "dist",
      "node_modules",
      "public",
      "src/openapi-generator",
    ],
  },
  {
    settings: {
      "import/resolver": {
        typescript: { project: "tsconfig*.json" },
      },
    },
  },
  { languageOptions: { globals: globals.browser } },
  pluginJs.configs.recommended,
  pluginImport.flatConfigs.recommended,
  ...tseslint.configs.recommended,
  ...pluginVue.configs["flat/essential"],
  {
    files: ["**/*.{ts,vue}"],
    languageOptions: { parserOptions: { parser: tseslint.parser } },
    rules: {
      "import/no-unresolved": [
        "error",
        {
          ignore: ["~icons"], // ignore unplugin-icon auto import paths.
        },
      ],
      "vue/multi-word-component-names": "off", // doesn't work with auto-router paths.
      "import/order": [
        "error",
        {
          alphabetize: { order: "asc" },
          named: true,
          "newlines-between": "always",
        },
      ],
    },
  },
];
