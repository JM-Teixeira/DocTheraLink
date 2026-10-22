import { APIRequestContext, Page, expect } from "@playwright/test";

/** Demo accounts (local demo data only). */
export const PASSWORD = "1234@1";

/** Signs in through the API and stores the tokens like the app does. */
export async function signIn(page: Page, request: APIRequestContext, username: string, lang = "pt") {
  const res = await request.post("/api/v1/auth/login/", { data: { username, password: PASSWORD } });
  expect(res.ok(), `login ${username}`).toBeTruthy();
  const { access, refresh } = await res.json();
  await page.addInitScript(
    ([a, r, l]) => {
      localStorage.setItem("care_access_token", a);
      localStorage.setItem("care_refresh_token", r);
      localStorage.setItem("i18nextLng", l);
      localStorage.removeItem("klyniva.board.filter");
    },
    [access, refresh, lang],
  );
  return access as string;
}

/** Calls the API as someone (for setup and clean-up). */
export async function api(request: APIRequestContext, token: string, method: string, path: string, data?: unknown) {
  return request.fetch(`/api/v1${path}`, { method, data, headers: { Authorization: `Bearer ${token}` } });
}

/** Text of the page without hidden elements. */
export const visibleText = (page: Page) => page.locator("body").innerText();
