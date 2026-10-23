import { afterEach, describe, expect, it } from "vitest";

import { getViewAs, setViewAs, viewAsHeaders } from "@/Utils/viewAs";

afterEach(() => sessionStorage.clear());

describe("view as", () => {
  it("is off by default and sends no headers", () => {
    expect(getViewAs()).toBeNull();
    expect(viewAsHeaders()).toEqual({});
  });

  it("keeps the choice for the tab and sends it with every call", () => {
    setViewAs({
      role: "PhysioTherapist",
      clinic: "c1",
      clinicName: "Climutua",
    });
    expect(viewAsHeaders()).toEqual({
      "X-View-As-Role": "PhysioTherapist",
      "X-View-As-Clinic": "c1",
    });
    setViewAs(null);
    expect(getViewAs()).toBeNull();
  });

  it("ignores broken data", () => {
    sessionStorage.setItem("klyniva:view-as", "{oops");
    expect(getViewAs()).toBeNull();
  });
});
