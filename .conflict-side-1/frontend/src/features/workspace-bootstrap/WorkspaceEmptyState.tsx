/* eslint-disable react/no-children-prop */
import { useForm } from "@tanstack/react-form";
import * as z from "zod";
import { FolderOpen, FolderPlus } from "lucide-react";
import logoDark from "../../assets/logo-dark.svg";
import logoLight from "../../assets/logo-light.svg";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Alert, AlertDescription } from "#components/ui/alert";
import { useThemeStore } from "#stores/useThemeStore";
import { useWorkspaceBootstrapStore } from "#stores/useWorkspaceBootstrap";
import { Field, FieldError, FieldGroup, FieldLabel } from "#components/ui/field";

const formSchema = z.object({
  name: z.string().min(1, "Workspace name is required."),
});

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
    validators: { onSubmit: formSchema },
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
        <form
          onSubmit={(e) => {
            e.preventDefault();
            void form.handleSubmit();
          }}
        >
          <FieldGroup>
          <form.Field
            name="name"
            children={(field) => {
              const isInvalid =
                field.state.meta.isTouched && !field.state.meta.isValid;
              return (
                <Field data-invalid={isInvalid}>
                  <FieldLabel htmlFor="workspace-name">Workspace name</FieldLabel>
                  <Input
                    id="workspace-name"
                    name={field.name}
                    value={field.state.value}
                    onBlur={field.handleBlur}
                    onChange={(e) => field.handleChange(e.target.value)}
                    placeholder={pendingCreate.suggestedName || "my-workspace"}
                    autoFocus
                    aria-invalid={isInvalid}
                  />
                  {isInvalid && <FieldError errors={field.state.meta.errors} />}
                  <p className="text-xs text-muted-foreground">
                    Created inside{" "}
                    <span className="font-mono">{pendingCreate.dir}</span>
                  </p>
                </Field>
              );
            }}
          />
          <div className="flex justify-end gap-2">
            <Button variant="outline" size="sm" onClick={cancelPendingCreate}>
              Back
            </Button>
            <form.Subscribe
              selector={(s) => ({ canSubmit: s.canSubmit })}
            >
              {({ canSubmit }: { canSubmit: boolean }) => (
                <Button
                  type="submit"
                  size="sm"
                  disabled={busy || !canSubmit}
                >
                  Create workspace
                </Button>
              )}
            </form.Subscribe>
          </div>
          </FieldGroup>
        </form>
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
