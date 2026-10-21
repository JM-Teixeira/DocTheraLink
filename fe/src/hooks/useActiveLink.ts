import { usePath } from "raviger";

/**
 * Keys can be any value that is present in the url.
 * Value must be equivalent to the link of the sidebar item.
 * Entries must be sorted based on priority.
 * Whichever key is present in the url first, becomes the active link returned
 * by `useActiveLink` hook.
 */
const activeLinkPriority = {
  "/patients": "/patients",
  "/patient/": "/patients",
  "/referrals": "/referrals",
  "/time-board": "/time-board",
  "/inventory": "/inventory",
  "/users": "/users",
  "/facility": "/facility",
};

/**
 * @returns The active link of the current path.
 */
export default function useActiveLink() {
  const path = usePath() || "";
  if (path === "/") return "/"; // home
  const key = Object.keys(activeLinkPriority).find((key) => path.includes(key));
  return key
    ? activeLinkPriority[key as keyof typeof activeLinkPriority]
    : undefined;
}
