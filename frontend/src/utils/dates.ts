// Date formatting shared across the app.

export function formatDate(dateString: string): string {
  try {
    const options: Intl.DateTimeFormatOptions = {
      year: "numeric",
      month: "2-digit",
      day: "2-digit",
    };
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
      return "Invalid Date";
    }
    return date.toLocaleDateString("en-GB", options);
  } catch (error) {
    console.error("Error formatting date:", error);
    return "Invalid Date";
  }
}

// A value with no 'T' is date-only: render without a time and parse from its parts, so
// 00:00 UTC never shifts to the previous evening local.
export const isDateOnly = (s: string) => !s.includes("T");

export function parseLocal(s: string): Date {
  if (isDateOnly(s)) {
    const [y, m, d] = s.slice(0, 10).split("-").map(Number);
    return new Date(y, (m || 1) - 1, d || 1);
  }
  return new Date(s);
}

export function formatEventDate(dateString: string, endDateString?: string): string {
  try {
    const dateOnly = isDateOnly(dateString);
    const options: Intl.DateTimeFormatOptions = {
      year: "numeric",
      month: "long",
      day: "numeric",
      ...(dateOnly ? {} : { hour: "numeric", minute: "numeric" }),
    };
    const startDate = parseLocal(dateString);
    if (isNaN(startDate.getTime())) {
      return "Invalid Date";
    }

    if (endDateString) {
      const endDate = parseLocal(endDateString);
      if (isNaN(endDate.getTime())) {
        return "Invalid Date";
      }
      return `${startDate.toLocaleDateString("en-GB", options)} – ${endDate.toLocaleDateString("en-GB", options)}`;
    }
    return startDate.toLocaleDateString("en-GB", options);
  } catch (error) {
    console.error("Error formatting event date:", error);
    return "Invalid Date";
  }
}

export function formatRelativeDate(dateString: string): string {
  try {
    const date = new Date(dateString);
    if (isNaN(date.getTime())) {
      return "Invalid Date";
    }
    const now = new Date();
    const diffMs = date.getTime() - now.getTime();
    const diffDays = Math.round(diffMs / (1000 * 60 * 60 * 24));

    if (diffDays === 0) return "Today";
    if (diffDays === 1) return "Tomorrow";
    if (diffDays === -1) return "Yesterday";
    if (diffDays > 0 && diffDays <= 7) return `In ${diffDays} days`;
    if (diffDays < 0 && diffDays >= -7) return `${Math.abs(diffDays)} days ago`;

    return formatDate(dateString);
  } catch {
    return "Invalid Date";
  }
}

// Compact "Mon DD" form for dense tables (e.g. "Nov 22").
export function formatShort(dateString: string): string {
  const date = new Date(dateString);
  if (isNaN(date.getTime())) return "—";
  return date.toLocaleDateString("en-GB", { month: "short", day: "numeric" });
}

// ISO-style "2026-10-06" (or compact "26-10-06") for dense, aligned, year-bearing
// date columns. Unambiguous and sorts visually.
export function formatISO(dateString: string, compact = false): string {
  // Date-only values are already ISO - take the prefix verbatim so no timezone shift.
  if (isDateOnly(dateString)) {
    const iso = dateString.slice(0, 10);
    return compact ? iso.slice(2) : iso;
  }
  const date = new Date(dateString);
  if (isNaN(date.getTime())) return "—";
  const y = date.getFullYear();
  const m = String(date.getMonth() + 1).padStart(2, "0");
  const d = String(date.getDate()).padStart(2, "0");
  return compact ? `${String(y).slice(2)}-${m}-${d}` : `${y}-${m}-${d}`;
}

// "Mon DD, YYYY" for slightly roomier contexts.
export function formatMedium(dateString: string): string {
  const date = new Date(dateString);
  if (isNaN(date.getTime())) return "—";
  return date.toLocaleDateString("en-GB", {
    year: "numeric",
    month: "short",
    day: "numeric",
  });
}
