"use client";

import type { ComponentType } from "react";
import Link from "next/link";
import { usePathname } from "next/navigation";
import { BarChart3, Bot, Brain, ClipboardList, Download, FlaskConical, Lightbulb, ListChecks, Map, Search, Settings, Sparkles, Upload, Users } from "lucide-react";
import { useProjectRole } from "@/components/role-selector";
import { cn } from "@/lib/utils";
import type { ProjectRole } from "@/lib/types";

const items: Array<{ href: string; label: string; icon: ComponentType<{ className?: string }>; roles: ProjectRole[] }> = [
  { href: "dashboard", label: "Dashboard", icon: BarChart3, roles: ["ml_engineer", "data_analyst", "admin"] },
  { href: "upload", label: "Analyze", icon: Upload, roles: ["ml_engineer"] },
  { href: "explorer", label: "Explorer", icon: Search, roles: ["ml_engineer", "annotator", "domain_expert", "data_analyst"] },
  { href: "probabilistic", label: "Probabilistic", icon: Brain, roles: ["ml_engineer", "data_analyst"] },
  { href: "queue", label: "Review Queue", icon: ListChecks, roles: ["annotator", "domain_expert", "ml_engineer"] },
  { href: "class-action-plan", label: "Class Plan", icon: ClipboardList, roles: ["ml_engineer"] },
  { href: "collection-tasks", label: "Collection", icon: Map, roles: ["data_analyst", "ml_engineer"] },
  { href: "synthetic-tasks", label: "Synthetic", icon: Sparkles, roles: ["ml_engineer", "domain_expert"] },
  { href: "strategies", label: "Strategies", icon: FlaskConical, roles: ["ml_engineer", "data_analyst"] },
  { href: "recommendations", label: "Recommendations", icon: Lightbulb, roles: ["ml_engineer", "data_analyst"] },
  { href: "roadmap", label: "Roadmap", icon: Map, roles: ["ml_engineer", "data_analyst"] },
  { href: "agent", label: "AI Summary", icon: Bot, roles: ["ml_engineer", "annotator", "domain_expert", "data_analyst", "admin"] },
  { href: "members", label: "Members", icon: Users, roles: ["admin"] },
  { href: "export", label: "Export", icon: Download, roles: ["admin", "data_analyst"] },
  { href: "settings", label: "Settings", icon: Settings, roles: ["admin"] },
];

export function ProjectNav({ projectId }: { projectId: string }) {
  const pathname = usePathname();
  const { role } = useProjectRole();
  const visibleItems = items.filter((item) => item.roles.includes(role));
  return (
    <div className="mb-5 border-b">
      <nav className="flex flex-wrap gap-x-1 gap-y-0">
        {visibleItems.map((item) => {
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
