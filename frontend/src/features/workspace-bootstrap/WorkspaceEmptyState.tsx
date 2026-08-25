import { useForm } from "@tanstack/react-form";
import { FolderOpen, FolderPlus } from "lucide-react";
import logoDark from "../../assets/logo-dark.svg";
import logoLight from "../../assets/logo-light.svg";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Alert, AlertDescription } from "#components/ui/alert";
import { useThemeStore } from "#stores/useThemeStore";
import {
  useWorkspaceBootstrapStore,
} from "#stores/useWorkspaceBootstrap";

export function WorkspaceEmptyState() {
  const dark = useThemeStore((s) => s.appearance === "dark");
  const busy = useWorkspaceBootstrapStore((s) => s.busy);
  const error = useWorkspaceBootstrapStore((s) => s.error);
  const pendingCreate = useWorkspaceBootstrapStore((s) => s.pendingCreate);
  const openFolder = useWorkspaceBootstrapStore((s) => s.openFolder);
  const createPending = useWorkspaceBootstrapStore((s) => s.createPending);
  const cancelPendingCreate = useWorkspaceBootstrapStore(
    (s) => s.cancelPendingCreate,
  );
  const form = useForm({
    defaultValues: { name: pendingCreate?.suggestedName ?? "" },
    onSubmit: ({ value }) => void createPending(value.name),
  });

  return (
    <div className="flex min-h-screen flex-col items-center justify-center gap-6 bg-background px-6 text-foreground">
      <img
        src={dark ? logoDark : logoLight}
        alt="Reqly"
        className="h-10 w-auto"
      />
      <div className="flex max-w-md flex-col items-center gap-2 text-center">
        <h1 className="text-lg font-semibold">Open a workspace to begin</h1>
        <p className="text-sm text-muted-foreground">
          Reqly keeps your collections, environments, and history as plain
          files in a folder you own — local-first, versionable with Git. Pick a
          folder with an existing workspace, or create a fresh one.
        </p>
      </div>

      {error && (
        <Alert variant="destructive" className="max-w-md">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      {pendingCreate ? (
        <div
          key={pendingCreate.dir}
          className="flex w-full max-w-sm flex-col gap-3 rounded-lg border border-border p-4"
        >
          <form.Field name="name">
            {(field) => (
              <div className="flex flex-col gap-1.5">
                <label htmlFor="workspace-name" className="text-xs font-medium">
                  Workspace name
                </label>
                <Input
                  id="workspace-name"
                  name={field.name}
                  value={field.state.value}
                  onBlur={field.handleBlur}
                  onChange={(e) => field.handleChange(e.target.value)}
                  placeholder={pendingCreate.suggestedName || "my-workspace"}
                  autoFocus
                />
                <p className="text-xs text-muted-foreground">
                  Created inside{" "}
                  <span className="font-mono">{pendingCreate.dir}</span>
                </p>
              </div>
            )}
          </form.Field>
          <div className="flex justify-end gap-2">
            <Button variant="outline" size="sm" onClick={cancelPendingCreate}>
              Back
            </Button>
            <Button
              size="sm"
              disabled={busy}
              onClick={() => void form.handleSubmit()}
            >
              Create workspace
            </Button>
          </div>
        </div>
      ) : (
        <div className="flex items-center gap-2">
          <Button onClick={() => void openFolder()} disabled={busy}>
            <FolderOpen data-icon="inline-start" />
            Open folder…
          </Button>
          <Button variant="outline" onClick={() => void openFolder()} disabled={busy}>
            <FolderPlus data-icon="inline-start" />
            Create workspace…
          </Button>
        </div>
      )}
    </div>
  );
}
