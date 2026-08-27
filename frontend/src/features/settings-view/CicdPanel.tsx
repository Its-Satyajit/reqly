import { useState } from "react";
import { useCicdStore } from "#stores/useCicdStore";
import { generateGitHubAction } from "#lib/cicd";
import { Input } from "#components/ui/input";
import { Button } from "#components/ui/button";
import { Copy, Check } from "lucide-react";
import { copyText } from "#lib/response";

export function CicdPanel() {
  const pipeline = useCicdStore((s) => s.pipeline);
  const setPipeline = useCicdStore((s) => s.setPipeline);
  const addSecret = useCicdStore((s) => s.addSecret);
  const removeSecret = useCicdStore((s) => s.removeSecret);
  const cliCommand = useCicdStore((s) => s.getCommand)();

  const [newSecret, setNewSecret] = useState("");
  const [copiedCli, setCopiedCli] = useState(false);
  const [copiedYaml, setCopiedYaml] = useState(false);

  const ghActionYaml = generateGitHubAction(pipeline);

  const handleAddSecret = (e: React.FormEvent) => {
    e.preventDefault();
    if (!newSecret.trim()) return;
    addSecret(newSecret.trim());
    setNewSecret("");
  };

  const handleCopyCli = async () => {
    const ok = await copyText(cliCommand);
    if (ok) {
      setCopiedCli(true);
      setTimeout(() => setCopiedCli(false), 2000);
    }
  };

  const handleCopyYaml = async () => {
    const ok = await copyText(ghActionYaml);
    if (ok) {
      setCopiedYaml(true);
      setTimeout(() => setCopiedYaml(false), 2000);
    }
  };

  return (
    <section className="rounded-lg border border-border bg-card p-4 space-y-4">
      <div>
        <h2 className="text-sm font-semibold">CI / CD & Automated Testing</h2>
        <p className="mt-1 text-xs text-muted-foreground">
          Generate CLI commands and GitHub Actions workflow files for running collection tests in pipelines.
        </p>
      </div>

      <div className="grid grid-cols-1 sm:grid-cols-2 gap-3">
        <div className="flex flex-col gap-1">
          <label htmlFor="cicd-pipeline-name" className="text-[11px] text-muted-foreground">Pipeline Name</label>
          <Input
            id="cicd-pipeline-name"
            value={pipeline.name}
            onChange={(e) => setPipeline({ name: e.target.value })}
            className="h-8 text-xs font-mono"
            placeholder="CI Tests"
          />
        </div>
        <div className="flex flex-col gap-1">
          <label htmlFor="cicd-env-name" className="text-[11px] text-muted-foreground">Environment Name</label>
          <Input
            id="cicd-env-name"
            value={pipeline.environment}
            onChange={(e) => setPipeline({ environment: e.target.value })}
            className="h-8 text-xs font-mono"
            placeholder="production"
          />
        </div>
        <div className="flex flex-col gap-1 sm:col-span-2">
          <label htmlFor="cicd-col-path" className="text-[11px] text-muted-foreground">Collection Path (optional)</label>
          <Input
            id="cicd-col-path"
            value={pipeline.collectionPath ?? ""}
            onChange={(e) => setPipeline({ collectionPath: e.target.value || undefined })}
            className="h-8 text-xs font-mono"
            placeholder="collections/users.json"
          />
        </div>
      </div>

      {/* Secrets mapping */}
      <div className="space-y-2">
        <label className="text-xs font-medium">Mapped GitHub Secrets</label>
        <form onSubmit={handleAddSecret} className="flex gap-2">
          <Input
            value={newSecret}
            onChange={(e) => setNewSecret(e.target.value)}
            placeholder="e.g. API_KEY"
            className="h-8 text-xs font-mono"
          />
          <Button type="submit" size="sm" variant="secondary" className="h-8 text-xs">
            Add Secret
          </Button>
        </form>

        {pipeline.secrets.length > 0 && (
          <div className="flex flex-wrap gap-1.5 pt-1">
            {pipeline.secrets.map((sec) => (
              <span
                key={sec}
                className="inline-flex items-center gap-1 rounded bg-muted px-2 py-0.5 font-mono text-[11px]"
              >
                {sec}
                <button
                  type="button"
                  onClick={() => removeSecret(sec)}
                  className="text-muted-foreground hover:text-foreground"
                >
                  ×
                </button>
              </span>
            ))}
          </div>
        )}
      </div>

      {/* CLI Output */}
      <div className="space-y-1.5">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium">CLI Command</span>
          <Button
            size="sm"
            variant="ghost"
            className="h-6 px-2 text-[11px]"
            onClick={handleCopyCli}
          >
            {copiedCli ? <Check className="size-3 mr-1 text-status-ok" /> : <Copy className="size-3 mr-1" />}
            {copiedCli ? "Copied" : "Copy CLI"}
          </Button>
        </div>
        <pre className="rounded border border-border bg-background p-2.5 font-mono text-xs overflow-x-auto text-foreground">
          {cliCommand}
        </pre>
      </div>

      {/* GitHub Actions YAML Preview */}
      <div className="space-y-1.5">
        <div className="flex items-center justify-between">
          <span className="text-xs font-medium">GitHub Action (.github/workflows/api-tests.yml)</span>
          <Button
            size="sm"
            variant="ghost"
            className="h-6 px-2 text-[11px]"
            onClick={handleCopyYaml}
          >
            {copiedYaml ? <Check className="size-3 mr-1 text-status-ok" /> : <Copy className="size-3 mr-1" />}
            {copiedYaml ? "Copied" : "Copy YAML"}
          </Button>
        </div>
        <pre className="rounded border border-border bg-background p-2.5 font-mono text-xs overflow-x-auto text-foreground max-h-48">
          {ghActionYaml}
        </pre>
      </div>
    </section>
  );
}
