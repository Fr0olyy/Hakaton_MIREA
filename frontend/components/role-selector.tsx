"use client";

import { useEffect, useState } from "react";
import { Shield } from "lucide-react";
import { roleOptions } from "@/lib/level2";
import type { ProjectRole } from "@/lib/types";
import { Select } from "@/components/ui/select";

const storageKey = "dataforge-role";

export function useProjectRole() {
  const [role, setRoleState] = useState<ProjectRole>("ml_engineer");

  useEffect(() => {
    const stored = window.localStorage.getItem(storageKey) as ProjectRole | null;
    if (stored) setRoleState(stored);
    const onStorage = () => {
      const next = window.localStorage.getItem(storageKey) as ProjectRole | null;
      if (next) setRoleState(next);
    };
    window.addEventListener("storage", onStorage);
    window.addEventListener("dataforge-role-change", onStorage);
    return () => {
      window.removeEventListener("storage", onStorage);
      window.removeEventListener("dataforge-role-change", onStorage);
    };
  }, []);

  function setRole(nextRole: ProjectRole) {
    window.localStorage.setItem(storageKey, nextRole);
    setRoleState(nextRole);
    window.dispatchEvent(new Event("dataforge-role-change"));
  }

  return { role, setRole };
}

export function RoleSelector() {
  const { role, setRole } = useProjectRole();
  return (
    <label className="flex items-center gap-2 text-sm text-muted-foreground">
      <Shield className="h-4 w-4" />
      <Select className="h-8 w-[190px]" value={role} onChange={(event) => setRole(event.target.value as ProjectRole)} aria-label="Current role">
        {roleOptions.map((option) => (
          <option key={option.value} value={option.value}>
            {option.label}
          </option>
        ))}
      </Select>
    </label>
  );
}
