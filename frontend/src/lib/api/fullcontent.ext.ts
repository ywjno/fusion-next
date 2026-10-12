import { api } from "./client";
import type { APIResponse } from "./types";

export interface StartFullContentBackfillParams {
  days: number;
  limit: number;
  feed_id?: number;
  interval_seconds: number;
}

export interface FullContentBackfillStatus {
  running: boolean;
  done: number;
  failed: number;
  skipped: number;
}

// Note: `api` lives in ./client; ./index does not re-export it.
export const fullContentAPI = {
  startBackfill: (params: StartFullContentBackfillParams) =>
    api.post<APIResponse<{ started: boolean }>>("/fullContent", params),

  getBackfillStatus: () =>
    api.get<APIResponse<FullContentBackfillStatus>>("/fullContent"),
};
