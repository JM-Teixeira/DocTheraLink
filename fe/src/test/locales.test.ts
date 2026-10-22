import { describe, expect, it } from "vitest";

import en from "../../public/locale/en.json";
import pt from "../../public/locale/pt.json";

// i18next uses Intl plural rules; for "pt" (Brazilian rules) 0 counts as
// "one", so a count of 0 would show the singular text ("1 pessoa"). Every
// pt plural pair needs a _zero form (and no long dashes in what people read).
describe("locales", () => {
  const ptKeys = pt as Record<string, string>;

  it("every Portuguese plural has a zero form", () => {
    const missing = Object.keys(ptKeys)
      .filter(
        (k) =>
          k.endsWith("_one") && ptKeys[`${k.slice(0, -4)}_other`] !== undefined,
      )
      .map((k) => `${k.slice(0, -4)}_zero`)
      .filter((k) => ptKeys[k] === undefined);
    expect(missing).toEqual([]);
  });

  it("no long dashes in the texts", () => {
    const withDash = [...Object.entries(en), ...Object.entries(pt)]
      .filter(([, v]) => typeof v === "string" && /[—–]/.test(v))
      .map(([k]) => k);
    expect(withDash).toEqual([]);
  });
});
