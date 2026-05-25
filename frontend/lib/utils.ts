import { type ClassValue, clsx } from "clsx";
import { twMerge } from "tailwind-merge";

export function cn(...inputs: ClassValue[]) {
  return twMerge(clsx(inputs));
}

export function formatNumber(value: number | undefined | null, digits = 0) {
  if (typeof value !== "number" || Number.isNaN(value)) return "0";
  return new Intl.NumberFormat("en-US", {
    maximumFractionDigits: digits,
    minimumFractionDigits: digits,
  }).format(value);
}

export function formatScore(value: number | undefined | null, digits = 2) {
  if (typeof value !== "number" || Number.isNaN(value)) return "0";
  return value.toFixed(digits);
}

export function formatPercent(value: number | undefined | null, digits = 1) {
  if (typeof value !== "number" || Number.isNaN(value)) return "0%";
  return `${(value * 100).toFixed(digits)}%`;
}

export function apiPath(path: string) {
  const base = "/api/backend";

  let normalizedPath = path.startsWith("/") ? path : `/${path}`;

  // requestJSON часто вызывает "/api/projects",
  // но proxy frontend уже сам находится на "/api/backend".
  // Поэтому убираем frontend-level "/api", чтобы не получать:
  // /api/backend/api/projects
  normalizedPath = normalizedPath.replace(/^\/api(?=\/)/, "");

  return `${base}${normalizedPath}`;
}

export function numericProbabilityEntries(probabilities?: Record<string, unknown>) {
  if (!probabilities) return [];
  return Object.entries(probabilities)
    .filter(([key, value]) => !["analysis_source", "ml_status"].includes(key) && typeof value === "number")
    .map(([label, value]) => ({ label, value: value as number }))
    .sort((a, b) => b.value - a.value);
}

export function maxProbability(probabilities?: Record<string, unknown>) {
  const values = numericProbabilityEntries(probabilities).map((entry) => entry.value);
  return values.length ? Math.max(...values) : 0;
}

export function downloadBlob(blob: Blob, fileName: string) {
  const url = URL.createObjectURL(blob);
  const anchor = document.createElement("a");
  anchor.href = url;
  anchor.download = fileName;
  document.body.appendChild(anchor);
  anchor.click();
  anchor.remove();
  URL.revokeObjectURL(url);
}
