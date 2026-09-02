import { Check, ChevronDown, Database } from "lucide-react";
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from "#components/ui/dropdown-menu";
import { cn } from "#lib/utils";
import { useWorkspaceStore, type Environment } from "#stores/useWorkspaceStore";

/** EnvironmentSelector — spec §2.1 Active Environment dropdown.
 *  Shows the active environment name (or "No Environment") and lets the user
 *  switch from a list of all environments in the workspace. */
export function EnvironmentSelector() {
  const environments = useWorkspaceStore((s) => s.environments);
  const activeEnvironmentId = useWorkspaceStore((s) => s.activeEnvironmentId);
  const setActiveEnvironment = useWorkspaceStore(
    (s) => s.setActiveEnvironment,
  );

  const active: Environment | undefined = environments.find(
    (e) => e.id === activeEnvironmentId,
  );

  return (
    <DropdownMenu>
      <DropdownMenuTrigger
        className="flex h-7 items-center gap-1.5 rounded-md border border-border bg-transparent px-2 text-xs font-normal transition-colors hover:bg-muted focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
        aria-label="Active environment"
      >
        <Database className="size-3.5 shrink-0 text-muted-foreground" aria-hidden />
        <span className="max-w-[100px] truncate">
          {active ? active.name : "No Environment"}
        </span>
        <ChevronDown className="size-3 shrink-0 text-muted-foreground" aria-hidden />
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" className="w-52">
        <DropdownMenuLabel className="text-[11px] font-medium uppercase tracking-wider text-muted-foreground">
          Environments
        </DropdownMenuLabel>
        <DropdownMenuSeparator />
        <DropdownMenuItem
          onClick={() => setActiveEnvironment(null)}
          className="gap-2 text-xs"
        >
          <Check
            className={cn(
              "size-3.5 shrink-0",
              !activeEnvironmentId ? "opacity-100" : "opacity-0",
            )}
            aria-hidden
          />
          No Environment
        </DropdownMenuItem>
        {environments.map((env) => (
          <DropdownMenuItem
            key={env.id}
            onClick={() => setActiveEnvironment(env.id)}
            className="gap-2 text-xs"
          >
            <Check
              className={cn(
                "size-3.5 shrink-0",
                activeEnvironmentId === env.id ? "opacity-100" : "opacity-0",
              )}
              aria-hidden
            />
            <span className="truncate">{env.name}</span>
          </DropdownMenuItem>
        ))}
        {environments.length === 0 && (
          <DropdownMenuItem disabled className="text-xs text-muted-foreground">
            No environments in workspace
          </DropdownMenuItem>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}
