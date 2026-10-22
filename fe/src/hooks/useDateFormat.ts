import { useAuthContext } from "@/hooks/useAuthUser";

import {
  dateFormatOf,
  formatDate,
  formatDateTime,
  parseTypedDate,
  typingPattern,
} from "@/Utils/dates";

/** The signed-in person's date format and formatters bound to it. */
export default function useDateFormat() {
  const pref = dateFormatOf(useAuthContext().user?.date_format);
  return {
    pref,
    fmt: (d?: string | Date | null) => formatDate(d, pref),
    fmtTime: (d?: string | Date | null) => formatDateTime(d, pref),
    parse: (text: string) => parseTypedDate(text, pref),
    pattern: typingPattern(pref),
  };
}
