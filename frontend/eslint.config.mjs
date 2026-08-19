import js from "@eslint/js";
import globals from "globals";
import pluginVue from "eslint-plugin-vue";
import tseslint from "typescript-eslint";
import prettierConfig from "eslint-config-prettier";

export default [
  js.configs.recommended,
  ...tseslint.configs.recommended,
  ...pluginVue.configs["flat/recommended"],
  prettierConfig,
  {
    files: ["**/*.{ts,vue}"],
    languageOptions: {
      globals: globals.browser,
      parserOptions: {
        parser: tseslint.parser,
      },
    },
    rules: {
      // TypeScript already resolves identifiers; no-undef only duplicates it and misfires on
      // type-only names.
      "no-undef": "off",
      // Views and single-instance shell components are named for what they are, not for a
      // vendor prefix.
      "vue/multi-word-component-names": "off",
      // Optional props without a default are deliberate here - absent means "not set", which
      // is not the same as any default this rule would want.
      "vue/require-default-prop": "off",
      // A leading underscore marks a binding that exists only to be discarded, as in the
      // `const { [id]: _drop, ...rest }` omit idiom.
      "@typescript-eslint/no-unused-vars": [
        "error",
        { argsIgnorePattern: "^_", varsIgnorePattern: "^_", caughtErrorsIgnorePattern: "^_" },
      ],
    },
  },
  {
    files: ["*.config.{js,mjs,ts}", "vite.config.ts"],
    languageOptions: {
      globals: globals.node,
    },
  },
  {
    ignores: ["dist/", "node_modules/"],
  },
];
