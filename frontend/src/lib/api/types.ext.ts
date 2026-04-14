import type { Feed, Group } from "./types";

export type ExtendedFeed = Feed & {
  auto_fetch_full_content?: boolean | null;
};

export type ExtendedGroup = Group & {
  auto_fetch_full_content?: boolean | null;
};
