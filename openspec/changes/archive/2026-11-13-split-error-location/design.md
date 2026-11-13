# Design

## Context

动机见 proposal.md. 现状: `UnmarshalRequest` 在出错时按 location 构造前缀 (`/body`, `/header/<Canonical-Name>`, `/query/<name>` ...) 并以 `validation.PrefixPointer` 前置到错误里已有的 pointer (来自 `Decoder.StackPointer()` 或多值参数的下标); `status` 在遍历 err 链时通过 `HasPointer` 取出 pointer 放进 `Detail.Pointer`.

位置信息分三层产生, 本期只改最外一层:

| 层                                                  | 产生的路径                                | 本期 |
|-----------------------------------------------------|-------------------------------------------|------|
| `validation/decode.go`, `union.go`, `tsfm_multipart.go` | body 内的 wire name 路径, 相对 body 根    | 不改 |
| `parameter.UnmarshalReaders` (多值参数)              | 参数内的下标, 如 `/1`                     | 不改 |
| `parameter.UnmarshalRequest`                         | location 段 + 参数名段, 前置到上面的路径  | 改   |

## Goals / Non-Goals

**Goals:**

- location 与路径分开表达, 使用方不需要解析字符串.
- `pointer` 是合法的 `jsontext.Pointer`, 相对 location 的根.

**Non-Goals:**

- 不引入 Go 数据路径 (见 Decisions 2).
- 不收集多个校验错误.
- 不改 validation 的遍历结构.
- 不改错误码的数值, 状态码注解, `title` 与 `detail` (change normalize-validation-codes).

## Decisions

### 1. `location` 与 `pointer` 的取值

| location | `pointer` 的根              | 示例                         |
|----------|-------------------------|------------------------------|
| `body`   | 整个请求体文档          | `/password`, `/items/1/age`  |
| `query`  | 该 location 的参数集合  | `/memberType`, `/ids/1`      |
| `header` | 同上, 名字取规范形式    | `/X-Create-Time`             |
| `cookie` | 同上                    | `/session`                   |
| `path`   | 同上                    | `/id`                        |

即 `pointer` = 原 `pointer` 去掉第一段 location. 错误只能定位到整个 location 时 (如请求体不是合法 JSON, 缺少整个 body) `pointer` 省略, 只有 `location`.

### 2. 路径继续使用 wire name

不改为 Go 数据路径. 理由: (a) `jsontext.Pointer` 只能生成 wire name 路径, 事后无法可靠映射回 Go 名 (数组下标与数字 key 无法区分, inline / embedded 字段无一一对应); 正向构建则要重写 `decode.go` 的遍历; (b) 前端拿到的路径可以直接对照自己发出的请求, 比暴露服务端类型更实用. plan D17 的 Go 路径方案及 T1 / T2 就此作废, 在 plan 中标注.

### 3. 位置在错误里的表示

`PointerError` 增加 location 字段. 在 `UnmarshalRequest` 出错处:

- 对 err 链调用新增的 `validation.WrapLocation(err, location)`: 链上已有位置错误时设置其 location (不覆盖已有值); 没有时包一层只有 location 的位置错误.
- `PrefixPointer` 只用于参数名段 (`/<name>`), 不再包含 location.

`internal/status` 新增 `HasLocation` 接口 (`Location() string`), 与 `HasPointer` 各自独立; `resolution` 分别记录 location 与 pointer, 第一个有值的节点为准. 要求沿用 "第一个" 规则, 与 pointer 现状一致.

### 4. 响应体与 `Detail`

`Detail` 新增 `Location string` (`json:"location,omitzero"`), `Pointer jsontext.Pointer` (`json:"pointer,omitzero"`) 保留名字但含义改为相对 location 的根. Go 字段, json 成员与 `HasPointer` / `PointerError` / `WrapPointer` / `PrefixPointer` 的命名保持一致 (最初曾命名为 `Pos` / `pos`, 提交前统一). 只出现在 `errors` 条目里, 与 `pointer` 现状一致; `Response` 顶层不含. 状态码兜底: `location` 非空时为 400.

### 5. 客户端回读

`UnmarshalResponse` 读 `location` / `pointer` 填入 `Detail`; 未知字段仍被忽略. `v1.0.0-rc1` / `rc2` 的服务端输出的 `pointer` 带 location 前缀, 读回时不做转换, 按新含义解释 (rc 版本之间的互通不在保证范围内).

### 6. 状态码兜底

`UnmarshalRequest` 是请求解析的唯一入口 (`transport/incoming.go`), 其字段循环内产生的**所有**错误都会带上 `location`, 包括读取请求体的 I/O 错误, 以及 (见 change normalize-validation-codes) 规则定义错误. 因此 "带 `location`" 只说明错误发生在请求解析过程中, 不能说明是客户端的问题.

状态码兜底规则保留 (无声明状态码时: 带 `location` 为 400, 否则 500), 用来覆盖没有 codex 标识层的错误 (json 语法错误, `SemanticError`, `strconv` 错误等, 它们确实是请求内容问题). 已声明状态码的错误以声明为准: 状态码解析是 "从外往里第一个声明的节点", 声明先于兜底生效 (规则定义错误如何利用这一点见 change normalize-validation-codes).

没有声明状态码、也不带 `location` 的错误 (`scanner.Structs.Scan` 失败即类型声明有误, 以及 operator 返回的未声明状态码的错误) 为 500: 这类错误是服务端的缺陷或未知情况, 4xx 会把责任推给客户端, 501 / 503 另有含义, 500 是唯一准确的描述. 这是沿用 D1 的默认值, 不是新规则.

## Risks / Trade-offs

- `pointer` 成员名不变而含义变化: 依赖旧含义的调用方不会报错, 但得到的路径少了第一段 -> 标 BREAKING, 迁移方式: 原 `pointer` = `"/" + location + pointer`. 受影响的只有已发布的 `v1.0.0-rc1` / `rc2`.
- header 的 `pointer` 取规范形式 (`textproto.CanonicalMIMEHeaderKey`), 与客户端发送的大小写可能不同 -> 与现状一致, 不变.
- `location` 在 `UnmarshalRequest` 外层设置, 若将来有别处产生位置错误, 需要同样设置 `location` -> 状态码兜底按 `location` 判定, 缺 `location` 会得到 500, 测试覆盖.
