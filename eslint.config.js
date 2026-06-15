const js = require("@eslint/js");

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
    ...js.configs.recommended,
    languageOptions: {
      ...js.configs.recommended.languageOptions,
      ecmaVersion: 2022,
      sourceType: "module",
      globals: {
        document: "readonly",
        fetch: "readonly",
        module: "readonly",
        navigator: "readonly",
        process: "readonly",
        require: "readonly",
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
