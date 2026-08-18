# Changelog

## [1.1.0](https://github.com/Its-Satyajit/reqly/compare/v1.0.0...v1.1.0) (2026-08-18)


### Features

* **auth:** add digest challenge/response scheme ([f06f0f6](https://github.com/Its-Satyajit/reqly/commit/f06f0f6a424cf6241f30c876951b0d7f67df0d49))
* **auth:** add JWT signing scheme ([51cd692](https://github.com/Its-Satyajit/reqly/commit/51cd692a9b6cbea1c8fefff8858f420097756c7a))
* **auth:** add oauth2 Client Credentials scheme with TokenSource pre-flight and masking ([c8120bc](https://github.com/Its-Satyajit/reqly/commit/c8120bc1c3a5d9dd1a7ced606695de880eae0811))
* **auth:** custom-scheme redirect transport for the auth-code flow ([9c68a91](https://github.com/Its-Satyajit/reqly/commit/9c68a91d47fc99d149b5394329613408996af48a))
* **auth:** introduce scheme registry for authentication ([a9f5274](https://github.com/Its-Satyajit/reqly/commit/a9f5274459b05985684be276e7e867126563cbd1))
* **auth:** mask auth credential values in CLI output ([265b378](https://github.com/Its-Satyajit/reqly/commit/265b3783444422ee20ff268120781a7f2d63b9ef))
* **auth:** OAuth 2.0 Authorization Code + PKCE, auth login, refresh-token reuse ([230cc8a](https://github.com/Its-Satyajit/reqly/commit/230cc8a197bd68183f391ae55ab9fbb07d03e516))
* **auth:** OAuth 2.0 Client Credentials — token cache, refresh, auth CLI, docs ([e306023](https://github.com/Its-Satyajit/reqly/commit/e306023e334766effb4721fe8ab757ffc19ce5fb))
* **auth:** OAuth 2.0 Device flow (RFC 8628) with reqly auth login --flow device ([00cd66c](https://github.com/Its-Satyajit/reqly/commit/00cd66cfb61b00a248ebf559629731449fbb3955))
* **desktop:** body-type editors in the request builder (Milestone 14 T2) ([3cae18b](https://github.com/Its-Satyajit/reqly/commit/3cae18b5d7b10d44f7cfac43f13c628a5ae11146))
* **desktop:** cookies view + milestone 14 docs (T5) ([76a948a](https://github.com/Its-Satyajit/reqly/commit/76a948a6f1423549a0d115802afa328ab8e6146e))
* **desktop:** cookies view + milestone 14 docs (T5) ([660921a](https://github.com/Its-Satyajit/reqly/commit/660921a101a3c864fd552fab84dd16ead0a8156a))
* **desktop:** OAuth auth in the desktop app — core AuthService, bridge, auth panel ([0532fa2](https://github.com/Its-Satyajit/reqly/commit/0532fa2660b675a384d62b67d775154a0e6421a1))
* **desktop:** request builder — params & headers tabs (Milestone 14 T1) ([495f184](https://github.com/Its-Satyajit/reqly/commit/495f184591ba7beb2b0520d95a3eaf04de4fdf8a))
* **desktop:** request builder — params & headers tabs (Milestone 14 T1) ([2090025](https://github.com/Its-Satyajit/reqly/commit/209002509abc848f7222f08189e5b8e2775887b0))
* **desktop:** response actions + JSONPath query bar (Milestone 14 T4) ([7b3e3a8](https://github.com/Its-Satyajit/reqly/commit/7b3e3a8ab2c49ef1856280d4871df9007758fde8))
* **desktop:** response viewer — raw/pretty/tree tabs + search (Milestone 14 T3) ([a034138](https://github.com/Its-Satyajit/reqly/commit/a0341385817d0736c2c96e095a0e50cd16a6d298))
* **secrets:** add Store interface with atomic 0600 file backend ([a28c380](https://github.com/Its-Satyajit/reqly/commit/a28c380b16e31f351df6ec32109f8bfec39c20e0))
* **secrets:** OS-keychain token store backend with --store/REQLY_TOKEN_STORE selection ([fc8548a](https://github.com/Its-Satyajit/reqly/commit/fc8548a3ae8056fac82aeffa05301756afc6e285))


### Bug Fixes

* **auth:** address OAuth code-review findings ([c36bd5a](https://github.com/Its-Satyajit/reqly/commit/c36bd5a755d5705c579a93281f21ed5784a5010b))
* **auth:** validate digest algorithm and apikey location up front ([9af6bae](https://github.com/Its-Satyajit/reqly/commit/9af6bae53ab1a9578a8940d8e11f36ea351d7f37))

## 1.0.0 (2026-08-16)


### Features

* add test engine with assertions, JSONPath, and reqly test command ([3746e20](https://github.com/Its-Satyajit/reqly/commit/3746e205be0f456e884df38b964100f622f432bb))
* add test engine, assertions, and reqly test command ([990d3fd](https://github.com/Its-Satyajit/reqly/commit/990d3fd908f2be9e5e5e00312fa8a601da6ae896))
* **collections:** Git-native workspaces with inheritance + collection CLI ([5f58d3d](https://github.com/Its-Satyajit/reqly/commit/5f58d3d4e57c6546456ea2aa2ac6fc33a3c2598e))
* **collections:** Git-native workspaces with inheritance + collection CLI ([5b9e778](https://github.com/Its-Satyajit/reqly/commit/5b9e778a1cb7e22aefbc50df2d0ef44284276e3d))
* **core:** validation and structural diff engines ([f1183df](https://github.com/Its-Satyajit/reqly/commit/f1183df3a39c6f2c7504a84efab829d3a22419cf))
* **desktop:** wire RequestEditor to the Go core over the Wails bridge ([9e5bf5b](https://github.com/Its-Satyajit/reqly/commit/9e5bf5b38c2a00d12d8a520d3360b7cc83cc9020))
* **desktop:** wire RequestEditor to the Go core over the Wails bridge ([5a1b33f](https://github.com/Its-Satyajit/reqly/commit/5a1b33f21ceacba1c29d07158babaf4cd828f0d9))
* **environments:** Git-native environments, selection, masking, validate, diff ([fe5859e](https://github.com/Its-Satyajit/reqly/commit/fe5859efd0f6f27fa529c3ef576c3cebf4acb817))
* **environments:** Git-native environments, selection, masking, validate, diff ([040ca43](https://github.com/Its-Satyajit/reqly/commit/040ca431465a4e1e9ff47c1e088500c908995e5d))
* **import-export:** cURL/OpenAPI import + Postman collection export ([e2c62b7](https://github.com/Its-Satyajit/reqly/commit/e2c62b75de52cb7a47f8cd56085c63e134056ccc))
* **import-export:** cURL/OpenAPI import + Postman collection export ([83a80a0](https://github.com/Its-Satyajit/reqly/commit/83a80a03aadfcf0042913298d4e52c6c5e27c4f8))
* **mock:** serve a mock API from OpenAPI specs with schema-driven responses ([907d3c7](https://github.com/Its-Satyajit/reqly/commit/907d3c7cdf7cf8a746401bdf0d6b95bdbf28622a))
* **realtime:** WebSocket client + SSE stream client with CLI ([806ce5b](https://github.com/Its-Satyajit/reqly/commit/806ce5b9ffcdafd87a2a587fb88f54373f40fc46))
* **realtime:** WebSocket client + SSE stream client with CLI ([c718d8b](https://github.com/Its-Satyajit/reqly/commit/c718d8bb4a5d48d32c4fb028e8d3fe6c0aeacedb))
* **scripting:** collection runner with pre/post scripts, chaining, and reqly.test assertions ([1c5eea6](https://github.com/Its-Satyajit/reqly/commit/1c5eea697b447f93486c26eae828c6daf5a7e362))
* **scripting:** collection runner with pre/post scripts, chaining, and reqly.test assertions ([6eb5456](https://github.com/Its-Satyajit/reqly/commit/6eb5456190ba9aed53b4c4d7598d6ebd7b8f6a8f))


### Bug Fixes

* **cli:** serialize ws output writes to avoid a data race ([ff54d48](https://github.com/Its-Satyajit/reqly/commit/ff54d4831bec75b32b4e09fbe3e4308a417bdfac))
