import { useState } from "react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from "#components/ui/dialog";
import { Input } from "#components/ui/input";
import { Badge } from "#components/ui/badge";
import { Search, Sparkles } from "lucide-react";
import {
  CATEGORIES,
  searchTemplates,
  instantiateTemplate,
  type RequestTemplate,
} from "#lib/templates";
import { useTemplateStore } from "#stores/useTemplateStore";

interface TemplatePickerSheetProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSelect: (instantiated: ReturnType<typeof instantiateTemplate>) => void;
}

export function TemplatePickerSheet({
  open,
  onOpenChange,
  onSelect,
}: TemplatePickerSheetProps) {
  const [query, setQuery] = useState("");
  const [selectedCategory, setSelectedCategory] = useState<string>("all");
  const customTemplates = useTemplateStore((s) => s.customTemplates);

  const allAvailable = [...CATEGORIES.flatMap((c) => c.templates), ...customTemplates];

  const filteredTemplates = query.trim()
    ? searchTemplates(query)
    : selectedCategory === "all"
    ? allAvailable
    : selectedCategory === "custom"
    ? customTemplates
    : CATEGORIES.find((c) => c.id === selectedCategory)?.templates ?? [];

  const handlePick = (t: RequestTemplate) => {
    const instantiated = instantiateTemplate(t);
    onSelect(instantiated);
    onOpenChange(false);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="flex max-h-[85vh] w-full flex-col gap-3 overflow-hidden sm:max-w-2xl">
        <DialogHeader>
          <div className="flex items-center gap-2">
            <Sparkles className="size-4 text-primary" />
            <DialogTitle>Request Templates</DialogTitle>
          </div>
          <DialogDescription>
            Choose a boilerplate template to populate your request draft.
          </DialogDescription>
        </DialogHeader>

        <div className="flex items-center gap-2 border-b border-border pb-2">
          <div className="relative flex-1">
            <Search className="absolute left-2.5 top-2.5 size-3.5 text-muted-foreground" />
            <Input
              value={query}
              onChange={(e) => setQuery(e.target.value)}
              placeholder="Search templates (e.g. CRUD, GraphQL, Bearer)..."
              className="h-8 pl-8 text-xs font-mono"
            />
          </div>

          <div className="flex items-center gap-1">
            <button
              type="button"
              onClick={() => setSelectedCategory("all")}
              className={`rounded px-2 py-1 text-xs transition-colors ${
                selectedCategory === "all"
                  ? "bg-primary/10 font-medium text-primary"
                  : "text-muted-foreground hover:bg-muted"
              }`}
            >
              All
            </button>
            {customTemplates.length > 0 && (
              <button
                type="button"
                onClick={() => setSelectedCategory("custom")}
                className={`rounded px-2 py-1 text-xs transition-colors ${
                  selectedCategory === "custom"
                    ? "bg-primary/10 font-medium text-primary"
                    : "text-muted-foreground hover:bg-muted"
                }`}
              >
                Custom ({customTemplates.length})
              </button>
            )}
            {CATEGORIES.map((c) => (
              <button
                key={c.id}
                type="button"
                onClick={() => setSelectedCategory(c.id)}
                className={`rounded px-2 py-1 text-xs transition-colors ${
                  selectedCategory === c.id
                    ? "bg-primary/10 font-medium text-primary"
                    : "text-muted-foreground hover:bg-muted"
                }`}
              >
                {c.label}
              </button>
            ))}
          </div>
        </div>

        <div className="min-h-0 flex-1 overflow-y-auto pr-1">
          {filteredTemplates.length === 0 ? (
            <p className="py-8 text-center text-xs text-muted-foreground">
              No templates match your query.
            </p>
          ) : (
            <div className="grid grid-cols-1 gap-2 sm:grid-cols-2">
              {filteredTemplates.map((t) => (
                <button
                  key={t.id}
                  type="button"
                  onClick={() => handlePick(t)}
                  className="flex flex-col items-start gap-1 rounded-md border border-border/80 bg-card p-3 text-left transition-colors hover:border-primary/50 hover:bg-muted/40"
                >
                  <div className="flex w-full items-center justify-between gap-1">
                    <span className="font-semibold text-xs text-foreground">
                      {t.name}
                    </span>
                    <Badge variant="outline" className="text-[10px] uppercase">
                      {t.category}
                    </Badge>
                  </div>
                  <p className="text-[11px] text-muted-foreground line-clamp-2">
                    {t.description}
                  </p>
                  {t.path && (
                    <span className="mt-1 font-mono text-[10px] text-primary/80">
                      {t.method} {t.path}
                    </span>
                  )}
                </button>
              ))}
            </div>
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
