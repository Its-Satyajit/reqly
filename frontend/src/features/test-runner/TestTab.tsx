import { useEffect, useState } from "react";
import { Play, Save } from "lucide-react";
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from "#components/ui/dialog";
import { Alert, AlertDescription } from "#components/ui/alert";
import { Badge } from "#components/ui/badge";
import { Button } from "#components/ui/button";
import { Input } from "#components/ui/input";
import { Spinner } from "#components/ui/spinner";
import { Checkbox } from "#components/ui/checkbox";
import { CodeMirrorEditor } from "../../editors/CodeMirrorEditor";
import {
  TEST_ASSERTION_KIND_OPTIONS,
  type TestAssertion,
} from "#lib/test";
import { useTestStore } from "#stores/useTestStore";

const ASSERTION_DIALOG_KINDS = TEST_ASSERTION_KIND_OPTIONS.map((o) => o.value);

function AssertionBuilder({
  tabId,
}: {
  tabId: string;
}) {
  const content = useTestStore((s) => s.tabs[tabId]?.content ?? "");
  const setContent = useTestStore((s) => s.setContent);
  const [open, setOpen] = useState(false);
  const [draft, setDraft] = useState<TestAssertion>({
    kind: "status",
    expected: 200,
  });

  const append = () => {
    // SAFETY: draft.kind is constrained to the six parser-known kinds.
    if (!ASSERTION_DIALOG_KINDS.includes(draft.kind)) return;
    const lines = buildAssertionLines(draft);
    const updated = insertIntoFirstAssertionsBlock(content, lines);
    setContent(tabId, updated);
    setOpen(false);
  };

  const canAppend = content.includes("assertions:");

  return (
    <>
      <Button variant="outline" size="sm" onClick={() => setOpen(true)}>
        Add assertion
      </Button>
      <Dialog open={open} onOpenChange={setOpen}>
        <DialogContent className="flex w-full flex-col gap-3 sm:max-w-md">
          <DialogHeader>
            <DialogTitle>Add assertion</DialogTitle>
            <DialogDescription>
              Appended to the first test's assertions block in the YAML editor.
            </DialogDescription>
          </DialogHeader>
          <div className="flex flex-col gap-2">
            <label htmlFor="assertion-kind" className="text-xs font-medium">
              Kind
            </label>
            <select
              id="assertion-kind"
              value={draft.kind}
              onChange={(e) => {
                // SAFETY: options are the six parser-known assertion kinds.
                setDraft({ ...draft, kind: e.target.value as TestAssertion["kind"] });
              }}
              className="rounded-md border border-border bg-transparent px-2 py-1 text-xs"
            >
              {TEST_ASSERTION_KIND_OPTIONS.map((o) => (
                <option key={o.value} value={o.value}>
                  {o.label}
                </option>
              ))}
            </select>

            {(draft.kind === "json" || draft.kind === "header") && (
              <>
                <label htmlFor="assertion-path" className="text-xs font-medium">
                  {draft.kind === "json" ? "JSONPath" : "Header name"}
                </label>
                <Input
                  id="assertion-path"
                  value={draft.path ?? ""}
                  onChange={(e) => setDraft({ ...draft, path: e.target.value })}
                  placeholder={draft.kind === "json" ? "$.count" : "content-type"}
                  spellCheck={false}
                  className="font-mono text-xs"
                />
              </>
            )}

            {(draft.kind === "header" ||
              draft.kind === "body_contains" ||
              draft.kind === "body_equals" ||
              (draft.kind === "json")) && (
              <>
                <label htmlFor="assertion-value" className="text-xs font-medium">
                  Expected value
                </label>
                <Input
                  id="assertion-value"
                  value={draft.value ?? ""}
                  onChange={(e) => setDraft({ ...draft, value: e.target.value })}
                  spellCheck={false}
                  className="font-mono text-xs"
                />
              </>
            )}

            {(draft.kind === "status" || draft.kind === "response_time") && (
              <>
                <label htmlFor="assertion-expected" className="text-xs font-medium">
                  {draft.kind === "status" ? "Expected status" : "Max duration (ms)"}
                </label>
                <Input
                  id="assertion-expected"
                  type="number"
                  value={draft.expected ?? ""}
                  onChange={(e) =>
                    setDraft({ ...draft, expected: Number(e.target.value) })
                  }
                  className="font-mono text-xs"
                />
              </>
            )}

            {draft.kind === "json" && (
              <label className="flex items-center gap-2 text-xs">
                <Checkbox
                  checked={draft.exact ?? false}
                  onCheckedChange={(checked) => setDraft({ ...draft, exact: checked })}
                />
                Compare as exact string
              </label>
            )}

            <label className="flex items-center gap-2 text-xs">
              <Checkbox
                checked={draft.not ?? false}
                onCheckedChange={(checked) => setDraft({ ...draft, not: checked })}
              />
              Invert (not)
            </label>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setOpen(false)}>
              Cancel
            </Button>
            <Button onClick={append} disabled={!canAppend}>
              Append to first test
            </Button>
          </DialogFooter>
          {!canAppend && (
            <Alert variant="destructive">
              <AlertDescription>
                No `assertions:` block found — add a test with assertions first.
              </AlertDescription>
            </Alert>
          )}
        </DialogContent>
      </Dialog>
    </>
  );
}

