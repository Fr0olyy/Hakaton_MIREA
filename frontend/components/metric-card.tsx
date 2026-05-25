import { ReactNode } from "react";
import { Card, CardContent } from "@/components/ui/card";

type MetricCardProps = {
  label: string;
  value: ReactNode;
  hint?: string;
  tone?: "blue" | "green" | "orange" | "red" | "slate";
  detail?: ReactNode;
};

const tones = {
  blue: "border-l-primary",
  green: "border-l-secondary",
  orange: "border-l-accent",
  red: "border-l-destructive",
  slate: "border-l-slate-400",
};

export function MetricCard({ label, value, hint, tone = "slate", detail }: MetricCardProps) {
  return (
    <Card className={`border-l-4 ${tones[tone]}`}>
      <CardContent className="p-4">
        <div className="text-[11px] font-semibold uppercase tracking-wide text-muted-foreground">{label}</div>
        <div className="mt-2 break-words text-2xl font-semibold leading-tight text-slate-950 md:text-3xl">{value}</div>
        {hint ? <div className="mt-1 text-xs text-muted-foreground">{hint}</div> : null}
        {detail ? <div className="mt-3 border-t pt-3 text-xs text-muted-foreground">{detail}</div> : null}
      </CardContent>
    </Card>
  );
}
