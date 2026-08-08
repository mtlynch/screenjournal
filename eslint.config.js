const js = require("@eslint/js");
const html = require("eslint-plugin-html");

module.exports = [
  {
    ignores: [
      "playwright-report/*",
      "handlers/static/third-party/**/*.js",
      "reference/*",
      "result/",
    ],
  },
  {
    files: ["handlers/templates/**/*.html"],
    plugins: { html },
  },
  {
    ...js.configs.recommended,
    languageOptions: {
      ...js.configs.recommended.languageOptions,
      ecmaVersion: 2022,
      sourceType: "module",
      globals: {
        document: "readonly",
        Event: "readonly",
        fetch: "readonly",
        Image: "readonly",
        module: "readonly",
        navigator: "readonly",
        process: "readonly",
        require: "readonly",
        setTimeout: "readonly",
        Temporal: "readonly",
        URL: "readonly",
        URLSearchParams: "readonly",
        window: "readonly",
      },
    },
    rules: {
      ...js.configs.recommended.rules,
      "no-console": [
        process.env.NODE_ENV === "production" ? "error" : "warn",
        { allow: ["error"] },
      ],
    },
  },
];
