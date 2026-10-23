import { navigate, usePath } from "raviger";
import { useEffect, useState } from "react";
import { useTranslation } from "react-i18next";

import useAuthUser from "@/hooks/useAuthUser";

import {
  appearanceClasses,
  startPath,
  textSizeOf,
  themeOf,
} from "@/Utils/settings";

const STARTED_KEY = "klyniva:started";

/**
 * Puts the signed-in person's settings on the page: theme and text size
 * (classes on <html>), language, and the start page right after signing in
 * (once per browser session, only from the home page).
 */
export default function useApplySettings() {
  const user = useAuthUser();
  const { i18n } = useTranslation();
  const path = usePath();
  const theme = themeOf(user.theme);
  const textSize = textSizeOf(user.text_size);

  // "System" follows the computer, also when it changes during the day
  const media =
    typeof window !== "undefined" && window.matchMedia
      ? window.matchMedia("(prefers-color-scheme: dark)")
      : null;
  const [systemDark, setSystemDark] = useState(!!media?.matches);
  useEffect(() => {
    if (!media) return;
    const onChange = (e: MediaQueryListEvent) => setSystemDark(e.matches);
    media.addEventListener("change", onChange);
    return () => media.removeEventListener("change", onChange);
  }, [media]);

  useEffect(() => {
    const root = document.documentElement;
    for (const [cls, on] of Object.entries(
      appearanceClasses(theme, textSize, systemDark),
    ))
      root.classList.toggle(cls, on);
    root.style.colorScheme = root.classList.contains("theme-dark")
      ? "dark"
      : "light";
  }, [theme, textSize, systemDark]);

  useEffect(() => {
    if (user.language && user.language !== i18n.language)
      i18n.changeLanguage(user.language);
  }, [user.language, i18n]);

  useEffect(() => {
    try {
      if (sessionStorage.getItem(STARTED_KEY)) return;
      sessionStorage.setItem(STARTED_KEY, "1");
    } catch {
      return;
    }
    const to = startPath(user.start_page);
    if (to && path === "/") navigate(to);
    // only on the first page of the session
    // eslint-disable-next-line react-hooks/exhaustive-deps
  }, []);
}