function buildAssertionLines(a: TestAssertion): string[] {
  const lines = [`        - kind: ${a.kind}`];
  if (a.path) lines.push(`          path: ${a.path}`);
  if (a.kind === "status" || a.kind === "response_time") {
    lines.push(`          expected: ${a.expected ?? 0}`);
  }
  if (a.value) lines.push(`          value: ${JSON.stringify(a.value)}`);
  if (a.exact) lines.push(`          exact: true`);
  if (a.not) lines.push(`          not: true`);
  return lines;
}

/** Inserts assertion lines after the last line of the first assertions block. */
function insertIntoFirstAssertionsBlock(content: string, lines: string[]): string {
  const src = content.split("\n");
  let idx = src.findIndex((l) => l.trimEnd().endsWith("assertions:"));
  if (idx === -1) return content;
  idx += 1;
  while (
    idx < src.length &&
    (src[idx].trim() === "" || /^\s*-\s+kind:/.test(src[idx]) || /^\s+(path|value|expected|exact|not):/.test(src[idx]))
  ) {
    idx += 1;
  }
  src.splice(idx, 0, ...lines);
  return src.join("\n");
}

export function TestTab({ tabId }: { tabId: string }) {
  const tab = useTestStore((s) => s.tabs[tabId]);
  const busyAction = tab?.busyAction ?? "none";
  const outcome = tab?.outcome ?? null;
  const error = tab?.error ?? null;
  const dirty = tab?.dirty ?? false;
  const format = tab?.format === "json" ? "json" : "yaml";
  const setContent = useTestStore((s) => s.setContent);
  const save = useTestStore((s) => s.save);
  const run = useTestStore((s) => s.run);
  const openPath = useTestStore((s) => s.openPath);
  const tests = useTestStore((s) => s.tests);

  useEffect(() => {
    if (tab && tab.content === "" && tab.path !== "" && !tab.dirty) {
      void openPath(tabId, tab.path);
    }
  }, [tabId]); // eslint-disable-line react-hooks/exhaustive-deps

  if (!tab) return null;

  return (
    <div className="flex h-full min-h-0 flex-col gap-2 p-3">
      <div className="flex items-center justify-between gap-2">
        <div className="flex min-w-0 items-center gap-2">
          <span className="truncate font-mono text-xs text-muted-foreground">
            {tab.path === "" ? "(unsaved)" : tab.path}
          </span>
          {dirty && <Badge variant="secondary">unsaved</Badge>}
        </div>
        <div className="flex shrink-0 items-center gap-1.5">
          <AssertionBuilder tabId={tabId} />
          <Button
            variant="outline"
            size="sm"
            disabled={busyAction !== "none" || tab.path === ""}
            onClick={() => void save(tabId)}
          >
            {busyAction === "save" ? <Spinner data-icon="inline-start" /> : <Save data-icon="inline-start" />}
            Save
          </Button>
          <Button size="sm" disabled={busyAction !== "none"} onClick={() => void run(tabId)}>
            {busyAction === "run" ? <Spinner data-icon="inline-start" /> : <Play data-icon="inline-start" />}
            Run tests
          </Button>
        </div>
      </div>

      {error && (
        <Alert variant="destructive">
          <AlertDescription>{error}</AlertDescription>
        </Alert>
      )}

      <div className="flex min-h-0 flex-1 gap-2">
        <div className="min-h-0 min-w-0 flex-1 overflow-y-auto rounded-md border border-border">
          <CodeMirrorEditor
            value={tab.content}
            language={format}
            onChange={(v) => setContent(tabId, v)}
            className="h-full"
          />
        </div>
        <div className="min-h-0 w-[38%] overflow-y-auto rounded-md border border-border p-2">
          <p className="pb-1.5 text-xs font-medium uppercase tracking-wide text-muted-foreground">
            Results
          </p>
          {outcome == null && <p className="text-xs text-muted-foreground">Run the suite to see results.</p>}
          {outcome?.error && (
            <Alert variant="destructive">
              <AlertDescription>{outcome.error}</AlertDescription>
            </Alert>
          )}
          {outcome != null && (
            <div className="flex flex-col gap-1 text-xs">
              <p className={outcome.passed ? "font-medium text-status-ok" : "font-medium text-status-error"}>
                {outcome.passCount}/{outcome.total} passed · {outcome.durationMs} ms
              </p>
              {(outcome.results ?? []).map((tr) => (
                <div key={tr.name} className="rounded border border-border px-2 py-1">
                  <p className="flex items-center gap-1.5 font-medium">
                    <span className={tr.passed ? "text-status-ok" : "text-status-error"}>
                      {tr.passed ? "✓" : "✗"}
                    </span>
                    {tr.name}
                  </p>
                  <ul className="mt-0.5 flex flex-col gap-0.5 pl-4 font-mono text-xs text-muted-foreground">
                    {tr.results.map((r, i) => (
                      <li key={i} className={r.passed ? "" : "text-status-error"}>
                        {r.passed ? "✓" : "✗"} {r.message || r.assertion.kind}
                      </li>
                    ))}
                  </ul>
                </div>
              ))}
            </div>
          )}
          {tests.length > 0 && (
            <div className="pt-3">
              <p className="pb-1 text-xs font-medium uppercase tracking-wide text-muted-foreground">
                In workspace ({tests.length})
              </p>
              <ul className="flex flex-col gap-0.5">
                {tests.map((t) => (
                  <li key={t.path}>
                    <button
                      type="button"
                      className="w-full truncate rounded px-1 py-0.5 text-left font-mono text-xs hover:bg-muted/60"
                      title={t.path}
                      onClick={() => void openPath(tabId, t.path)}
                    >
                      {t.path}
                    </button>
                  </li>
                ))}
              </ul>
            </div>
          )}
        </div>
      </div>
    </div>
  );
}
