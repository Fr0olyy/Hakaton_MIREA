import { formatPercent, formatScore } from "@/lib/utils";

type BarChartProps = {
  data: Record<string, number>;
  valueKind?: "count" | "ratio" | "score";
  maxItems?: number;
  color?: "blue" | "green" | "orange";
};

const chartColors = {
  blue: "bg-primary",
  green: "bg-secondary",
  orange: "bg-accent",
};

const donutColors = ["#0e7490", "#2f855a", "#d97706", "#475569", "#dc2626", "#7c3aed", "#0891b2", "#65a30d"];

export function BarChart({ data, valueKind = "count", maxItems = 12, color = "blue" }: BarChartProps) {
  const entries = Object.entries(data || {})
    .sort((a, b) => b[1] - a[1])
    .slice(0, maxItems);
  const max = Math.max(...entries.map(([, value]) => value), 1);

  if (!entries.length) {
    return <EmptyChart />;
  }

  return (
    <div className="space-y-3">
      {entries.map(([label, value]) => (
        <div key={label} className="grid grid-cols-[minmax(110px,180px)_1fr_76px] items-center gap-3 text-sm">
          <div className="truncate text-muted-foreground" title={label}>
            {label}
          </div>
          <div className="h-3 overflow-hidden rounded-full bg-muted">
            <div className={`h-full rounded-full ${chartColors[color]}`} style={{ width: `${Math.max(3, (value / max) * 100)}%` }} />
          </div>
          <div className="text-right font-medium">{formatChartValue(value, valueKind)}</div>
        </div>
      ))}
    </div>
  );
}

type HistogramProps = {
  values: number[];
  bins?: number;
  valueKind?: "ratio" | "score";
};

export function Histogram({ values, bins = 10, valueKind = "ratio" }: HistogramProps) {
  const clean = values.filter((value) => Number.isFinite(value));
  if (!clean.length) return <EmptyChart />;

  const counts = Array.from({ length: bins }, () => 0);
  clean.forEach((value) => {
    const clamped = Math.max(0, Math.min(1, value));
    const index = Math.min(bins - 1, Math.floor(clamped * bins));
    counts[index] += 1;
  });
  const max = Math.max(...counts, 1);

  return (
    <div className="h-48 rounded-md bg-slate-50 p-3">
      <div className="flex h-36 items-end gap-1 border-b border-l pl-2">
        {counts.map((count, index) => (
          <div key={index} className="flex min-w-0 flex-1 flex-col items-center gap-1">
            <div className="w-full rounded-t-sm bg-secondary" style={{ height: `${Math.max(count ? 8 : 0, (count / max) * 124)}px` }} title={`${count} objects`} />
          </div>
        ))}
      </div>
      <div className="mt-2 flex justify-between text-xs text-muted-foreground">
        <span>{valueKind === "ratio" ? "0%" : "0"}</span>
        <span>{valueKind === "ratio" ? "100%" : "1"}</span>
      </div>
    </div>
  );
}

export function DonutChart({ data, valueKind = "count", maxItems = 8 }: BarChartProps) {
  const entries = Object.entries(data || {})
    .filter(([, value]) => value > 0)
    .sort((a, b) => b[1] - a[1])
    .slice(0, maxItems);
  const total = entries.reduce((sum, [, value]) => sum + value, 0);
  if (!entries.length || total <= 0) return <EmptyChart />;

  let offset = 25;
  const radius = 38;
  const circumference = 2 * Math.PI * radius;

  return (
    <div className="grid gap-4 sm:grid-cols-[150px_1fr] sm:items-center">
      <svg viewBox="0 0 100 100" className="h-36 w-36 justify-self-center">
        <circle cx="50" cy="50" r={radius} fill="none" stroke="#e2e8f0" strokeWidth="12" />
        {entries.map(([label, value], index) => {
          const dash = (value / total) * circumference;
          const segment = (
            <circle
              key={label}
              cx="50"
              cy="50"
              r={radius}
              fill="none"
              stroke={donutColors[index % donutColors.length]}
              strokeWidth="12"
              strokeLinecap="round"
              strokeDasharray={`${dash} ${circumference - dash}`}
              strokeDashoffset={offset}
              transform="rotate(-90 50 50)"
            />
          );
          offset -= dash;
          return segment;
        })}
        <text x="50" y="48" textAnchor="middle" className="fill-slate-950 text-[12px] font-semibold">
          {formatChartValue(total, valueKind)}
        </text>
        <text x="50" y="61" textAnchor="middle" className="fill-slate-500 text-[7px]">
          total
        </text>
      </svg>
      <div className="space-y-2">
        {entries.map(([label, value], index) => (
          <div key={label} className="flex items-center justify-between gap-3 text-sm">
            <span className="flex min-w-0 items-center gap-2">
              <span className="h-2.5 w-2.5 shrink-0 rounded-full" style={{ background: donutColors[index % donutColors.length] }} />
              <span className="truncate text-muted-foreground">{label}</span>
            </span>
            <span className="font-medium">{formatChartValue(value, valueKind)}</span>
          </div>
        ))}
      </div>
    </div>
  );
}

export function DistributionSummary({ values }: { values: number[] }) {
  const clean = values.filter((value) => Number.isFinite(value));
  if (!clean.length) return null;
  const avg = clean.reduce((sum, value) => sum + value, 0) / clean.length;
  const max = Math.max(...clean);
  return (
    <div className="mt-3 flex gap-3 text-xs text-muted-foreground">
      <span>avg {formatScore(avg, 3)}</span>
      <span>max {formatScore(max, 3)}</span>
      <span>n {clean.length}</span>
    </div>
  );
}

function formatChartValue(value: number, kind: BarChartProps["valueKind"]) {
  if (kind === "ratio") return formatPercent(value);
  if (kind === "score") return formatScore(value, 2);
  return String(value);
}

function EmptyChart() {
  return <div className="flex h-32 items-center justify-center rounded-md border border-dashed text-sm text-muted-foreground">No data yet</div>;
}
