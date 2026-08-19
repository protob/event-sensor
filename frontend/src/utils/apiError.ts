import type { ApiError } from "@/types";

// Requests reject with an ApiRequestError, which carries the server's problem-detail
// fields alongside the usual Error shape - prefer `detail` (the specific reason) over
// `title` (the generic class) before falling back.
export function errMessage(err: unknown, fallback: string): string {
  if (typeof err === "object" && err !== null && "detail" in err) {
    return (err as ApiError).detail || (err as ApiError).title || fallback;
  }
  if (err instanceof Error) return err.message;
  return fallback;
}
