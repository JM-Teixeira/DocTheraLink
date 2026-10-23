import { useEffect, useRef, useState } from "react";

/**
 * "Today" was clicked (signal changed): returns true for a moment so the
 * views can pulse today's cell, and a ref to scroll it into view.
 */
export default function useTodayFlash<T extends HTMLElement>(signal: number) {
  const ref = useRef<T | null>(null);
  const [on, setOn] = useState(false);
  useEffect(() => {
    if (!signal) return;
    setOn(true);
    ref.current?.scrollIntoView?.({
      block: "nearest",
      inline: "center",
      behavior: "smooth",
    });
    const id = setTimeout(() => setOn(false), 1800);
    return () => clearTimeout(id);
  }, [signal]);
  return { ref, flashing: on };
}
