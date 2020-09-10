const enablePurge = process.env.NODE_ENV === "production";

module.exports = {
  future: {
    purgeLayersByDefault: true,
    removeDeprecatedGapUtilities: true,
  },
  plugins: [],
  purge: {
    content: ["./src/**/*.html", "./src/**/*.scss"],
    enabled: enablePurge,
  },
  theme: {
    extend: {},
  },
  variants: {},
};
