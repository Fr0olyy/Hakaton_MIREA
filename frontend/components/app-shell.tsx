import Link from "next/link";
import { Database, FolderKanban } from "lucide-react";
import { RoleSelector } from "@/components/role-selector";

export function AppShell({ children }: { children: React.ReactNode }) {
  return (
    <div className="min-h-screen">
      <header className="sticky top-0 z-40 border-b bg-card/90 backdrop-blur-md">
        <div className="mx-auto flex h-14 max-w-7xl items-center justify-between px-4">
          <Link href="/projects" className="flex items-center gap-2 font-semibold">
            <span className="flex h-8 w-8 items-center justify-center rounded-md bg-primary text-primary-foreground">
              <Database className="h-4 w-4" />
            </span>
            DataForge AI
          </Link>
          <nav className="flex items-center gap-2 text-sm text-muted-foreground">
            <Link className="flex items-center gap-2 rounded-md px-3 py-2 hover:bg-muted hover:text-foreground" href="/projects">
              <FolderKanban className="h-4 w-4" />
              Projects
            </Link>
            <div className="hidden md:block">
              <RoleSelector />
            </div>
          </nav>
        </div>
      </header>
      <main className="mx-auto max-w-7xl px-4 py-7">{children}</main>
    </div>
  );
}
