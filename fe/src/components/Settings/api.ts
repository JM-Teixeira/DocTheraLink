import { Type } from "@/Utils/request/api";

/** A suggestion about the app (be/internal/modules/feedback). */
export interface Suggestion {
  id: string;
  author_id: number;
  author?: {
    id: number;
    username: string;
    first_name: string;
    last_name: string;
    user_type: string;
  };
  clinic_id: string | null;
  kind: SuggestionKind;
  page: string;
  text: string;
  status: SuggestionStatus;
  admin_note: string;
  created_date: string;
  modified_date: string;
}

export const SUGGESTION_KINDS = ["idea", "problem", "other"] as const;
export type SuggestionKind = (typeof SUGGESTION_KINDS)[number];
export const SUGGESTION_STATUSES = [
  "new",
  "read",
  "planned",
  "done",
  "declined",
] as const;
export type SuggestionStatus = (typeof SUGGESTION_STATUSES)[number];

const SuggestionRoutes = {
  send: {
    path: "/api/v1/suggestions/",
    method: "POST",
    TBody: Type<{ kind: SuggestionKind; page: string; text: string }>(),
    TRes: Type<Suggestion>(),
  },
  mine: {
    path: "/api/v1/suggestions/mine/",
    method: "GET",
    TRes: Type<Suggestion[]>(),
  },
  all: {
    path: "/api/v1/suggestions/",
    method: "GET",
    TRes: Type<Suggestion[]>(),
  },
  count: {
    path: "/api/v1/suggestions/count/",
    method: "GET",
    TRes: Type<{ new: number }>(),
  },
  review: {
    path: "/api/v1/suggestions/{id}/",
    method: "PATCH",
    TBody: Type<{ status?: SuggestionStatus; admin_note?: string }>(),
    TRes: Type<Suggestion>(),
  },
} as const;

export default SuggestionRoutes;
