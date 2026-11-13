# Proposal

## Why

请求参数校验失败时, 错误响应用一个 `pointer` 定位出错位置, 如 `/body/password`, `/query/memberType`. 第一段其实是 location (`payload.Locations`), 其余才是该 location 内的路径; 两层语义混在一个字符串里, 使用方要自己按 `/` 拆分才能知道错在哪一类参数.

实测 (example 服务): `PUT /login` 缺 `password` 返回 `"pointer": "/body/password"`; `GET /user?type=1` 缺 `memberType` 返回 `"pointer": "/query/memberType"`. 路径用的是请求中的字段名 (wire name), 这对前端更直接, 保持不变.

plan D17 曾在 "Go 数据路径" 与 JSON Pointer 之间摇摆 (T1 / T2 待定). 本期收敛: 路径继续用 `jsontext.Pointer` 的 wire name 路径, 只把 location 独立出来. T1 / T2 因此不再需要.

目标: 把校验错误的出错位置拆成 `location` + `pointer`, 使前端无需解析字符串就能知道错在哪类参数的哪个字段.

## What Changes

- 错误详情用两个字段定位: `location` 为请求参数所在的位置 (`header` / `path` / `query` / `cookie` / `body`), `pointer` 为 RFC 6901 JSON Pointer, 相对该 location 的根, 不再带 location 前缀.
- **BREAKING** `pointer` 的含义变化: 原为带 location 前缀的 `/body/password`, 现为 `location: body` + `pointer: /password`. 成员名不变 (`v1.0.0-rc1` / `rc2` 已发布过旧含义), 读取方不会报错, 但路径少了第一段. 迁移: 原 `pointer` = `"/" + location + pointer`.
- 状态码兜底由 "有 pointer 时为 400" 改为 "有 location 时为 400"; 已声明状态码的错误以声明为准.
- `status.UnmarshalResponse` 读回下游错误体的 `location` / `pointer`.
- 快速失败的顺序成文: 只报告处理顺序中的第一个错误.

## Capabilities

### New Capabilities

无.

### Modified Capabilities

- `response-model`: 新增 "错误位置", "校验的失败顺序", "位置决定状态码兜底" 三条要求.

## Impact

- 代码:
  - `internal/validation/validation_error.go`: 错误位置类型增加 location.
  - `internal/payload/transformer/arshaler.go`: `UnmarshalRequest` 不再把 location 拼进 pointer, 改为设置 location.
  - `internal/status`: `HasPointer` 之外增加 location 接口, `Detail` 新增 `Location`, `Detail.Pointer` 改为相对 location, `UnmarshalResponse` 读回两者, 状态码兜底, `doc.go`.
- 不改: `internal/validation/decode.go` 与 `pkg/validation/union.go` 对 `StackPointer` 的使用 (json 文档内的路径本来就是相对 body 根的); `tsfm_multipart.go` 的前缀. `pkg/spec/openapi` 的 `JSONPointer` 属于 OpenAPI 文档解析, 与本期无关.
- 测试: 验收测试 `example/cmd/example/main_test.go`; 单测 `position_test.go` (原 `pointer_test.go`), `validation_error_test.go`, `detail_test.go`, `response_test.go`.
- 不在本期 (范围外): 收集全部校验错误 (保持快速失败), Go 数据路径, 错误响应日志; 错误码 offset / 状态码 / `title` / 正则 / 规则定义错误 (change `normalize-validation-codes`, 计划); 启动期规则校验 (plan P21); 校验错误 `detail` 收紧 (plan P22).
