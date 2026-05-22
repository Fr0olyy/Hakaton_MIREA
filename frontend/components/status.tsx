import { Badge } from "@/components/ui/badge";

export function StatusBadge({ status }: { status?: string }) {
  const value = status || "unknown";
  const variant =
    value === "ok" || value === "done" || value === "uploaded"
      ? "success"
      : value.includes("error") || value.includes("bad") || value.includes("failed") || value === "invalid"
        ? "danger"
        : value.includes("duplicate") || value.includes("label") || value.includes("review")
          ? "warning"
          : "muted";
  return <Badge variant={variant}>{value}</Badge>;
}

export function PriorityBadge({ priority }: { priority?: string | number }) {
  const text = String(priority ?? "normal");
  const normalized = text.toLowerCase();
  const variant = normalized.includes("high") || normalized === "1" ? "danger" : normalized.includes("medium") || normalized === "2" ? "warning" : "muted";
  return <Badge variant={variant}>{text}</Badge>;
}
