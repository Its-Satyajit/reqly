import { type LucideIcon } from "lucide-react";
import { cn } from "#lib/utils";

interface PageHeaderProps {
  icon?: LucideIcon;
  title: string;
  description?: string;
  actions?: React.ReactNode;
  className?: string;
}

/** PageHeader — consistent tool page header following spec §61 shared patterns.
 *  Renders a hairline-bordered bar with an optional icon, title, description
 *  and an optional right-side actions slot. */
export function PageHeader({
  icon: Icon,
  title,
  description,
  actions,
  className,
}: PageHeaderProps) {
  return (
    <div
      className={cn(
        "flex shrink-0 items-center gap-2.5 border-b border-border px-4 py-2.5",
        className,
      )}
    >
      {Icon && (
        <Icon
          className="size-[18px] shrink-0 text-muted-foreground"
          aria-hidden
        />
      )}
      <div className="min-w-0 flex-1">
        <h2 className="text-sm font-semibold leading-tight">{title}</h2>
        {description && (
          <p className="mt-0.5 text-[11px] text-muted-foreground">
            {description}
          </p>
        )}
      </div>
      {actions && (
        <div className="flex shrink-0 items-center gap-1">{actions}</div>
      )}
    </div>
  );
}
