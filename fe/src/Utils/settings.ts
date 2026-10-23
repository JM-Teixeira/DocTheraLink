/**
 * The person's settings (stored on their account, so they follow them to any
 * computer): theme, text size, start page and language. Pure helpers, tested
 * in settings.test.ts; useApplySettings puts them on the page.
 */

export const THEMES = ["light", "dark", "system"] as const;
export type Theme = (typeof THEMES)[number];
export const TEXT_SIZES = ["normal", "large"] as const;
export type TextSize = (typeof TEXT_SIZES)[number];
export const START_PAGES = [
  "home",
  "patients",
  "referrals",
  "time-board",
  "inventory",
] as const;
export type StartPage = (typeof START_PAGES)[number];
export const LANGUAGES = ["pt", "en"] as const;

export const themeOf = (v?: string): Theme =>
  (THEMES as readonly string[]).includes(v ?? "") ? (v as Theme) : "light";
export const textSizeOf = (v?: string): TextSize =>
  v === "large" ? "large" : "normal";

/** Whether the page should be dark ("system" follows the computer). */
export const isDark = (theme: Theme, systemDark: boolean) =>
  theme === "dark" || (theme === "system" && systemDark);

/** Classes on <html> for a theme and text size. */
export function appearanceClasses(
  theme: Theme,
  textSize: TextSize,
  systemDark: boolean,
) {
  return {
    // not Tailwind's "dark": see style/theme.css
    "theme-dark": isDark(theme, systemDark),
    "text-large": textSize === "large",
  };
}

/** Where to go after signing in ("" = stay on the home page). */
export function startPath(page?: string) {
  if (!page || page === "home") return "";
  return (START_PAGES as readonly string[]).includes(page) ? `/${page}` : "";
}
