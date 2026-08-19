import type { EventStatus } from "@/types";

export const EVENT_STATUSES: EventStatus[] = ["interested", "going", "attended", "missed"];

export const MISSED_REASONS = ["sold_out", "clash", "cancelled", "didnt_go", "other"] as const;
