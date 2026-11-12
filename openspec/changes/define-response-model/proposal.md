# Proposal

## Why

httpx 的响应, 尤其是错误响应, 至今没有成文的契约: codex 错误一律返回 500, 隔一层 `fmt.Errorf("%w")` 就丢状态码, 响应中没有业务码, 400 隐藏信息而 500 反而泄露 cause (见 `obsidian/httpx/httpx.plan.md` P1-P20). 依赖已就绪 (x/codex v0.6.1 提供非泛型 `codex.Error`, `Detail()`, `Error.As`; genx v0.4.2 生成 `<DOMAIN>::<KEY>` 标识), 本期把 operator 各种返回方式对应的响应定下来, 并保证错误响应符合 RFC 9457 的语义.

## What Changes

- 明确 Output 支持的六种返回方式及其响应: `nil, nil`; `data, nil`; `httpx.WrapResponse(v, ...), nil`; `nil, codex 错误`; `nil, 普通错误`; `httpx.ErrorResponse(err, ...), nil`.
- 错误响应体按 RFC 9457 对齐: `title` / `detail` / `errors` 与扩展字段 `code` / `key`, 媒体类型 `application/problem+json`; 不输出 `type` / `status` / `instance`, 没有值的字段一律省略.
- **BREAKING** 错误响应字段语义: `code` 由 HTTP 状态码改为业务码; `key` / `title` 不再用 status text 兜底; `detail` 只放可公开的本次说明.
- 引入 "标识层": `code` / `key` / `title` / `detail` 取自 err 链从外往里第一个 codex 错误; 状态码独立沿链顺延, 隔层包装不再丢失.
- 内部信息 (Message, cause, `Error()`) 不出现在响应体; `Desc` 改为 `json:"-"`.
- `httpx.ErrorResponse(err, ...), nil` 在任意 operator (包括前置 operator) 中都写出错误响应并中断后续 operator.
- `ErrorResponse` 的响应选项按字段覆盖解析结果: `WithResponseStatusCode` / `WithResponseTitle` / `WithResponseDetail`, 为未使用 codex 的开发者留出入口.
- **BREAKING** 响应选项统一命名为 `WithResponse*` (`WithStatusCode` -> `WithResponseStatusCode` 等), 移除 `WithContentLength`.

## Capabilities

### New Capabilities

- `response-model`: operator 各种返回方式对应的 HTTP 响应, 包括成功响应的状态码, 响应头与 Content-Type, 以及错误响应的结构, 字段来源与内部信息不外泄的边界.

### Modified Capabilities

无 (`openspec/specs/` 目前为空).

## Impact

- 代码: `internal/status` (`ErrorsFrom` / `AsDetail` / `AsResponse` 遍历重写, `Desc`), `internal/response` (`WriteResponse` 中响应选项的优先级), `internal/route` (返回值为错误响应时中断), `pkg/httpx` (响应选项).
- 依赖: `github.com/xoctopus/x` v0.6.1, `github.com/xoctopus/genx` v0.4.2 (已升级, 重新生成 `*_genx_code.go`).
- API: 错误响应体字段语义变化, 响应选项改名为 `WithResponse*`, 新增 `WithResponseTitle` / `WithResponseDetail`, 移除 `WithContentLength`; 需在 CHANGELOG 标为 BREAKING.
- 验收: `example/cmd/example/main_test.go` (服务启动, confclient 请求, `Result.Into` 解码).
- 不在本期 (下一期): 错误位置 (`loc` / `pos`), `ErrorResponse` 作为 error 返回, `WithResponseCode`, 多个错误 (`errors.Join`), 下游错误树, 转发 `client.Result`, panic recover, 误用 panic (`STATUS__2xx.Wrap` 等), 错误响应日志, 重定向与自定义 Writer 的契约, 客户端解码 (`client-response`), OpenAPI 文档 (`response-document`).
