import { ofetch, FetchError } from "ofetch";
import type { FetchOptions } from "ofetch";
import type { ApiError, ValidationError } from "@/types";

const TIMEOUT_MS = 15_000;

// Every rejection from this module is an ApiRequestError: a real Error - for the stack
// trace and readable devtools output - carrying the server's problem-detail fields.
// `tag` is the nominal marker, since a structural type alone cannot be recognized in a
// catch block.
export interface ApiRequestError extends Error {
  readonly tag: "ApiRequestError";
  readonly title: string;
  readonly status: number;
  readonly detail?: string;
  readonly errors?: ValidationError[];
}

const apiRequestError = (body: ApiError): ApiRequestError =>
  Object.assign(new Error(body.detail || body.title), {
    name: "ApiRequestError",
    tag: "ApiRequestError" as const,
    title: body.title,
    status: body.status,
    detail: body.detail,
    errors: body.errors,
  });

// Narrows a caught value, so callers can branch on `status` without casting.
export const isApiRequestError = (err: unknown): err is ApiRequestError =>
  err instanceof Error && (err as Partial<ApiRequestError>).tag === "ApiRequestError";

const client = ofetch.create({
  baseURL: import.meta.env.VITE_API_BASE_URL || "/api",
  timeout: TIMEOUT_MS,
  // No request is ever replayed. 409 is a considered answer from this API (deleting a
  // claimed event, editing a Ticketmaster one), not a transient fault, and the write
  // endpoints are not all idempotent - a silent retry could double a bulk mutation.
  retry: false,

  onRequest({ options }) {
    const token = localStorage.getItem("auth_token");
    if (token) options.headers.set("Authorization", `Bearer ${token}`);
  },

  // A 401 means the stored token is gone or expired: drop it and send the user to
  // login, remembering where they were.
  onResponseError({ response }) {
    if (response.status !== 401) return;
    localStorage.removeItem("auth_token");
    const { pathname } = window.location;
    if (pathname !== "/login" && pathname !== "/reset-password") {
      window.location.href = `/login?redirect=${encodeURIComponent(pathname)}`;
    }
  },
});

// The timeout aborts with a named error rather than a plain AbortError, which is what
// separates "took too long" from "the network is down" in the message. `cause` is read
// through a cast because the app targets lib ES2020, where it is not yet on Error.
function isTimeout(err: FetchError): boolean {
  const cause = (err as { cause?: unknown }).cause;
  return cause instanceof Error && cause.name === "TimeoutError";
}

function toApiError(err: unknown): ApiRequestError {
  if (err instanceof FetchError) {
    // The happy path for a failure: Huma answered with a problem-detail body.
    const body = err.data as ApiError | undefined;
    if (body && typeof body === "object" && typeof body.status === "number") {
      return apiRequestError(body);
    }
    if (err.response) {
      return apiRequestError({
        title: err.response.statusText || "Request failed",
        status: err.response.status,
        detail: err.message,
      });
    }
    // No response at all: transport failure or timeout. status 0 marks "never reached
    // the server", which is what callers branch on.
    return apiRequestError({
      title: "Network Error",
      status: 0,
      detail: isTimeout(err)
        ? `Request timed out after ${err.options?.timeout ?? TIMEOUT_MS}ms`
        : err.message,
    });
  }
  return apiRequestError({
    title: "Network Error",
    status: 0,
    detail: err instanceof Error ? err.message : String(err),
  });
}

type RequestOptions = Omit<FetchOptions<"json">, "method" | "body">;

async function request<T>(url: string, options: FetchOptions<"json">): Promise<T> {
  try {
    return await client<T>(url, options);
  } catch (err) {
    throw toApiError(err);
  }
}

// Resource modules speak in verbs and get the parsed body back; nothing below this
// point knows how the request was made.
export const api = {
  get: <T>(url: string, options?: RequestOptions) => request<T>(url, { ...options, method: "GET" }),

  post: <T>(url: string, body?: unknown, options?: RequestOptions) =>
    request<T>(url, { ...options, method: "POST", body: body as FetchOptions["body"] }),

  put: <T>(url: string, body?: unknown, options?: RequestOptions) =>
    request<T>(url, { ...options, method: "PUT", body: body as FetchOptions["body"] }),

  delete: <T>(url: string, options?: RequestOptions) =>
    request<T>(url, { ...options, method: "DELETE" }),
};

export default api;
