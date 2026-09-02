import { useState } from "react";
import { FolderPlus } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogDescription,
  DialogFooter,
} from "#components/ui/dialog";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { useWorkspaceBootstrapStore } from "#stores/useWorkspaceBootstrap";

export function CreateWorkspaceModal() {
  const open = useWorkspaceBootstrapStore((s) => s.createModalOpen);
  const setOpen = useWorkspaceBootstrapStore((s) => s.setCreateModalOpen);
  const busy = useWorkspaceBootstrapStore((s) => s.busy);
  const adapter = useWorkspaceBootstrapStore((s) => s.adapter);
  const createInFolder = useWorkspaceBootstrapStore((s) => s.createInFolder);

  const [dirPath, setDirPath] = useState("");
  const [name, setName] = useState("");
  const [error, setError] = useState<string | null>(null);

  const handlePickFolder = async () => {
    try {
      const selected = await adapter.pickFolder();
      if (selected) {
        setDirPath(selected);
        if (!name) {
          const inferred = selected.split(/[\\/]/).filter(Boolean).pop() ?? "";
          setName(inferred);
        }
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  };

  const handleCreate = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!dirPath) {
      setError("Please select a target folder.");
      return;
    }
    setError(null);
    try {
      await createInFolder(dirPath, name.trim() || undefined);
      setOpen(false);
      setDirPath("");
      setName("");
    } catch (err) {
      setError(err instanceof Error ? err.message : String(err));
    }
  };

  return (
    <Dialog open={open} onOpenChange={setOpen}>
      <DialogContent className="sm:max-w-md">
        <form onSubmit={handleCreate}>
          <DialogHeader>
            <div className="flex items-center gap-2">
              <FolderPlus className="size-4 text-primary" />
              <DialogTitle>Create New Workspace</DialogTitle>
            </div>
            <DialogDescription>
              Scaffold a fresh Git-native Reqly workspace with standard collections, environments, and configuration.
            </DialogDescription>
          </DialogHeader>

          <div className="flex flex-col gap-3 py-4">
            <div className="flex flex-col gap-1.5">
              <label htmlFor="ws-folder" className="text-xs font-medium text-muted-foreground">
                Target Folder Directory
              </label>
              <div className="flex gap-2">
                <Input
                  id="ws-folder"
                  value={dirPath}
                  onChange={(e) => setDirPath(e.target.value)}
                  placeholder="/path/to/folder or choose directory…"
                  className="font-mono text-xs"
                  required
                />
                <Button type="button" variant="outline" size="sm" onClick={() => void handlePickFolder()}>
                  Browse…
                </Button>
              </div>
            </div>

            <div className="flex flex-col gap-1.5">
              <label htmlFor="ws-name" className="text-xs font-medium text-muted-foreground">
                Workspace Name (optional)
              </label>
              <Input
                id="ws-name"
                value={name}
                onChange={(e) => setName(e.target.value)}
                placeholder="My API Workspace"
                className="text-xs"
              />
            </div>

            {error && (
              <p className="text-xs text-destructive bg-destructive/10 border border-destructive/20 rounded p-2">
                {error}
              </p>
            )}
          </div>

          <DialogFooter>
            <Button type="button" variant="outline" size="sm" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button type="submit" size="sm" disabled={busy || !dirPath}>
              Create Workspace
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
