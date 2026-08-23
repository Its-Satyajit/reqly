(function () {
  const now = Date.now();
  const ago = (m) => new Date(now - m * 60000).toISOString();

  const workspaces = [
    { id: 'ws-paycore', name: 'PayCore API', path: '~/dev/paycore-ws', branch: 'main', requests: 42, envs: 3 },
    { id: 'ws-petstore', name: 'Petstore Sandbox', path: '~/dev/petstore-ws', branch: 'main', requests: 20, envs: 2 },
    { id: 'ws-edge', name: 'Edgegrid Experiments', path: '~/dev/edge-ws', branch: 'feat/webhooks', requests: 9, envs: 1 }
  ];

  const collectionTree = [
    {
      type: 'folder', id: 'f-auth', name: 'Authentication', open: true, auth: { type: 'oauth2' }, children: [
        { type: 'request', id: 'r-oauth-token', name: 'Get OAuth token', method: 'POST', url: '{{baseUrl}}/oauth/token', saved: true },
        { type: 'request', id: 'r-refresh', name: 'Refresh access token', method: 'POST', url: '{{baseUrl}}/oauth/refresh', saved: true },
        { type: 'folder', id: 'f-edge', name: 'Akamai EdgeGrid', open: false, children: [
          { type: 'request', id: 'r-edge-purge', name: 'Purge cache', method: 'POST', url: '{{edgeHost}}/ccu/v3/invalidate', saved: true }
        ] }
      ]
    },
    {
      type: 'folder', id: 'f-payments', name: 'Payments', open: true, children: [
        { type: 'request', id: 'r-list-payments', name: 'List payments', method: 'GET', url: '{{baseUrl}}/v1/payments?limit={{limit}}', saved: true },
        { type: 'request', id: 'r-create-payment', name: 'Create payment', method: 'POST', url: '{{baseUrl}}/v1/payments', saved: true },
        { type: 'request', id: 'r-get-payment', name: 'Get payment by id', method: 'GET', url: '{{baseUrl}}/v1/payments/{{paymentId}}', saved: false },
        { type: 'request', id: 'r-refund', name: 'Refund payment', method: 'POST', url: '{{baseUrl}}/v1/payments/{{paymentId}}/refund', saved: true },
        { type: 'folder', id: 'f-webhooks', name: 'Webhooks', open: false, children: [
          { type: 'request', id: 'r-wh-list', name: 'List webhook endpoints', method: 'GET', url: '{{baseUrl}}/v1/webhooks', saved: true },
          { type: 'request', id: 'r-wh-test', name: 'Send test event', method: 'POST', url: '{{baseUrl}}/v1/webhooks/{{webhookId}}/test', saved: true }
        ] }
      ]
    },
    {
      type: 'folder', id: 'f-customers', name: 'Customers', open: false, children: [
        { type: 'request', id: 'r-customer-get', name: 'Get customer', method: 'GET', url: '{{baseUrl}}/v1/customers/{{customerId}}', saved: true },
        { type: 'request', id: 'r-slow', name: 'Reports export (slow)', method: 'GET', url: '{{baseUrl}}/v1/reports/export/slow', saved: true },
        { type: 'request', id: 'r-flaky', name: 'Legacy charge endpoint', method: 'POST', url: '{{baseUrl}}/v1/legacy/charge/fail', saved: true }
      ]
    },
    { type: 'request', id: 'r-graphql', name: 'GraphQL — customer orders', method: 'POST', url: '{{graphqlUrl}}', protocol: 'graphql', saved: true }
  ];

  const requestDetail = {
    'r-oauth-token': {
      params: [], headers: [
        { key: 'Content-Type', value: 'application/x-www-form-urlencoded', enabled: true },
        { key: 'Accept', value: 'application/json', enabled: true }
      ],
      auth: { type: 'none' },
      body: { type: 'form', form: [
        { key: 'grant_type', value: 'client_credentials', enabled: true },
        { key: 'client_id', value: '{{clientId}}', enabled: true },
        { key: 'client_secret', value: '{{clientSecret}}', enabled: true }
      ] },
      scripts: {
        pre: "reqly.env.set('tokenEpoch', Date.now());\nreqly.console.log('requesting client credentials token');",
        post: "const j = reqly.response.json();\nif (j.access_token) {\n  reqly.env.setSecret('accessToken', j.access_token);\n  reqly.console.log('stored accessToken secret');\n}"
      }
    },
    'r-create-payment': {
      params: [{ key: 'idempotency_key', value: '{{$uuid}}', enabled: true }],
      headers: [
        { key: 'Content-Type', value: 'application/json', enabled: true },
        { key: 'Idempotency-Key', value: '{{$uuid}}', enabled: true },
        { key: 'X-Request-Signature', value: 'sig_9f2c…', enabled: true }
      ],
      auth: { type: 'bearer', token: '{{accessToken}}' },
      body: { type: 'json', text: '{\n  \"amount\": 4250,\n  \"currency\": \"usd\",\n  \"customer\": \"{{customerId}}\",\n  \"payment_method\": \"pm_card_visa\",\n  \"metadata\": {\n    \"order_id\": \"ord_8841\",\n    \"channel\": \"web\"\n  }\n}' },
      scripts: {
        pre: "reqly.vars.set('amount', 4250);\nreqly.console.log('creating payment for', reqly.vars.get('customerId'));",
        post: "reqly.test('payment created', () => reqly.expect(reqly.response.status).toBe(201));\nreqly.test('has id', () => reqly.expect(reqly.response.json().id).toStartWith('pay_'));"
      }
    },
    'r-list-payments': {
      params: [
        { key: 'limit', value: '25', enabled: true },
        { key: 'status', value: 'succeeded', enabled: true },
        { key: 'created[gt]', value: '{{$timestamp}}', enabled: false }
      ],
      headers: [{ key: 'Accept', value: 'application/json', enabled: true }],
      auth: { type: 'bearer', token: '{{accessToken}}' },
      body: { type: 'none' },
      scripts: { pre: '', post: "reqly.console.info('fetched page', reqly.response.headers['x-page']);" }
    }
  };

  const defaultDetail = {
    params: [], headers: [{ key: 'Accept', value: 'application/json', enabled: true }],
    auth: { type: 'bearer', token: '{{accessToken}}' }, body: { type: 'none' },
    scripts: { pre: '', post: '' }
  };

  const responses = {
    ok: { status: 200, statusText: 'OK', time: 148, size: 1846,
      headers: { 'content-type': 'application/json; charset=utf-8', 'x-request-id': 'req_01J8ZK3M9Q', 'x-ratelimit-remaining': '741', 'cache-control': 'no-store', 'x-page': '1' },
      cookies: [{ name: 'session', value: 's%3A9hGk…', domain: 'api.paycore.dev', expires: '2026-09-23', httpOnly: true, secure: true }],
      body: JSON.stringify({ data: [
        { id: 'pay_1Q84xa', amount: 4250, currency: 'usd', status: 'succeeded', customer: 'cus_NffrFe', created: '2026-08-21T14:02:11Z' },
        { id: 'pay_1Q84td', amount: 1200, currency: 'eur', status: 'pending', customer: 'cus_K2xwZa', created: '2026-08-21T13:44:02Z' },
        { id: 'pay_1Q80kz', amount: 9900, currency: 'usd', status: 'refunded', customer: 'cus_Mm19Qa', created: '2026-08-20T09:12:47Z' },
        { id: 'pay_1Q7zwp', amount: 550, currency: 'gbp', status: 'succeeded', customer: 'cus_J81Ldp', created: '2026-08-19T22:31:15Z' },
        { id: 'pay_1Q7ybn', amount: 1875, currency: 'usd', status: 'failed', customer: 'cus_NffrFe', created: '2026-08-19T18:05:59Z' }
      ], has_more: true, next_cursor: 'cur_1Q7ybn' }, null, 2) },
    created: { status: 201, statusText: 'Created', time: 212, size: 402,
      headers: { 'content-type': 'application/json', 'location': '/v1/payments/pay_1Q85ne', 'x-request-id': 'req_01J8ZKD7PW' },
      cookies: [],
      body: JSON.stringify({ id: 'pay_1Q85ne', object: 'payment', amount: 4250, currency: 'usd', status: 'processing', captured: false, created: '2026-08-24T10:31:04Z' }, null, 2) },
    token: { status: 200, statusText: 'OK', time: 96, size: 214,
      headers: { 'content-type': 'application/json' }, cookies: [],
      body: JSON.stringify({ access_token: 'at_9hGkQ2mX…', token_type: 'Bearer', expires_in: 3600, scope: 'payments:read payments:write' }, null, 2) },
    fail: { status: 500, statusText: 'Internal Server Error', time: 381, size: 178,
      headers: { 'content-type': 'application/json' }, cookies: [],
      body: JSON.stringify({ error: { type: 'internal_error', message: 'ledger write failed after debit hold', code: 'ledger_unavailable' } }, null, 2) },
    teapot: { status: 418, statusText: "I'm a Teapot", time: 42, size: 64,
      headers: { 'content-type': 'text/plain' }, cookies: [],
      body: 'short and stout' },
    notFound: { status: 404, statusText: 'Not Found', time: 61, size: 109,
      headers: { 'content-type': 'application/json' }, cookies: [],
      body: JSON.stringify({ error: { type: 'invalid_request_error', message: 'Unknown payment id', code: 'resource_missing' } }, null, 2) },
    large: { status: 200, statusText: 'OK', time: 894, size: 1048576,
      headers: { 'content-type': 'application/json', 'x-truncated-preview': 'true' }, cookies: [],
      largeRows: true,
      body: '' }
  };

  const environments = [
    { id: 'env-dev', name: 'development', color: 'ok', vars: [
      { key: 'baseUrl', value: 'https://api.paycore.dev/v1', enabled: true },
      { key: 'graphqlUrl', value: 'https://api.paycore.dev/graphql', enabled: true },
      { key: 'limit', value: '25', enabled: true },
      { key: 'paymentId', value: 'pay_1Q84xa', enabled: true },
      { key: 'customerId', value: 'cus_NffrFe', enabled: true }
    ], secrets: [
      { key: 'accessToken', value: 'at_live_9hGkQ2mXvB7nR4pL', enabled: true },
      { key: 'clientSecret', value: 'cs_dev_77aa21', enabled: true }
    ] },
    { id: 'env-staging', name: 'staging', color: 'warn', vars: [
      { key: 'baseUrl', value: 'https://staging.paycore.io/v1', enabled: true },
      { key: 'graphqlUrl', value: 'https://staging.paycore.io/graphql', enabled: true },
      { key: 'limit', value: '100', enabled: true },
      { key: 'paymentId', value: 'pay_stg_0042', enabled: true },
      { key: 'customerId', value: 'cus_stg_9001', enabled: true }
    ], secrets: [ { key: 'accessToken', value: 'at_stg_cryptic42', enabled: true } ] },
    { id: 'env-prod', name: 'production', color: 'error', vars: [
      { key: 'baseUrl', value: 'https://api.paycore.com/v1', enabled: true },
      { key: 'graphqlUrl', value: 'https://api.paycore.com/graphql', enabled: true },
      { key: 'limit', value: '50', enabled: true },
      { key: 'paymentId', value: '', enabled: true },
      { key: 'customerId', value: '', enabled: true }
    ], secrets: [] }
  ];

  const globalScope = { vars: [
    { key: 'locale', value: 'en-US', enabled: true },
    { key: 'traceHeader', value: 'X-Trace-Id', enabled: true }
  ] };
  const workspaceScope = { vars: [ { key: 'team', value: 'payments-platform', enabled: true } ] };
  const collectionScope = { vars: [ { key: 'apiVersion', value: '2026-06-30', enabled: true } ] };
  const processEnv = [
    { key: 'REQLY_TOKEN_STORE', value: 'keychain', source: '.env' },
    { key: 'PAYCORE_TIMEOUT_MS', value: '15000', source: '.env' }
  ];
  const dynamicTags = ['{{$uuid}}', '{{$timestamp}}', '{{$isoTimestamp}}', '{{$randomInt}}', '{{$randomEmail}}', '{{$randomFirstName}}', '{{$guid}}'];

  const history = [
    { id: 'h1', ts: ago(3), method: 'POST', url: '{{baseUrl}}/v1/payments', resolved: 'https://api.paycore.dev/v1/payments', status: 201, time: 212, size: 402, env: 'development', requestId: 'r-create-payment' },
    { id: 'h2', ts: ago(9), method: 'GET', url: '{{baseUrl}}/v1/payments?limit=25', resolved: 'https://api.paycore.dev/v1/payments?limit=25', status: 200, time: 148, size: 1846, env: 'development', requestId: 'r-list-payments' },
    { id: 'h3', ts: ago(17), method: 'POST', url: '{{baseUrl}}/oauth/token', resolved: 'https://api.paycore.dev/oauth/token', status: 200, time: 96, size: 214, env: 'development', requestId: 'r-oauth-token' },
    { id: 'h4', ts: ago(34), method: 'POST', url: '{{baseUrl}}/v1/legacy/charge/fail', resolved: 'https://api.paycore.dev/v1/legacy/charge/fail', status: 500, time: 381, size: 178, env: 'staging', requestId: 'r-flaky' },
    { id: 'h5', ts: ago(58), method: 'GET', url: '{{baseUrl}}/v1/reports/export/slow', resolved: 'https://api.paycore.dev/v1/reports/export/slow', status: 0, time: 30000, size: 0, env: 'development', timeout: true, requestId: 'r-slow' },
    { id: 'h6', ts: ago(75), method: 'DELETE', url: '{{baseUrl}}/v1/webhooks/wh_221', resolved: 'https://api.paycore.dev/v1/webhooks/wh_221', status: 204, time: 88, size: 0, env: 'development', requestId: null },
    { id: 'h7', ts: ago(121), method: 'PUT', url: '{{baseUrl}}/v1/customers/cus_NffrFe', resolved: 'https://api.paycore.dev/v1/customers/cus_NffrFe', status: 301, time: 132, size: 96, env: 'staging', requestId: null },
    { id: 'h8', ts: ago(190), method: 'GET', url: '{{baseUrl}}/v1/payments/pay_missing', resolved: 'https://api.paycore.dev/v1/payments/pay_missing', status: 404, time: 61, size: 109, env: 'development', requestId: null },
    { id: 'h9', ts: ago(300), method: 'GET', url: '{{baseUrl}}/v1/payments?cursor=cur_prev', resolved: 'https://api.paycore.dev/v1/payments?cursor=cur_prev', status: 200, time: 171, size: 1922, env: 'development', requestId: 'r-list-payments' },
    { id: 'h10', ts: ago(610), method: 'POST', url: '{{baseUrl}}/v1/payments/pay_1Q84xa/refund', resolved: 'https://api.paycore.dev/v1/payments/pay_1Q84xa/refund', status: 200, time: 240, size: 356, env: 'development', requestId: 'r-refund' },
    { id: 'h11', ts: ago(1440), method: 'GET', url: '{{edgeHost}}/ccu/v3/invalidate', resolved: 'https://api.edgegrid.net/ccu/v3/invalidate', status: 200, time: 305, size: 142, env: 'development', requestId: 'r-edge-purge' },
    { id: 'h12', ts: ago(1500), method: 'PATCH', url: '{{baseUrl}}/v1/webhooks/wh_118', resolved: 'https://api.paycore.dev/v1/webhooks/wh_118', status: 400, time: 74, size: 156, env: 'staging', requestId: null },
    { id: 'h13', ts: ago(2880), method: 'GET', url: '{{baseUrl}}/v1/customers/cus_K2xwZa', resolved: 'https://api.paycore.dev/v1/customers/cus_K2xwZa', status: 200, time: 119, size: 488, env: 'development', requestId: 'r-customer-get' },
    { id: 'h14', ts: ago(4300), method: 'POST', url: 'https://echo.reqly.dev/teapot', resolved: 'https://echo.reqly.dev/teapot', status: 418, time: 42, size: 64, env: 'development', requestId: null }
  ];

  const testSuites = [
    { id: 'ts-payments', name: 'payments.spec', file: 'collections/tests/payments.spec.yaml', tests: [
      { id: 't1', name: 'list payments returns 200', type: 'status', op: 'equals', expected: '200', lastRun: 'pass' },
      { id: 't2', name: 'response is a page', type: 'jsonpath', source: '$.data', op: 'isArray', expected: '', lastRun: 'pass' },
      { id: 't3', name: 'amount is integer cents', type: 'jsonpath', source: '$.data[0].amount', op: 'isNumber', expected: '', lastRun: 'pass' },
      { id: 't4', name: 'latency under 500ms', type: 'performance', op: 'lessThan', expected: '500ms', lastRun: 'pass' },
      { id: 't5', name: 'next cursor present', type: 'jsonpath', source: '$.next_cursor', op: 'exists', expected: '', lastRun: 'fail' },
      { id: 't6', name: 'deprecated field still served', type: 'header', source: 'x-legacy-mode', op: 'exists', expected: '', lastRun: 'skip' }
    ] },
    { id: 'ts-auth', name: 'auth.spec', file: 'collections/tests/auth.spec.yaml', tests: [
      { id: 't7', name: 'token endpoint returns bearer', type: 'jsonpath', source: '$.token_type', op: 'equals', expected: 'Bearer', lastRun: 'pass' },
      { id: 't8', name: 'expires within the hour', type: 'jsonpath', source: '$.expires_in', op: 'lessThan', expected: '3601', lastRun: 'pass' }
    ] }
  ];

  const assertionOps = ['equals', 'notEquals', 'contains', 'exists', 'isNumber', 'isArray', 'isString', 'lessThan', 'greaterThan', 'matches regex', 'schema validates'];
  const assertionTypes = [
    { id: 'status', label: 'Status code' }, { id: 'jsonpath', label: 'JSONPath' },
    { id: 'header', label: 'Response header' }, { id: 'body', label: 'Body text' },
    { id: 'performance', label: 'Performance' }, { id: 'variable', label: 'Variable' }
  ];

  const scriptSnippets = [
    { label: 'reqly.env.set(name, value)', code: "reqly.env.set('baseUrl', 'https://api.paycore.dev/v1');" },
    { label: 'reqly.env.get(name)', code: "const base = reqly.env.get('baseUrl');" },
    { label: 'reqly.env.setSecret(name, value)', code: "reqly.env.setSecret('accessToken', resp.access_token);" },
    { label: 'reqly.request.setHeader(key, value)', code: "reqly.request.setHeader('X-Trace-Id', '{{$uuid}}');" },
    { label: 'reqly.response.json()', code: "const body = reqly.response.json();" },
    { label: 'reqly.expect(value).toBe(x)', code: "reqly.expect(reqly.response.status).toBe(200);" },
    { label: 'reqly.console.log(...)', code: "reqly.console.log('payload size', reqly.response.size);" }
  ];

  const graphqlSchema = {
    queryType: 'Query',
    types: [
      { name: 'Query', kind: 'object', fields: [
        { name: 'customer', type: 'Customer', args: [{ name: 'id', type: 'ID!' }], desc: 'Fetch a single customer by id.' },
        { name: 'payments', type: '[Payment!]!', args: [{ name: 'limit', type: 'Int' }, { name: 'status', type: 'PaymentStatus' }], desc: 'Paginated payment list.' }
      ] },
      { name: 'Customer', kind: 'object', fields: [
        { name: 'id', type: 'ID!' }, { name: 'email', type: 'String!' },
        { name: 'name', type: 'String' }, { name: 'defaultPaymentMethod', type: 'PaymentMethod' },
        { name: 'orders', type: '[Order!]!', args: [{ name: 'first', type: 'Int' }] }
      ] },
      { name: 'Payment', kind: 'object', fields: [
        { name: 'id', type: 'ID!' }, { name: 'amount', type: 'Int!' }, { name: 'currency', type: 'String!' },
        { name: 'status', type: 'PaymentStatus!' }, { name: 'created', type: 'DateTime!' }
      ] },
      { name: 'PaymentStatus', kind: 'enum', fields: [ { name: 'PENDING' }, { name: 'SUCCEEDED' }, { name: 'FAILED' }, { name: 'REFUNDED' } ] },
      { name: 'Mutation', kind: 'object', fields: [
        { name: 'createPayment', type: 'Payment!', args: [{ name: 'input', type: 'CreatePaymentInput!' }] }
      ] }
    ]
  };

  const grpcServices = [
    { name: 'paycore.v1.PaymentService', methods: [
      { name: 'CreatePayment', streaming: 'unary', input: 'CreatePaymentRequest', output: 'Payment' },
      { name: 'GetPayment', streaming: 'unary', input: 'GetPaymentRequest', output: 'Payment' },
      { name: 'StreamPaymentEvents', streaming: 'server', input: 'StreamEventsRequest', output: 'PaymentEvent' },
      { name: 'UploadReceipts', streaming: 'client', input: 'ReceiptChunk', output: 'UploadSummary' },
      { name: 'ChatWithAgent', streaming: 'bidi', input: 'AgentMessage', output: 'AgentMessage' }
    ] },
    { name: 'paycore.v1.CustomerService', methods: [
      { name: 'GetCustomer', streaming: 'unary', input: 'GetCustomerRequest', output: 'Customer' },
      { name: 'ListCustomers', streaming: 'unary', input: 'ListCustomersRequest', output: 'CustomerPage' }
    ] }
  ];

  const openapiSpecs = [
    {
      id: 'spec-paycore', name: 'PayCore API v1', version: '1.4.2', format: 'OpenAPI 3.0',
      servers: ['https://api.paycore.dev/v1', 'https://staging.paycore.io/v1'],
      security: [
        { scheme: 'bearerAuth', type: 'http', bearerFormat: 'JWT', desc: 'OAuth2 access token via /oauth/token.' },
        { scheme: 'apiKey', type: 'apiKey', in: 'header', name: 'X-Api-Key', desc: 'Server-to-server integrations.' }
      ],
      tags: ['Payments', 'Customers', 'Webhooks'],
      endpoints: [
        { id: 'e1', tag: 'Payments', method: 'GET', path: '/payments', summary: 'List payments', deprecated: false,
          params: [
            { name: 'limit', in: 'query', schema: 'integer (1–100), default 25' },
            { name: 'status', in: 'query', schema: 'enum: succeeded | pending | failed | refunded' },
            { name: 'created[gt]', in: 'query', schema: 'unix timestamp' }
          ],
          requestBody: null,
          responses: [
            { code: '200', desc: 'A page of payments', example: '{"data":[{"id":"pay_1Q84xa","amount":4250,"currency":"usd","status":"succeeded"}],"has_more":true}' },
            { code: '401', desc: 'Invalid or expired token' }
          ] },
        { id: 'e2', tag: 'Payments', method: 'POST', path: '/payments', summary: 'Create a payment', deprecated: false,
          params: [{ name: 'Idempotency-Key', in: 'header', schema: 'string (uuid)' }],
          requestBody: { contentType: 'application/json', required: true, props: [
            { name: 'amount', type: 'integer', required: true, desc: 'Amount in minor units' },
            { name: 'currency', type: 'string', required: true, desc: 'ISO 4217 alpha-3' },
            { name: 'customer', type: 'string', required: true, desc: 'Customer id (cus_…)' },
            { name: 'payment_method', type: 'string', required: false },
            { name: 'metadata', type: 'object<string,string>', required: false }
          ] },
          responses: [
            { code: '201', desc: 'Payment created', example: '{"id":"pay_1Q85ne","amount":4250,"currency":"usd","status":"processing"}' },
            { code: '402', desc: 'Card declined' }
          ] },
        { id: 'e3', tag: 'Payments', method: 'GET', path: '/payments/{id}', summary: 'Retrieve a payment', deprecated: false,
          params: [{ name: 'id', in: 'path', required: true, schema: 'string, pay_ prefix' }],
          requestBody: null,
          responses: [ { code: '200', desc: 'The payment', example: '{"id":"pay_1Q84xa","status":"succeeded"}' }, { code: '404', desc: 'Not found' } ] },
        { id: 'e4', tag: 'Payments', method: 'POST', path: '/payments/{id}/refund', summary: 'Refund a settled payment', deprecated: false,
          params: [{ name: 'id', in: 'path', required: true, schema: 'string' }],
          requestBody: { contentType: 'application/json', props: [ { name: 'amount', type: 'integer', required: false, desc: 'Partial refund when set' }, { name: 'reason', type: 'string', required: false, enum: 'duplicate | fraudulent | requested_by_customer' } ] },
          responses: [ { code: '200', desc: 'Refund record' } ] },
        { id: 'e5', tag: 'Customers', method: 'GET', path: '/customers/{id}', summary: 'Retrieve a customer', deprecated: false,
          params: [{ name: 'id', in: 'path', required: true, schema: 'string, cus_ prefix' }], requestBody: null,
          responses: [ { code: '200', desc: 'The customer' } ] },
        { id: 'e6', tag: 'Customers', method: 'GET', path: '/customers/{id}/orders', summary: 'List customer orders', deprecated: true,
          params: [{ name: 'id', in: 'path', required: true, schema: 'string' }], requestBody: null,
          responses: [ { code: '200', desc: 'Deprecated since 1.3 — use GraphQL orders field' } ] },
        { id: 'e7', tag: 'Webhooks', method: 'GET', path: '/webhooks', summary: 'List webhook endpoints', deprecated: false, params: [], requestBody: null,
          responses: [ { code: '200', desc: 'Endpoints list' } ] },
        { id: 'e8', tag: 'Webhooks', method: 'POST', path: '/webhooks/{id}/test', summary: 'Trigger test event', deprecated: false,
          params: [{ name: 'id', in: 'path', required: true, schema: 'string, wh_ prefix' }],
          requestBody: { contentType: 'application/json', props: [ { name: 'event_type', type: 'string', required: true } ] },
          responses: [ { code: '202', desc: 'Event queued' } ] }
      ],
      schemas: [
        { name: 'Payment', props: [
          { name: 'id', type: 'string' }, { name: 'object', type: 'string', const: 'payment' },
          { name: 'amount', type: 'integer' }, { name: 'currency', type: 'string' },
          { name: 'status', type: 'string', enum: 'succeeded | pending | failed | refunded' },
          { name: 'customer', type: 'string' }, { name: 'created', type: 'string(date-time)' } ] },
        { name: 'Customer', props: [ { name: 'id', type: 'string' }, { name: 'email', type: 'string' }, { name: 'name', type: 'string|null' }, { name: 'balance', type: 'integer' } ] },
        { name: 'Error', props: [ { name: 'error.type', type: 'string' }, { name: 'error.message', type: 'string' }, { name: 'error.code', type: 'string' } ] }
      ]
    },
    {
      id: 'spec-shipping', name: 'Shipping Partners API', version: '0.9.0-beta', format: 'OpenAPI 3.1',
      servers: ['https://partners.shipex.io/v0'], security: [ { scheme: 'apiKey', type: 'apiKey', in: 'header', name: 'ShipEx-Key', desc: 'Partner key' } ],
      tags: ['Shipments'],
      endpoints: [
        { id: 's1', tag: 'Shipments', method: 'GET', path: '/shipments', summary: 'List shipments', params: [], requestBody: null, responses: [ { code: '200', desc: 'Shipments page' } ] },
        { id: 's2', tag: 'Shipments', method: 'POST', path: '/shipments', summary: 'Book a shipment', params: [], requestBody: { contentType: 'application/json', props: [ { name: 'from_postal', type: 'string', required: true }, { name: 'to_postal', type: 'string', required: true }, { name: 'weight_g', type: 'integer', required: true } ] }, responses: [ { code: '201', desc: 'Shipment booked' } ] }
      ],
      schemas: [ { name: 'Shipment', props: [ { name: 'id', type: 'string' }, { name: 'tracking_code', type: 'string' }, { name: 'state', type: 'string', enum: 'booked | picked_up | delivered' } ] } ]
    }
  ];

  const importFormats = [
    { id: 'curl', name: 'cURL', ext: 'terminal command', detect: 'curl -X POST https://api.example.com …' },
    { id: 'openapi', name: 'OpenAPI 3.x', ext: '.yaml .yml .json', detect: 'openapi: 3.0.3' },
    { id: 'har', name: 'HAR 1.2', ext: '.har', detect: '{"log":{"version":"1.2"}}' },
    { id: 'postman', name: 'Postman v2.1', ext: '.json', detect: '{"info":{"schema":"…/v2.1.0"}}' },
    { id: 'insomnia', name: 'Insomnia v4/v5', ext: '.json .yaml', detect: '{"__export_date": …} | yaml doc' },
    { id: 'bruno', name: 'Bruno', ext: 'collection dir .bru', detect: 'bruno collection items tree' }
  ];

  const importReportSample = {
    source: 'Postman — PayCore.postman_collection.json',
    entries: [
      { category: 'translated', severity: 'translated', items: ['pre-request script → preRequest (12 lines)', 'tests → postRequest assertions (5)', 'collection variables → environments/PayCore.yaml (7 keys)', 'basic auth → Auth Editor (inherited)'] },
      { category: 'warnings', severity: 'warned', items: ['formdata file body on “Upload receipt” — pick file after import', 'dynamic variable {{$randomInt}} kept as-is'] },
      { category: 'dropped', severity: 'dropped', items: ['Postman sandbox pm.* API on line 41 — preserved as TODO(reqly-import)'] }
    ],
    counts: { requests: 18, folders: 4, environments: 2, translated: 4, warned: 2, dropped: 1 }
  };

  const mockServers = [
    { id: 'ms1', name: 'PayCore local stub', port: 9090, running: true, uptimeStart: now - 4520000, hitCount: 218,
      endpoints: [
        { id: 'me1', method: 'GET', path: '/v1/payments', status: 200, delay: 120, bodyType: 'json', body: '{\n  "data": [\n    {"id": "pay_mock_001", "amount": 1000, "status": "succeeded"}\n  ],\n  "has_more": false\n}', headers: [{ key: 'content-type', value: 'application/json', enabled: true }] },
        { id: 'me2', method: 'POST', path: '/v1/payments', status: 201, delay: 250, bodyType: 'json', body: '{\n  "id": "pay_mock_new",\n  "status": "processing"\n}', headers: [] },
        { id: 'me3', method: 'GET', path: '/v1/failures/flaky', status: 503, delay: 60, bodyType: 'json', body: '{ "error": { "type": "service_unavailable" } }', headers: [] }
      ],
      logs: [ { t: ago(1), m: 'GET', p: '/v1/payments', s: 200, ms: 118 }, { t: ago(2), m: 'POST', p: '/v1/payments', s: 201, ms: 246 }, { t: ago(4), m: 'GET', p: '/v1/failures/flaky', s: 503, ms: 57 } ] }
  ];

  const diffResult = {
    base: 'PayCore API v1.3.0', target: 'PayCore API v1.4.2',
    stats: { added: 3, removed: 2, changed: 4, breaking: 2 },
    changes: [
      { id: 'c1', severity: 'breaking', kind: 'removed', path: 'GET /customers/{id}/loyalty', detail: 'Endpoint removed without replacement.' },
      { id: 'c2', severity: 'breaking', kind: 'changed', path: 'POST /payments → amount', detail: 'Required field added to request body (amount).' },
      { id: 'c3', severity: 'info', kind: 'added', path: 'POST /payments/{id}/capture', detail: 'New endpoint added.' },
      { id: 'c4', severity: 'info', kind: 'added', path: 'GET /payments → cursor param', detail: 'New optional cursor parameter.' },
      { id: 'c5', severity: 'warn', kind: 'changed', path: 'GET /payments/{id} → status enum', detail: 'Enum narrowed: removed \'authorized\'.' , diff: [['ctx','"status": "succeeded"'],['del','"authorized"'],['add','"captured"']] },
      { id: 'c6', severity: 'info', kind: 'changed', path: 'Payment schema → description', detail: 'Docs updated only.' },
      { id: 'c7', severity: 'breaking', kind: 'removed', path: 'Header X-Ledger-Trace (response)', detail: 'Clients relying on this header will lose tracing.' },
      { id: 'c8', severity: 'warn', kind: 'changed', path: 'Error.code pattern', detail: 'Pattern ^[a-z_]+$ now enforced.' }
    ]
  };

  const gitState = {
    repo: '~/dev/paycore-ws', branch: 'main', upstream: 'origin/main', ahead: 2, behind: 1,
    branches: ['main', 'feat/oauth2-refactor', 'feat/webhook-retries'],
    commits: [
      { hash: 'a3f92c1', msg: 'feat(payments): capture endpoint', who: 'satyajit', when: '2 hours ago' },
      { hash: '7b21de0', msg: 'fix(auth): refresh race on expiry skew', who: 'priya', when: '5 hours ago' },
      { hash: '1c99ab4', msg: 'chore: bump apiVersion to 2026-06-30', who: 'satyajit', when: 'yesterday' },
      { hash: 'f0042ba', msg: 'test: payments.spec cursor case', who: 'ci-bot', when: '2 days ago' }
    ],
    changes: [
      { file: 'collections/payments/create-payment.json', status: 'M', staged: true, lines: '+12 −3' },
      { file: 'collections/payments/get-payment-by-id.json', status: 'M', staged: false, lines: '+4 −4', conflict: false },
      { file: 'environments/dev.yaml', status: 'U', staged: false, lines: '++−−', conflict: true },
      { file: 'collections/webhooks/send-test-event.json', status: 'A', staged: true, lines: '+28' },
      { file: 'collections/tests/auth.spec.yaml', status: 'D', staged: false, lines: '−14' }
    ],
    conflictFile: {
      path: 'environments/dev.yaml',
      ours: ['variables:', '  baseUrl: https://api.paycore.dev/v1', '  limit: 25'],
      theirs: ['variables:', '  baseUrl: https://staging.paycore.io/v1', '  limit: 25']
    }
  };

  const codegenTargets = [
    { lang: 'cURL', clients: ['shell (curl)'] },
    { lang: 'JavaScript', clients: ['fetch', 'axios'] },
    { lang: 'Python', clients: ['requests', 'httpx'] },
    { lang: 'Go', clients: ['net/http'] },
    { lang: 'PHP', clients: ['guzzle'] },
    { lang: 'Rust', clients: ['reqwest'] }
  ];

  const generatedCode = {
    'JavaScript|fetch': `const res = await fetch("https://api.paycore.dev/v1/payments", {
  method: "POST",
  headers: {
    "Authorization": \`Bearer \${process.env.REQLY_MASKED_SECRET}\`,
    "Content-Type": "application/json",
    "Idempotency-Key": crypto.randomUUID(),
  },
  body: JSON.stringify({ amount: 4250, currency: "usd", customer: "cus_NffrFe" }),
});
if (!res.ok) throw new Error(\`HTTP \${res.status}\`);
console.log(await res.json());`,
    'Python|requests': `import os, uuid, requests

res = requests.post(
    "https://api.paycore.dev/v1/payments",
    headers={
        "Authorization": f"Bearer {os.environ['REQLY_MASKED_SECRET']}",
        "Idempotency-Key": str(uuid.uuid4()),
    },
    json={"amount": 4250, "currency": "usd", "customer": "cus_NffrFe"},
)
res.raise_for_status()
print(res.json())`,
    'Go|net/http': `payload := strings.NewReader(\`{"amount":4250,"currency":"usd","customer":"cus_NffrFe"}\`)
req, _ := http.NewRequest(http.MethodPost, "https://api.paycore.dev/v1/payments", payload)
req.Header.Set("Authorization", "Bearer "+os.Getenv("REQLY_MASKED_SECRET"))
req.Header.Set("Content-Type", "application/json")
res, err := http.DefaultClient.Do(req)
if err != nil { log.Fatal(err) }
defer res.Body.Close()`,
    'cURL|shell (curl)': `curl -X POST "https://api.paycore.dev/v1/payments" \\
  -H "Authorization: Bearer $REQLY_MASKED_SECRET" \\
  -H "Content-Type: application/json" \\
  -H "Idempotency-Key: $(uuidgen)" \\
  -d '{"amount":4250,"currency":"usd","customer":"cus_NffrFe"}'`,
    'PHP|guzzle': `$res = $client->post('https://api.paycore.dev/v1/payments', [
  'headers' => ['Authorization' => 'Bearer ' . getenv('REQLY_MASKED_SECRET')],
  'json' => ['amount' => 4250, 'currency' => 'usd', 'customer' => 'cus_NffrFe'],
]);`,
    'Rust|reqwest': `let res = client.post("https://api.paycore.dev/v1/payments")
  .bearer_auth(std::env::var("REQLY_MASKED_SECRET")?)
  .json(&serde_json::json!({"amount": 4250, "currency": "usd", "customer": "cus_NffrFe"}))
  .send().await?;`
  };

  const jwtSample = 'eyJhbGciOiJSUzI1NiIsImtpZCI6IjIwMjYtc2lnLTAxIiwidHlwIjoiSldUIn0.eyJpc3MiOiJodHRwczovL2F1dGgucGF5Y29yZS5kZXYiLCJzdWIiOiJ1c2VyXzEwMSIsImF1ZCI6ImFwaTpwYXljb3JlIiwic2NvcGUiOiJwYXltZW50czpyZWFkIHBheW1lbnRzOndyaXRlIiwiZXhwIjoxNzg3OTM0NDAwLCJpYXQiOjE3NTYwMDAwMDB9.c2FtcGxlX3NpZ25hdHVyZV9ub3RfdmVyaWZpZWRfZGVmZXJyZWQ';

  const bulkInputSample = 'paymentId,amount\ncus_Aaa111,1200\ncus_Bbb222,3400\ncus_Ccc333,999\ncus_Ddd444,1875';

  window.REQLY_DATA = {
    workspaces, collectionTree, requestDetail, defaultDetail, responses,
    environments, globalScope, workspaceScope, collectionScope, processEnv, dynamicTags,
    history, testSuites, assertionTypes, assertionOps, scriptSnippets,
    graphqlSchema, grpcServices, openapiSpecs, importFormats, importReportSample,
    mockServers, diffResult, gitState, codegenTargets, generatedCode, jwtSample, bulkInputSample
  };
})();