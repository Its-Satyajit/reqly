import type { RequestAuth } from "./request"
import { bytesToBase64 } from "./response"

export type CodeLang = "curl" | "js" | "python" | "go"

function escapeJs(value: string): string {
  return value
    .replace(/\\/g, "\\\\")
    .replace(/'/g, "\\'")
    .replace(/\r\n/g, "\\n")
    .replace(/\n/g, "\\n")
    .replace(/\r/g, "\\n")
}

function escapeGo(value: string): string {
  return value
    .replace(/\\/g, "\\\\")
    .replace(/"/g, '\\"')
    .replace(/\r\n/g, "\\n")
    .replace(/\n/g, "\\n")
    .replace(/\r/g, "\\n")
}

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
    if (req.auth.type === "basic" && req.auth.config?.username) headers.push({ key: "Authorization", value: `Basic ${bytesToBase64(req.auth.config.username + ":" + (req.auth.config.password ?? ""))}` })
  }
  const body = req.body ?? ""
  switch (lang) {
    case "curl": {
      let s = `curl --request ${method} '${url}'`
      for (const h of headers) s += ` --header '${h.key}: ${h.value.replace(/'/g, "'\\''")}'`
      if (body) s += ` --data-raw '${body.replace(/'/g, "'\\''")}'`
      return s
    }
    case "js": {
      let s = `fetch('${escapeJs(url)}', {\n  method: '${method}',\n`
      if (headers.length) { s += "  headers: {\n"; for (const h of headers) s += `    '${escapeJs(h.key)}': '${escapeJs(h.value)}',\n`; s += "  },\n" }
      if (body) s += `  body: '${escapeJs(body)}',\n`
      s += "});"
      return s
    }
    case "python": {
      let s = "import requests\n\n"
      s += `resp = requests.request('${method}', '${escapeJs(url)}'`
      if (headers.length) { s += ", headers={"; s += headers.map((h) => `'${escapeJs(h.key)}': '${escapeJs(h.value)}'`).join(", "); s += "}" }
      if (body) s += `, data='${escapeJs(body)}'`
      s += ")\n"
      return s
    }
    case "go": {
      let s = "package main\n\nimport (\n\t\"context\"\n\t\"net/http\"\n\t\"strings\"\n)\n\n"
      s += "func send(ctx context.Context) (*http.Response, error) {\n"
      s += `\treq, err := http.NewRequestWithContext(ctx, "${method}", "${escapeGo(url)}", strings.NewReader("${escapeGo(body)}"))\n`
      s += "\tif err != nil {\n\t\treturn nil, err\n\t}\n"
      for (const h of headers) s += `\treq.Header.Set("${escapeGo(h.key)}", "${escapeGo(h.value)}")\n`
      s += "\treturn http.DefaultClient.Do(req)\n}"
      return s
    }
    default: return ""
  }
}
