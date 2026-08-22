# Changelog

## [1.3.0](https://github.com/Its-Satyajit/reqly/compare/v1.2.0...v1.3.0) (2026-08-22)


### Features

* add anti-slop rules and CodeMirror typing fix ([25c1b35](https://github.com/Its-Satyajit/reqly/commit/25c1b35f85ab6efe90f30479ed435ed91f15413b))
* **auth:** add AWS SigV4 and EdgeGrid schemes (M20 T1) ([4b8bc6c](https://github.com/Its-Satyajit/reqly/commit/4b8bc6cf61b49b5697a113441ea07bc11b95840b))
* **body:** add binary + GraphQL body types (M21 T1) ([a7fc993](https://github.com/Its-Satyajit/reqly/commit/a7fc9937a1ec15a1c7c1da81b0345ecf19e70f3d))
* **bridge:** expose file auth and pass draft auth on save/send (M19 T2) ([7d1952a](https://github.com/Its-Satyajit/reqly/commit/7d1952a0edc03004ab76ffcfdb7a4d072311c310))
* **bridge:** file-backed send + save for editable request tabs ([5a12b89](https://github.com/Its-Satyajit/reqly/commit/5a12b8946b0bfef86a64ebf0592620c789f5d8c0))
* **core:** collection run service + RunFolder engine support (T2) ([bf28df3](https://github.com/Its-Satyajit/reqly/commit/bf28df3b61b2f82515e8f2453b9ce75af9c0f150))
* **core:** draft auth is editable on save and send (M19 T1) ([efb54d3](https://github.com/Its-Satyajit/reqly/commit/efb54d32786346bdd118dcb2f0d17fb3de4bf3c2))
* **core:** workspace save/re-resolve seams for editable request tabs ([112e659](https://github.com/Its-Satyajit/reqly/commit/112e659c3915bb955b16b8716032008a3c50024f))
* **desktop:** collection-run bindings + streamed Wails events (T3) ([9f145c0](https://github.com/Its-Satyajit/reqly/commit/9f145c0c9e838c9c35252cf8265e90689bf3de0e))
* **desktop:** merge M17 request-file editing into main ([2a24af0](https://github.com/Its-Satyajit/reqly/commit/2a24af013386e2a9a8ef20a40db1804c5f38d3f5))
* **frontend:** collection-run adapter + run store (T4) ([9114501](https://github.com/Its-Satyajit/reqly/commit/9114501a94c2323b1e23509370030b31b6b96b4e))
* **frontend:** editable Auth tab with scheme forms (M19 T3) ([63cf7ed](https://github.com/Its-Satyajit/reqly/commit/63cf7edfb89af5041bced2b37a0d07ad8f96d8e3))
* **frontend:** editable request tabs with save + conflict handling ([91f87ef](https://github.com/Its-Satyajit/reqly/commit/91f87ef799eb0437c141c4e4de62bfb7003420bb))
* **frontend:** inherited auth view, save warnings, roadmap (M19 T5) ([c74b57c](https://github.com/Its-Satyajit/reqly/commit/c74b57c2deded88d1d82164bba1137352826d6c2))
* **frontend:** oauth2 grant form wired to Auth Panel (M19 T4) ([8b1794b](https://github.com/Its-Satyajit/reqly/commit/8b1794b2c5992d144fb49aa9b729891e706968f9))
* **frontend:** register AWS + EdgeGrid in auth schemes (M20 T2) ([380524e](https://github.com/Its-Satyajit/reqly/commit/380524e8a70c1b5603c93a7c5af4678f24f780b1))
* **frontend:** sidebar run buttons + Run View tab (T5) ([779ccdc](https://github.com/Its-Satyajit/reqly/commit/779ccdce6c9daeb82c144eb40c42712267b16056))
* **history:** SQLite history + cookie jar, HistoryService + masking (M22 T1-T2) ([7f67b07](https://github.com/Its-Satyajit/reqly/commit/7f67b07dba68ee9bbc00638e27ffd0944f86bff4))
* **install:** bash for Linux+macOS (amd64/arm64, pacman/apt/dnf/zypper) + ps1 for Windows 64 ([aa4aedb](https://github.com/Its-Satyajit/reqly/commit/aa4aedb38c46a295c53613d734222a3f00590485))
* **jwt:** decode tooling — internal/jwt Decode + reqly jwt decode CLI (M29) ([702c271](https://github.com/Its-Satyajit/reqly/commit/702c271c1ca5c05c47c69c0fe4c9115c9328a79c))
* **jwt:** JWT decode + HAR close + alpha banner (M28/M29 + M30 spec) ([67651e6](https://github.com/Its-Satyajit/reqly/commit/67651e67a9d1fc7fa77b861fd79131118acbf310))
* **m22:** CLI history, desktop history bridge, table + binary preview ([20c48c7](https://github.com/Its-Satyajit/reqly/commit/20c48c7e1e9aba4edec6681fb621f23959273684))
* **m22:** history + cookie jar + table + binary preview ([4125071](https://github.com/Its-Satyajit/reqly/commit/412507176a2053bc612a06e6656a640e784b43da))
* **m23:** dynamic values & template tags ([870bc68](https://github.com/Its-Satyajit/reqly/commit/870bc68bcf276a70571b24cf87a33b784a9c8558))
* **m24:** code generation core + CLI/desktop ([45b06d6](https://github.com/Its-Satyajit/reqly/commit/45b06d6dbd82e78962bd7b7eaa98d05dd4e1f57c))
* **m24:** code generation core + CLI/desktop (M24 T1-T3) ([9575b99](https://github.com/Its-Satyajit/reqly/commit/9575b99243108780e67f779b69d7936bce9be694))
* **m25:** workspace save & export ([dd18e65](https://github.com/Its-Satyajit/reqly/commit/dd18e658e29f5b1b5aba32c2f79aa39b491468be))
* **m25:** workspace save & export (M25 T1-T2) ([bd04b0d](https://github.com/Its-Satyajit/reqly/commit/bd04b0df7d950abd8e2fcac9598d7d26155e37d7))
* **m26:** docs generation ([60056e4](https://github.com/Its-Satyajit/reqly/commit/60056e41a065743f91bdfa798e79a59f847a739b))
* **m26:** docs generation core + CLI (M26 T1-T2) ([eac09d4](https://github.com/Its-Satyajit/reqly/commit/eac09d4b1be05b9c3510211729ddfff099559c66))
* **m27:** cross-platform desktop + install scripts ([a4ae463](https://github.com/Its-Satyajit/reqly/commit/a4ae46313efaa52c7bea3b0acfaa8d476bdc2fa7))
* **pagination:** page/offset/cursor/link-header runner (M30) ([ddcab17](https://github.com/Its-Satyajit/reqly/commit/ddcab17f280e905d2c4475e7016ff23aaf8429fa))
* **pagination:** page/offset/cursor/link-header runner (M30) ([f7e81d6](https://github.com/Its-Satyajit/reqly/commit/f7e81d6fd1e537d9030d81882023205df8b941eb))
* **requestfile:** format-preserving atomic save + content fingerprint ([41dd31f](https://github.com/Its-Satyajit/reqly/commit/41dd31fd137ee7754eb9465639f01cf674add059))
* **runner:** stream per-step results via OnStep callback (T1) ([039c1d2](https://github.com/Its-Satyajit/reqly/commit/039c1d22dd9019ac6e48213097ba8a91b64ee9a7))
* **variables:** dynamic values & template tags (M23 T1-T3) ([08ede34](https://github.com/Its-Satyajit/reqly/commit/08ede34a4b5e5748dc87a9526caa5c3f4c2c14d8))
* **variables:** dynamic values & template tags (M23 T1-T3) ([d566c4e](https://github.com/Its-Satyajit/reqly/commit/d566c4e58ae68d4704f1b603c4cb30cdcc55c4b4))


### Bug Fixes

* **bindings:** correct desktop:bindings output path ([8aacdcf](https://github.com/Its-Satyajit/reqly/commit/8aacdcf0ff511c5b00732484456cdf7496e4175d))
* **body:** file-aware form-data, binary mime, graphql lang, RequestEditor pickers (M21 review) ([9b6404d](https://github.com/Its-Satyajit/reqly/commit/9b6404d91e435456fbf27efd762ee477a39d7c27))
* **bridge:** allow nullable Auth config in normalizeAuth ([0769608](https://github.com/Its-Satyajit/reqly/commit/076960891c7e3599c0a54f63e03a321a6630d76c))
* **ci,bridge:** correct wails bindings generation and strict type errors ([aea296c](https://github.com/Its-Satyajit/reqly/commit/aea296c1aa19b537067ae8acd228495bd5cd2e49))
* **deps:** update module github.com/wailsapp/wails/v3 to v3.0.0-beta.12 ([bf1dff7](https://github.com/Its-Satyajit/reqly/commit/bf1dff7a2c9b9cacb32db0e74ea0e45f8a194d1d))
* **deps:** update module github.com/wailsapp/wails/v3 to v3.0.0-beta.12 ([cfc8c9d](https://github.com/Its-Satyajit/reqly/commit/cfc8c9d0ab20d84959a198c16bdf8abad43239bf))
* **deps:** update wails v3 ([06f7756](https://github.com/Its-Satyajit/reqly/commit/06f77562a78b75ab869e8db2413642159297a07e))
* **deps:** update wails v3 ([f9fdb6a](https://github.com/Its-Satyajit/reqly/commit/f9fdb6a10198709847da4134a8726227bd6f47f6))
* **desktop:** scan shared frontend sources for tailwind classes ([48625ae](https://github.com/Its-Satyajit/reqly/commit/48625ae770dae61dc12201165a0b7b38b04484e8))
* **desktop:** serve frontend from relocated apps/desktop/frontend dir ([1100cc8](https://github.com/Its-Satyajit/reqly/commit/1100cc80cbf43ea84ff884a73b1faa2bc92d1548))
* **frontend:** add missing KeyValueRow import for RequestEditor (m21 typecheck) ([3337314](https://github.com/Its-Satyajit/reqly/commit/3337314134e533f7713c6ebf9b1255e84133413a))
* **frontend:** anti-slop lint 0/0 — typeGuards, bridge split assertions, body GraphQLVariables, response/jsonpath safety ([cf262a0](https://github.com/Its-Satyajit/reqly/commit/cf262a0db0afe42bbfbc5f7d0617d9aaeede43de))
* **jwt:** review — MarshalIndent errors, remove os keep-alive, iat age via Now, IsExpired seam ([37172db](https://github.com/Its-Satyajit/reqly/commit/37172db6500964b86585d74819dd67861db2356a))
* **m17-merge:** restore core import in app_test.go, drop circular RequestAuth import in request.ts ([0ec3d88](https://github.com/Its-Satyajit/reqly/commit/0ec3d88ead230d6e28bdd532ca082b0a7782117f))
* **m17:** save-time warnings + structured changed-on-disk error ([c12a8ef](https://github.com/Its-Satyajit/reqly/commit/c12a8ef30816af0765683fbc9b768311b85ea1af))

## [1.2.0](https://github.com/Its-Satyajit/reqly/compare/v1.1.1...v1.2.0) (2026-08-19)


### Features

* **desktop:** environment delete + inline validation + milestone docs (M15 T5) ([782b86b](https://github.com/Its-Satyajit/reqly/commit/782b86bcf5043c2e72194aa4d970b03b2c4bf466))
* **desktop:** environment delete + inline validation + milestone docs (M15 T5) ([211e413](https://github.com/Its-Satyajit/reqly/commit/211e4132b32863ca722c0c66951f8f832998a665))
* **desktop:** environment editor — description + variables with explicit save (M15 T3) ([5f89315](https://github.com/Its-Satyajit/reqly/commit/5f893156ac60109ccfba1598e2efb89fe7660d7f))
* **desktop:** environment editor — description + variables with explicit save (M15 T3) ([5ba5959](https://github.com/Its-Satyajit/reqly/commit/5ba59596da1d4da79a518a77b2ab500974402192))
* **desktop:** environment service + header selector wiring (M15 T1) ([4553aa7](https://github.com/Its-Satyajit/reqly/commit/4553aa7c1946e8e703bf930cd96759ef1a95da16))
* **desktop:** environment service + header selector wiring (M15 T1) ([bf70ebb](https://github.com/Its-Satyajit/reqly/commit/bf70ebb1a08b64607606084e49e3a8292dac5fda))
* **desktop:** environments view — list, create, set active (M15 T2) ([5dd2332](https://github.com/Its-Satyajit/reqly/commit/5dd2332b1773dafe62ec6d777a1f61d499c40416))
* **desktop:** environments view — list, create, set active (M15 T2) ([a452db0](https://github.com/Its-Satyajit/reqly/commit/a452db026eeb9e6e9002865deb7d9d4ae3fea247))
* **desktop:** masked secrets editing — changed-only writes, reveal toggle, remove (M15 T4) ([aa8ed25](https://github.com/Its-Satyajit/reqly/commit/aa8ed252693d13f74163c6011bab06c4d715e3d9))
* **desktop:** masked secrets editing — changed-only writes, reveal toggle, remove (M15 T4) ([10bcdc2](https://github.com/Its-Satyajit/reqly/commit/10bcdc2c70ac33e62087538c7bb9de19f3e895b8))
* **m16:** desktop collections browser ([#135](https://github.com/Its-Satyajit/reqly/issues/135)) ([ad4339d](https://github.com/Its-Satyajit/reqly/commit/ad4339d094678a93a1e19274fa9be524397db755))


### Bug Fixes

* **deps:** update module github.com/coder/websocket to v1.8.15 ([e88a544](https://github.com/Its-Satyajit/reqly/commit/e88a544fa332c412255fc0efc438505dcf89b921))
* **deps:** update module github.com/coder/websocket to v1.8.15 ([e88a544](https://github.com/Its-Satyajit/reqly/commit/e88a544fa332c412255fc0efc438505dcf89b921))
* **deps:** update module github.com/coder/websocket to v1.8.15 ([379f791](https://github.com/Its-Satyajit/reqly/commit/379f791a59a24a5537af06030e1fdddccddee7f6))
* **deps:** update module github.com/getkin/kin-openapi to v0.147.0 ([81be1b0](https://github.com/Its-Satyajit/reqly/commit/81be1b02fd24d70843a9fb3b5508e784cc68f113))
* **deps:** update module github.com/getkin/kin-openapi to v0.147.0 ([81be1b0](https://github.com/Its-Satyajit/reqly/commit/81be1b02fd24d70843a9fb3b5508e784cc68f113))
* **deps:** update module github.com/getkin/kin-openapi to v0.147.0 ([7bc3216](https://github.com/Its-Satyajit/reqly/commit/7bc3216bb6081c7a45ac4f313fd3878459d9a58b))
* **deps:** update module github.com/spf13/cobra to v1.10.2 ([8b8c9bc](https://github.com/Its-Satyajit/reqly/commit/8b8c9bc39792d51789df3d3202831fe5ede5ba29))
* **deps:** update module github.com/spf13/cobra to v1.10.2 ([8b8c9bc](https://github.com/Its-Satyajit/reqly/commit/8b8c9bc39792d51789df3d3202831fe5ede5ba29))
* **deps:** update module github.com/spf13/cobra to v1.10.2 ([0ea7201](https://github.com/Its-Satyajit/reqly/commit/0ea72016055bce0a8f5e25371adaca6fe9486bce))
* **deps:** update module github.com/wailsapp/wails/v3 to v3.0.0-beta.9 ([33ec967](https://github.com/Its-Satyajit/reqly/commit/33ec967bccf80afec36b88e53ddb97b17cdc5093))
* **deps:** update module github.com/wailsapp/wails/v3 to v3.0.0-beta.9 ([33ec967](https://github.com/Its-Satyajit/reqly/commit/33ec967bccf80afec36b88e53ddb97b17cdc5093))
* **deps:** update module github.com/wailsapp/wails/v3 to v3.0.0-beta.9 ([3f4340a](https://github.com/Its-Satyajit/reqly/commit/3f4340ad5230294f23a899b74440c4a6e673ccc6))
* **desktop:** review fixes — secrets panic, cross-collision, dirty-flag, empty-active warning (M15) ([26716ee](https://github.com/Its-Satyajit/reqly/commit/26716ee96640f0b8271b2dfd4094c9459e7cea12))
* **desktop:** review fixes — secrets panic, cross-collision, dirty-flag, empty-active warning (M15) ([a9bb28c](https://github.com/Its-Satyajit/reqly/commit/a9bb28c9ea4b92a70528b91a4a351e0733620bb7))
* **desktop:** wire environments view state into store (M15 T2 follow-up) ([076d97e](https://github.com/Its-Satyajit/reqly/commit/076d97e32c38dda403783cd27b418cb86127547f))
* **desktop:** wire environments view state into store (M15 T2 follow-up) ([27ea8c1](https://github.com/Its-Satyajit/reqly/commit/27ea8c10d804a47ef187b0a172e748a7457f3c72))

## [1.1.1](https://github.com/Its-Satyajit/reqly/compare/v1.1.0...v1.1.1) (2026-08-18)


### Bug Fixes

* **release:** pass tag_name to asset upload step ([4efc46a](https://github.com/Its-Satyajit/reqly/commit/4efc46a57720786519bcb56a4b9e1b4da80203df))
* **release:** pass tag_name to asset upload step ([f7c6526](https://github.com/Its-Satyajit/reqly/commit/f7c65260e3c8d9220e173b47f52832b596c5e4a8))

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
