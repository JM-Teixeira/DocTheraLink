/** Klyniva release number (major.minor.patch), set at build time from package.json. */
export const APP_VERSION =
  typeof __APP_VERSION__ === "string" ? __APP_VERSION__ : "dev";
