
<a name="HEAD"></a>
## [HEAD](https://github.com/xoctopus/httpx/compare/v0.0.2...HEAD) (2026-11-12)

### Chore

* format code
* ignore agent and editor generated files
* ignore generated openspec commands and skills
* regenerate code
* bump go to 1.27.3 and regenerate ci scripts
* **ci:** enable golangci modernize and apply fixes
* **deps:** bump x, genx, logx and regenerate code

### Docs

* **method:** document HTTP verb capability package
* **openspec:** add define-response-model change
* **operator:** document pipeline model and export capabilities

### Feat

* **midx:** adopt OTel semconv for access log and split metric handler

### Fix

* **routex:** build context injections per request

### Refactor

* adopt std encoding/json/v2 and drop internal/jsonv2 shim
* **httpx:** use metadata.Merge directly in confclient
* **midx:** rename middlex, add docs and BDD tests
* **path:** use value receiver for named segment

### BREAKING CHANGE


error response fields change meaning and options are renamed.
- code is the business code instead of the HTTP status code: read the status line for the status, and key for the error identity
- key and title are no longer filled with the status text, absent when the error has no codex identity
- response options are renamed WithResponse* (WithStatusCode -> WithResponseStatusCode, WithCookies -> WithResponseCookies, WithContentType -> WithResponseContentType, WithMetadata -> WithResponseMetadata), WithContentLength is removed
- an ErrorResponse returned as value by a pre operator now interrupts the chain instead of being injected into the context
- response options take precedence over the status resolved from the returned value or error
- note: confhttp Result.Into still fills empty key / title of decoded *status.Detail with the status text


<a name="v0.0.2"></a>
## [v0.0.2](https://github.com/xoctopus/httpx/compare/v0.0.1...v0.0.2) (2026-08-28)

### Chore

* bump Go to 1.27.0
* **deps:** upgrade dependencies
* **deps:** bump k8s.io/apimachinery ([#21](https://github.com/xoctopus/httpx/issues/21))
* **deps:** bump golang.org/x/net from 0.56.0 to 0.57.0 ([#17](https://github.com/xoctopus/httpx/issues/17))
* **deps:** bump github.com/xoctopus/logx from 0.3.2 to 0.3.5 ([#19](https://github.com/xoctopus/httpx/issues/19))
* **deps:** bump golang.org/x/sync from 0.21.0 to 0.22.0 ([#16](https://github.com/xoctopus/httpx/issues/16))
* **deps:** bump github.com/andybalholm/brotli from 1.2.1 to 1.2.2 ([#14](https://github.com/xoctopus/httpx/issues/14))
* **deps:** bump codecov/codecov-action from 5 to 7 ([#7](https://github.com/xoctopus/httpx/issues/7))
* **deps:** bump lint actions


<a name="v0.0.1"></a>
## v0.0.1 (2026-07-02)

### Chore

* pretty routes output
* move pathprefix to internal/payload/path
* export context hijack
* add LICENSE and gitignore
* keep generator directory
* **deps:** bump github.com/xoctopus/x from 0.4.4 to 0.4.5
* **deps:** bump codecov/codecov-action from 5 to 6

### Ci

* devgen regen
* add github workflow

### Feat

* confhttp.Server
* openapi spec add servers
* mux and example
* handler and spec
* validation, transformer etc...
* **middlex:** add middleware for Metrics handler
* **middlex:** add middlewares CORS, Canonical, Compress, PProf and Logging
* **statusx:** add statusx for richer http status

### Fix

* fix lint
* **statusx:** use sort.Slice instead of sort.Sort

