import { describe, expect, it } from "vitest";

import { roleKey, staffQuery, statusOf } from "@/components/Staff/staffUtils";

describe("staffQuery", () => {
  const base = {
    search: "",
    role: "",
    clinic: "",
    status: "all" as const,
    limit: 20,
    offset: 0,
  };

  it("leaves out empty filters and 'all'", () => {
    expect(staffQuery(base)).toEqual({ limit: 20, offset: 0 });
  });

  it("maps every filter to its API name", () => {
    expect(
      staffQuery({
        ...base,
        search: "  inês ",
        role: "Doctor",
        clinic: "c1",
        status: "left",
        offset: 40,
      }),
    ).toEqual({
      limit: 20,
      offset: 40,
      search: "inês",
      user_type: "Doctor",
      home_facility: "c1",
      employment_status: "left",
    });
  });
});

describe("statusOf", () => {
  it("treats a missing status as active", () => {
    expect(statusOf(undefined)).toBe("active");
    expect(statusOf("")).toBe("active");
    expect(statusOf("left")).toBe("left");
    expect(statusOf("on_leave")).toBe("on_leave");
  });
});

it("roleKey prefixes the role", () => {
  expect(roleKey("Doctor")).toBe("role__Doctor");
});
