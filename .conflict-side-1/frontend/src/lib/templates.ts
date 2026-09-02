export interface RequestTemplate {
  id: string;
  category: "rest" | "graphql" | "grpc" | "realtime";
  subcategory?: string;
  name: string;
  description: string;
  method?: string;
  path?: string;
  headers?: Headers;
  body?: string;
  contentType?: string;
}

export interface TemplateCategory {
  id: string;
  label: string;
  templates: RequestTemplate[];
}

/** Header map — string keys to string values, no known-value widening. */
export type Headers = { [key: string]: string };

export interface InstantiatedRequest {
  method: string;
  path: string;
  headers: Headers;
  body?: string;
}

const REST_TEMPLATES: RequestTemplate[] = [
  {
    id: "rest-crud-list",
    category: "rest",
    subcategory: "CRUD",
    name: "GET /items",
    description: "List all items",
    method: "GET",
    path: "/items",
    headers: { Accept: "application/json" },
  },
  {
    id: "rest-crud-get",
    category: "rest",
    subcategory: "CRUD",
    name: "GET /items/:id",
    description: "Get a single item by ID",
    method: "GET",
    path: "/items/:id",
    headers: { Accept: "application/json" },
  },
  {
    id: "rest-crud-create",
    category: "rest",
    subcategory: "CRUD",
    name: "POST /items",
    description: "Create a new item",
    method: "POST",
    path: "/items",
    headers: { "Content-Type": "application/json" },
    body: '{\n  "name": ""\n}',
  },
  {
    id: "rest-crud-update",
    category: "rest",
    subcategory: "CRUD",
    name: "PUT /items/:id",
    description: "Update an existing item",
    method: "PUT",
    path: "/items/:id",
    headers: { "Content-Type": "application/json" },
    body: '{\n  "name": ""\n}',
  },
  {
    id: "rest-crud-delete",
    category: "rest",
    subcategory: "CRUD",
    name: "DELETE /items/:id",
    description: "Delete an item",
    method: "DELETE",
    path: "/items/:id",
  },
  {
    id: "rest-pagination",
    category: "rest",
    subcategory: "Pagination",
    name: "GET /items?page=1&limit=20",
    description: "Paginated list with page and limit parameters",
    method: "GET",
    path: "/items?page=1&limit=20",
    headers: { Accept: "application/json" },
  },
  {
    id: "rest-upload",
    category: "rest",
    subcategory: "Upload",
    name: "POST /upload (multipart)",
    description: "Upload a file via multipart form-data",
    method: "POST",
    path: "/upload",
    headers: { "Content-Type": "multipart/form-data" },
  },
  {
    id: "rest-auth-bearer",
    category: "rest",
    subcategory: "Authentication",
    name: "GET /me (Bearer token)",
    description: "Authenticated request with Bearer token",
    method: "GET",
    path: "/me",
    headers: { Authorization: "Bearer {{token}}" },
  },
];

const GRAPHQL_TEMPLATES: RequestTemplate[] = [
  {
    id: "gql-query",
    category: "graphql",
    name: "Query",
    description: "GraphQL query request",
    method: "POST",
    path: "/graphql",
    headers: { "Content-Type": "application/json" },
    body: '{\n  "query": "{ __typename }"\n}',
  },
  {
    id: "gql-mutation",
    category: "graphql",
    name: "Mutation",
    description: "GraphQL mutation request",
    method: "POST",
    path: "/graphql",
    headers: { "Content-Type": "application/json" },
    body: '{\n  "query": "mutation { __typename }"\n}',
  },
  {
    id: "gql-subscription",
    category: "graphql",
    name: "Subscription",
    description: "GraphQL subscription (over WebSocket)",
    method: "POST",
    path: "/graphql",
    headers: { "Content-Type": "application/json" },
    body: '{\n  "query": "subscription { __typename }"\n}',
  },
];

const GRPC_TEMPLATES: RequestTemplate[] = [
  {
    id: "grpc-unary",
    category: "grpc",
    name: "Unary call",
    description: "Single request-response gRPC call",
    method: "UNARY",
    path: "/service/method",
    headers: { "content-type": "application/grpc" },
    body: "{}",
  },
  {
    id: "grpc-server-stream",
    category: "grpc",
    name: "Server streaming",
    description: "Client sends one request, server streams responses",
    method: "SERVER_STREAM",
    path: "/service/method",
    headers: { "content-type": "application/grpc" },
  },
  {
    id: "grpc-client-stream",
    category: "grpc",
    name: "Client streaming",
    description: "Client streams requests, server sends one response",
    method: "CLIENT_STREAM",
    path: "/service/method",
    headers: { "content-type": "application/grpc" },
  },
  {
    id: "grpc-bidi",
    category: "grpc",
    name: "Bidirectional streaming",
    description: "Both client and server stream messages",
    method: "BIDI_STREAM",
    path: "/service/method",
    headers: { "content-type": "application/grpc" },
  },
];

const REALTIME_TEMPLATES: RequestTemplate[] = [
  {
    id: "ws-connect",
    category: "realtime",
    name: "WebSocket connect",
    description: "Connect to a WebSocket endpoint",
    method: "WS",
    path: "ws://localhost:8080/socket",
    headers: { Upgrade: "websocket", Connection: "Upgrade" },
  },
  {
    id: "sse-subscribe",
    category: "realtime",
    name: "SSE subscribe",
    description: "Subscribe to a Server-Sent Events stream",
    method: "GET",
    path: "/events",
    headers: { Accept: "text/event-stream" },
  },
];

export const CATEGORIES: TemplateCategory[] = [
  { id: "rest", label: "REST", templates: REST_TEMPLATES },
  { id: "graphql", label: "GraphQL", templates: GRAPHQL_TEMPLATES },
  { id: "grpc", label: "gRPC", templates: GRPC_TEMPLATES },
  { id: "realtime", label: "Realtime", templates: REALTIME_TEMPLATES },
];

const ALL_TEMPLATES = CATEGORIES.flatMap((c) => c.templates);

export function getCategoryList(): TemplateCategory[] {
  return CATEGORIES;
}

export function getTemplatesByCategory(category: RequestTemplate["category"]): RequestTemplate[] {
  const cat = CATEGORIES.find((c) => c.id === category);
  return cat ? cat.templates : [];
}

export function getTemplateById(id: string): RequestTemplate | undefined {
  return ALL_TEMPLATES.find((t) => t.id === id);
}

export function instantiateTemplate(template: RequestTemplate): InstantiatedRequest {
  const headers = { ...template.headers };
  if (template.contentType && !headers["Content-Type"]) {
    headers["Content-Type"] = template.contentType;
  }
  return {
    method: template.method ?? "GET",
    path: template.path ?? "/",
    headers,
    body: template.body,
  };
}

export function searchTemplates(query: string): RequestTemplate[] {
  const lower = query.toLowerCase();
  return ALL_TEMPLATES.filter(
    (t) =>
      t.name.toLowerCase().includes(lower) ||
      t.description.toLowerCase().includes(lower) ||
      (t.subcategory?.toLowerCase().includes(lower) ?? false),
  );
}
