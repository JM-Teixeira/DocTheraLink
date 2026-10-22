import { describe, expect, it } from "vitest";

import {
  appearanceClasses,
  startPath,
  textSizeOf,
  themeOf,
} from "@/Utils/settings";

describe("settings", () => {
  it("reads stored values with safe defaults", () => {
    expect(themeOf("dark")).toBe("dark");
    expect(themeOf("pink")).toBe("light");
    expect(themeOf(undefined)).toBe("light");
    expect(textSizeOf("large")).toBe("large");
    expect(textSizeOf("")).toBe("normal");
  });

  it("follows the computer for the system theme", () => {
    expect(appearanceClasses("system", "normal", true)).toEqual({
      dark: true,
      "text-large": false,
    });
    expect(appearanceClasses("system", "large", false)).toEqual({
      dark: false,
      "text-large": true,
    });
    expect(appearanceClasses("dark", "normal", false).dark).toBe(true);
  });

  it("only goes to known start pages", () => {
    expect(startPath("time-board")).toBe("/time-board");
    expect(startPath("home")).toBe("");
    expect(startPath("//evil.example")).toBe("");
    expect(startPath(undefined)).toBe("");
  });
});
