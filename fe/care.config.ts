const env = import.meta.env;
const windowEnv = (window as any)?.ENV || {};

interface ILogo {
  light: string;
  dark: string;
}

const logo = (value?: string, fallback?: ILogo) => {
  if (!value) {
    return fallback;
  }

  try {
    return JSON.parse(value) as ILogo;
  } catch {
    // TODO: define vite plugin to validate care.config.ts during build step
    return fallback;
  }
};
/**
 * Where the API is. "same-origin" means the address the page was opened from
 * (nginx forwards /api/ to the backend), so the same deployment works from
 * localhost and from other computers on the network.
 */
export const resolveApiUrl = (raw: string | undefined, origin: string) =>
  !raw || raw === "same-origin" ? origin : raw;

const careConfig = {
  apiUrl: resolveApiUrl(
    windowEnv.REACT_CARE_API_URL || env.REACT_CARE_API_URL,
    window.location.origin,
  ),

  urls: {
    dashboard: env.REACT_DASHBOARD_URL,
    github: env.REACT_GITHUB_URL || "https://github.com/klyniva",
    ohcn: env.REACT_OHCN_URL || "https://klyniva.com",
  },

  mainLogo: logo(env.REACT_MAIN_LOGO, {
    light: "/images/klyniva.png",
    dark: "/images/klyniva.png",
  }),
  stateLogo: logo(env.REACT_STATE_LOGO),
  customLogo: logo(env.REACT_CUSTOM_LOGO),
  customLogoAlt: logo(env.REACT_CUSTOM_LOGO_ALT),
  availableLocales: (env.REACT_ALLOWED_LOCALES || "")
    .split(",")
    .map((l) => l.trim()),

  gmapsApiKey:
    env.REACT_GMAPS_API_KEY || "AIzaSyDsBAc3y7deI5ZO3NtK5GuzKwtUzQNJNUk",

  reCaptchaSiteKey:
    env.REACT_RECAPTCHA_SITE_KEY || "6LdvxuQUAAAAADDWVflgBqyHGfq-xmvNJaToM0pN",

  kasp: {
    enabled: env.REACT_KASP_ENABLED === "true",
    string: env.REACT_KASP_STRING || "KASP",
    fullString:
      env.REACT_KASP_FULL_STRING || "Karunya Arogya Suraksha Padhathi",
  },

  auth: {
    tokenRefreshInterval: env.REACT_JWT_TOKEN_REFRESH_INTERVAL
      ? parseInt(env.REACT_JWT_TOKEN_REFRESH_INTERVAL)
      : 5 * 60e3,
  },

  careApps: env.REACT_ENABLED_APPS
    ? env.REACT_ENABLED_APPS.split(",").map((app) => ({
        branch: app.split("@")[1],
        package: app.split("@")[0],
      }))
    : [],
} as const;

export default careConfig;
