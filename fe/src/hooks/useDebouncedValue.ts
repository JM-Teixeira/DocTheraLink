import { useEffect, useState } from "react";

/**
 * Returns `value` after it has stopped changing for `ms` milliseconds, so
 * typing in a search box doesn't fire one request per keystroke.
 */
export default function useDebouncedValue<T>(value: T, ms = 300) {
  const [debounced, setDebounced] = useState(value);
  useEffect(() => {
    const id = setTimeout(() => setDebounced(value), ms);
    return () => clearTimeout(id);
  }, [value, ms]);
  return debounced;
}
