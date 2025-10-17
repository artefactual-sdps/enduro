import globals from "globals";
import pluginJs from "@eslint/js";
import tseslint from "typescript-eslint";
import pluginVue from "eslint-plugin-vue";
import pluginImport from "eslint-plugin-import";
import eslintPluginPrettierRecommended from "eslint-plugin-prettier/recommended";

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
  ...pluginVue.configs["flat/strongly-recommended"],
  ...pluginVue.configs["flat/recommended"],
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
      "vue/no-v-html": "off", // must be fixed soon.
      "@typescript-eslint/naming-convention": [
        "error",
        {
          selector: "function",
          format: ["camelCase"],
        },
        {
          selector: "variableLike",
          format: ["camelCase"],
          leadingUnderscore: "allow",
        },
        {
          selector: "typeLike",
          format: ["PascalCase"],
        },
      ],
      "vue/component-name-in-template-casing": [
        "error",
        "PascalCase",
        { registeredComponentsOnly: false },
      ],
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
  eslintPluginPrettierRecommended,
];
