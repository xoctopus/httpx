
<a name="HEAD"></a>
## [HEAD](https://github.com/xoctopus/httpx/compare/v0.0.1...HEAD) (2026-07-17)

### Chore

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

