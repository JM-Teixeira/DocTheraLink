/// <reference types="vite/client" />
/// <reference types="vite-plugin-pwa/client" />
interface ImportMetaEnv {
  readonly CARE_CDN_URL?: string;

  readonly REACT_APP_TITLE: string;
  readonly REACT_APP_META_DESCRIPTION: string;
  readonly REACT_APP_COVER_IMAGE: string;
  readonly REACT_APP_COVER_IMAGE_ALT: string;
  readonly REACT_PUBLIC_URL: string;
  readonly REACT_CARE_API_URL: string;
  readonly REACT_DASHBOARD_URL?: string;
  readonly REACT_GITHUB_URL?: string;
  readonly REACT_OHCN_URL?: string;
  readonly REACT_HEADER_LOGO?: string;
  readonly REACT_MAIN_LOGO?: string;
  readonly REACT_STATE_LOGO?: string;
  readonly REACT_CUSTOM_LOGO?: string;
  readonly REACT_CUSTOM_LOGO_ALT?: string;
  readonly REACT_CUSTOM_DESCRIPTION?: string;
  readonly REACT_GMAPS_API_KEY?: string;
  readonly REACT_RECAPTCHA_SITE_KEY?: string;
  readonly REACT_KASP_ENABLED?: string;
  readonly REACT_KASP_STRING?: string;
  readonly REACT_KASP_FULL_STRING?: string;
  readonly REACT_JWT_TOKEN_REFRESH_INTERVAL?: string;
  readonly REACT_ALLOWED_LOCALES?: string;
  readonly REACT_ENABLED_APPS?: string;
}

interface ImportMeta {
  readonly env: ImportMetaEnv;
}
