import { useEffect, useState } from "react"
import { Button } from "../../components/ui/button"
import { useHistoryStore } from "../../stores/useHistoryStore"

export function HistoryView() {
  const { entries, loading, refresh } = useHistoryStore()
  const [query, setQuery] = useState("")
  const [status, setStatus] = useState("")

  useEffect(() => { void refresh() }, [refresh])

  const filtered = entries.filter((e) => {
    if (query && !`${e.url} ${e.requestPath}`.toLowerCase().includes(query.toLowerCase())) return false
    if (status === "2xx" && !(e.status >= 200 && e.status < 300)) return false
    if (status === "4xx" && !(e.status >= 400 && e.status < 500)) return false
    if (status === "5xx" && !(e.status >= 500)) return false
    return true
  })

  return (
    <div className="flex h-full flex-col p-2">
      <div className="flex gap-1 pb-2">
        <input value={query} onChange={(e) => setQuery(e.target.value)} placeholder="Search history…" className="flex-1 rounded-md border border-input bg-background px-2 py-1 text-xs" />
        <select value={status} onChange={(e) => setStatus(e.target.value)} className="rounded-md border border-input bg-background px-2 py-1 text-xs">
          <option value="">All</option><option value="2xx">2xx</option><option value="4xx">4xx</option><option value="5xx">5xx</option>
        </select>
        <Button size="xs" variant="outline" onClick={() => void refresh()} disabled={loading}>{loading ? "…" : "Refresh"}</Button>
      </div>
      <div className="flex-1 overflow-y-auto rounded-md border border-border">
        <table className="w-full text-left text-xs">
          <thead><tr><th className="px-2 py-1">Time</th><th className="px-2 py-1">Method</th><th className="px-2 py-1">URL</th><th className="px-2 py-1">Status</th><th className="px-2 py-1">Duration</th><th className="px-2 py-1">Env</th></tr></thead>
          <tbody>{filtered.map((e) => (<tr key={e.id} className="border-t border-border/50"><td className="px-2 py-1">{new Date(e.createdAt).toLocaleString()}</td><td className="px-2 py-1">{e.method}</td><td className="px-2 py-1 truncate max-w-[240px]">{e.url || e.requestPath}</td><td className="px-2 py-1">{e.status}</td><td className="px-2 py-1">{e.durationMs}ms</td><td className="px-2 py-1">{e.env}</td></tr>))}</tbody>
        </table>
        {filtered.length === 0 ? <p className="p-4 text-xs text-muted-foreground">No history entries.</p> : null}
      </div>
    </div>
  )
}
