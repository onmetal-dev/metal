const colors = require("tailwindcss/colors");

// via https://github.com/saadeghi/daisyui/blob/1ca6dc092285b8af8c200eab3866795483968374/src/theming/themes.js#L429
const dim = {
  primary: "#9FE88D",
  secondary: "#FF7D5C",
  accent: "#C792E9",
  neutral: "#1c212b",
  "neutral-content": "#B2CCD6",
  "base-100": "#2A303C",
  "base-200": "#242933",
  "base-300": "#20252E",
  "base-content": "#B2CCD6",
  info: "#28ebff",
  success: "#62efbd",
  warning: "#efd057",
  error: "#ffae9b",
};

const cyberpunk = {
  fontFamily:
    "ui-monospace,SFMono-Regular,Menlo,Monaco,Consolas,Liberation Mono,Courier New,monospace",
  primary: "oklch(74.22% 0.209 6.35)",
  secondary: "oklch(83.33% 0.184 204.72)",
  accent: "oklch(71.86% 0.2176 310.43)",
  neutral: "oklch(23.04% 0.065 269.31)",
  "neutral-content": "oklch(94.51% 0.179 104.32)",
  "base-100": "oklch(94.51% 0.179 104.32)",
  "--rounded-box": "0",
  "--rounded-btn": "0",
  "--rounded-badge": "0",
  "--tab-radius": "0",
};

/** @type {import('tailwindcss').Config} */
module.exports = {
  content: ["./cmd/app/templates/*.templ"],
  theme: {
    container: {
      center: true,
      padding: {
        DEFAULT: "1rem",
        mobile: "2rem",
        tablet: "4rem",
        desktop: "5rem",
      },
    },
    extend: {
      colors: {
        primary: colors.blue,
        secondary: colors.yellow,
        neutral: colors.gray,
      },
    },
  },
  plugins: [
    require("@tailwindcss/forms"),
    require("@tailwindcss/typography"),
    require("daisyui"),
  ],
  daisyui: {
    themes: [
      {
        dark: {
          ...require("daisyui/src/theming/themes")["cyberpunk"],
          ...require("daisyui/src/theming/themes")["dim"],
        },
        light: {
          ...require("daisyui/src/theming/themes")["cyberpunk"],
          ...require("daisyui/src/theming/themes")["lemonade"],
        },
      },
    ],
  },
};
