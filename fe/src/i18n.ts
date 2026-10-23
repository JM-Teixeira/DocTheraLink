import dayjs from "dayjs";
import "dayjs/locale/pt";
import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import HttpApi from "i18next-http-backend";
import { initReactI18next } from "react-i18next";

export const LANGUAGES: Record<string, string> = {
  en: "English",
  pt: "Português",
};

i18n
  .use(HttpApi)
  .use(initReactI18next)
  .use(LanguageDetector)
  .init({
    backend: {
      loadPath: "/locale/{{lng}}.json",
    },
    fallbackLng: "en",
    load: "currentOnly",
    supportedLngs: Object.keys(LANGUAGES),
    interpolation: {
      escapeValue: false,
      skipOnVariables: false,
    },
  });

// Dates ("quarta-feira, 21 de outubro") follow the interface language
const dateLocale = (lng?: string) => (lng?.startsWith("pt") ? "pt" : "en");
i18n.on("languageChanged", (lng) => dayjs.locale(dateLocale(lng)));
dayjs.locale(dateLocale(i18n.language));

export default i18n;
