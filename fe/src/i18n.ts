import i18n from "i18next";
import LanguageDetector from "i18next-browser-languagedetector";
import HttpApi from "i18next-http-backend";
import { initReactI18next } from "react-i18next";

export const LANGUAGES: { [key: string]: any } = {
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

export default i18n;
