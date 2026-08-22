import type { RequestAuth } from "./request"

export type CodeLang = "curl" | "js" | "python" | "go"

export function generateCode(req: { method: string; url: string; headers?: { key: string; value: string }[]; query?: { key: string; value: string }[]; body?: string; auth?: RequestAuth }, lang: CodeLang): string {
  const method = req.method || "GET"
  let url = req.url || "https://example.com"
  if (req.query && req.query.length > 0) {
    const q = req.query.map((p) => `${encodeURIComponent(p.key)}=${encodeURIComponent(p.value)}`).join("&")
    url += (url.includes("?") ? "&" : "?") + q
  }
  const headers = [...(req.headers ?? [])]
  if (req.auth) {
    if (req.auth.type === "bearer" && req.auth.config?.token) headers.push({ key: "Authorization", value: `Bearer ${req.auth.config.token}` })
    if (req.auth.type === "apikey" && req.auth.config?.key) {
      const k = req.auth.config.key
      const v = req.auth.config.value ?? ""
      const inQ = req.auth.config.in
      if (inQ === "query") url += (url.includes("?") ? "&" : "?") + `${encodeURIComponent(k)}=${encodeURIComponent(v)}`
      else headers.push({ key: k, value: v })
    }
    if (req.auth.type === "basic" && req.auth.config?.username) headers.push({ key: "Authorization", value: `Basic ${btoa(req.auth.config.username + ":" + (req.auth.config.password ?? ""))}` })
  }
  const body = req.body ?? ""
  switch (lang) {
    case "curl": {
      let s = `curl --request ${method} '${url}'`
      for (const h of headers) s += ` --header '${h.key}: ${h.value}'`
      if (body) s += ` --data-raw '${body.replace(/'/g, "'\\''")}'`
      return s
    }
    case "js": {
      let s = `fetch('${url}', {\n  method: '${method}',\n`
      if (headers.length) { s += "  headers: {\n"; for (const h of headers) s += `    '${h.key}': '${h.value}',\n`; s += "  },\n" }
      if (body) s += `  body: '${body.replace(/'/g, "\\'")}',\n`
      s += "});"
      return s
    }
    case "python": {
      let s = "import requests\n\n"
      s += `resp = requests.request('${method}', '${url}'`
      if (headers.length) { s += ", headers={"; s += headers.map((h) => `'${h.key}': '${h.value}'`).join(", "); s += "}" }
      if (body) s += `, data='${body.replace(/'/g, "\\'")}'`
      s += ")\n"
      return s
    }
    case "go": {
      let s = "package main\n\nimport (\"net/http\"; \"strings\")\n\n"
      s += `req, _ := http.NewRequestWithContext(ctx, "${method}", "${url}", strings.NewReader("${body.replace(/"/g, '\\"')}"))\n`
      for (const h of headers) s += `req.Header.Set("${h.key}", "${h.value}")\n`
      s += "resp, _ := http.DefaultClient.Do(req)\n"
      return s
    }
    default: return ""
  }
}
