"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { BarChart3, Brain, Download, Lightbulb, ListChecks, Map, Upload } from "lucide-react";
import { cn } from "@/lib/utils";

const items = [
  { href: "upload", label: "Upload", icon: Upload },
  { href: "dashboard", label: "Dashboard", icon: BarChart3 },
  { href: "probabilistic", label: "Probabilistic", icon: Brain },
  { href: "queue", label: "Queue", icon: ListChecks },
  { href: "recommendations", label: "Recommendations", icon: Lightbulb },
  { href: "roadmap", label: "Roadmap", icon: Map },
  { href: "export", label: "Export", icon: Download },
];

export function ProjectNav({ projectId }: { projectId: string }) {
  const pathname = usePathname();
  return (
    <div className="mb-5 border-b">
      <nav className="flex flex-wrap gap-x-1 gap-y-0">
        {items.map((item) => {
          const href = `/projects/${projectId}/${item.href}`;
          const Icon = item.icon;
          const active = pathname === href;
          return (
            <Link
              key={item.href}
              href={href}
              className={cn(
                "flex items-center gap-2 border-b-2 px-3 py-3 text-sm font-medium",
                active
                  ? "border-primary text-primary"
                  : "border-transparent text-muted-foreground hover:border-muted-foreground/40 hover:text-foreground",
              )}
            >
              <Icon className="h-4 w-4" />
              {item.label}
            </Link>
          );
        })}
      </nav>
    </div>
  );
}
