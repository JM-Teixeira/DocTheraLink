import { describe, expect, it } from "vitest";

import en from "../../public/locale/en.json";
import pt from "../../public/locale/pt.json";

// i18next uses Intl plural rules; for "pt" (Brazilian rules) 0 counts as
// "one", so a count of 0 would show the singular text ("1 pessoa"). Every
// pt plural pair needs a _zero form (and no long dashes in what people read).
describe("locales", () => {
  const ptKeys = pt as unknown as Record<string, string>;
  const enKeys = en as unknown as Record<string, string>;

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

  it("every page help topic has a text in both languages", () => {
    // The source of every component, to find the help="<topic>" of each page
    const sources = import.meta.glob("../components/**/*.tsx", {
      query: "?raw",
      import: "default",
      eager: true,
    }) as Record<string, string>;
    const topics = new Set<string>();
    for (const src of Object.values(sources))
      for (const m of src.matchAll(/\shelp="([a-z_]+)"/g)) topics.add(m[1]);
    expect(topics.size).toBeGreaterThan(10);
    const missing = [...topics].filter(
      (k) => !enKeys[`help__${k}`] || !ptKeys[`help__${k}`],
    );
    expect(missing).toEqual([]);
  });
});
