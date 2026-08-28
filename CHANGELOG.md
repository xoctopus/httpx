
<a name="HEAD"></a>
## [HEAD](https://github.com/xoctopus/httpx/compare/v0.0.1...HEAD) (2026-07-23)

### Chore

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

