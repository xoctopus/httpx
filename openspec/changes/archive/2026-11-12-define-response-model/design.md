# Design

## Context

动机见 proposal.md. 设计决策的完整推导记录在 `obsidian/httpx/httpx.plan.md` (D1-D21, 待定 T1-T8) 与 `httpx.specs.errors_rfc9457_problem_detail.md`, 本文只写本期的实现方案.

现状 (工作区有未提交的改动):

- `internal/status.ErrorsFrom` 遇到第一个 `HasPointer` / `Describer` 即停, 中间的包装节点被丢弃; codex 错误不实现 `Describer`, 一律落到 500 (P1 / P2).
- `AsDetail` 把 cause 的 `Error()` 放进 `detail`, 用 status text 兜底 `key` / `title`.
- `Response` / `Detail` 已按 RFC 9457 改名, 错误响应已是 `application/problem+json`.
- `internal/response.WriteResponse` 用 `AsResponse` 的状态码覆盖 applier (P5).
- `internal/route` 的 handler 对非最后一个 operator 的返回值只做 ctx 注入, 返回值为 `ErrorResponse` 时链路不中断.
- 响应选项已改名为 `WithResponse*`; `WithResponseTitle` 未实现, `WithResponseDetail` 只存值, 写出时未生效.
- `status.Response` 是对外的暴露面 (非 error), `status.Detail` 是详情 (error); 两者不合并 (Response 为信息隐藏的扩展保留).

## Goals / Non-Goals

**Goals:**

- 错误解析只有一个出口, 运行时与将来的 OpenAPI 文档可以共用.
- 标识层与状态码顺延在一次遍历内完成, 逐节点判断, 不依赖 `errors.As` 的整链查找.
- example 的验收测试 (`example/cmd/example/main_test.go`) 全部通过.

**Non-Goals:**

- proposal "不在本期" 所列内容; 错误位置保持现有 `pointer` 输出, 本期不改.
- 不合并 `status.Response` 与 `status.Detail`; 不改客户端解码 (`UnmarshalResponse` 的回填待定).

## Decisions

### 1. 一次遍历, 两条规则

重写 `ErrorsFrom` + `AsDetail`: 沿 err 链从外往里访问每个节点 (包括 `fmt.wrapError` 等结构节点), 维护两条独立的状态:

| 状态   | 规则                                                                                                   |
|--------|--------------------------------------------------------------------------------------------------------|
| status | 第一个 `Describer.StatusCode() > 0` 的节点写入后不再改变 (D1)                                          |
| 标识层 | 第一个 `codex.Error` 节点写入 `Code` / `Key` / `Title` / `Detail` 后不再改变 (D8, D19)                  |

codex 节点通过 `Error.As(&target)` 把当前节点视为 code, 再断言 `status.Describer`, `codex.HasIdentifierCode`, `codex.HasHintCode`; 只看当前节点, 保证四个字段同源. `Code` 取 `Error.Code()`, `Detail` 取 `Error.Detail()`.

替代方案: 用 `errors.As` 分别查找各能力 - 会跨节点取值, 破坏 key 与 title 同源, 否决.

### 2. 删除 httpx 侧的 `HasIdentifier` / `HasHint`

改用 codex 的 `HasIdentifierCode` / `HasHintCode` (D11 修订).

### 3. 返回值为错误响应时中断

`internal/route` 的 handler: operator 的返回值是错误响应 (`ErrorResponse` 构造的值) 时, 不注入 ctx, 直接写出并中断, 前置 operator 与最后一个 operator 一致.

### 4. 响应选项按字段覆盖

选项把值记在 response 上; `WriteResponse` 在 `AsResponse` 的结果之上按字段覆盖: 状态码 (`WithResponseStatusCode`), `Title` (`WithResponseTitle`), `Detail` (`WithResponseDetail`). 修复 P5: 状态码不再被 `AsResponse` 覆盖.

### 5. Desc

`Detail.Desc` / `Response.Desc` 改为 `json:"-"`, 填入对应节点的 `Error()`, 保证内部信息不进响应体. 写日志不在本期.

## Risks / Trade-offs

- [错误响应字段语义变化, 依赖 `code` 为 HTTP 状态码或依赖 message 的客户端会受影响] → CHANGELOG 标为 BREAKING.
- [客户端 `UnmarshalResponse` 仍用 status text 回填空的 `key` / `title`, 与服务端不兜底不一致] → 本期不改, 验收测试对解码结果只断言响应体中存在的字段; 下一期在 client-response 中处理.
- [验收测试共享 example 的内存用户表, 依赖执行顺序] → 注册类用例放在最后.

## Migration Plan

1. 合入本 change; CHANGELOG 列出 BREAKING 项与迁移方式 (`code` 改读状态行与 `key`, 选项改名 `WithResponse*`).
2. 回滚: 本 change 不涉及数据, 回退版本即可.
