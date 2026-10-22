import { describe, expect, it } from "vitest";

import { helpBlocks } from "@/components/Common/PageHelp";

describe("helpBlocks", () => {
  it("turns '- ' lines into lists and other lines into paragraphs", () => {
    expect(helpBlocks("Intro.\n- one\n- two\n\nMore.\n- three")).toEqual([
      { kind: "p", text: "Intro." },
      { kind: "ul", items: ["one", "two"] },
      { kind: "p", text: "More." },
      { kind: "ul", items: ["three"] },
    ]);
  });
});
