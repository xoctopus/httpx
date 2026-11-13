# Design

## Context

本 change 是计划, 在 `split-error-location` 完成之后实施. 动机见 proposal.md.

现状 (摘自代码核对):

- `validation.Error` / `rule.Error` / `va.Error` 都从 0 开始编号, 响应里的 `code: 7` 无法区分来源.
- 这些错误码没有 `@code api` / `@code hint` 注解: 请求输入错误靠 "带 pointer 兜底为 400" 得到 400, 没有 `title`; 规则定义错误 (请求期构造 validator 时产生) 也被同一兜底变成 400.
- 标识层的 `detail` 取自 `codex.Errorf` 的格式化文本, 规则定义错误与正则校验的文本含规则原文, pattern 与输入.
- `validation.ERROR__NOT_MATCH_REGEXP` 与 `va.ERROR__NOT_MATCH_REGEXP` 含义重复.

## Goals / Non-Goals

**Goals:**

- 三个类型的错误码数值互不重叠, 且同一含义只有一个错误码.
- 请求输入错误 400 + `title: 非法请求`; 规则定义错误 500, 无 `title` 与 `detail`, 输出 `key`.
- 正则校验失败不暴露 pattern 与输入.

**Non-Goals:**

- 启动期规则校验 (plan P21, 下一期): 规则错误仍保留流入展示层的可能.
- 请求输入错误的 `detail` 收紧 (plan P22, 暂不处理).
- 为内置 regexp provider 补 hint 文案.

## Decisions

### 1. 错误码聚合: 状态码与提示

三个类型都是 `+genx:code`, 响应里的 `key` 分别为 `VALIDATION::` / `RULE::` / `VA::` 前缀. 按错误产生的位置 (调用点) 分成两类, 两类都显式声明状态码:

| 类别                 | 产生位置                                                       | 声明                                                         |
|----------------------|----------------------------------------------------------------|--------------------------------------------------------------|
| 请求输入错误         | validator 的 `Validate` 在运行时发现请求内容不符合规则         | `@code api=StatusBadRequest`, `@code hint=非法请求`          |
| 规则定义错误         | 规则解析, provider / validator 的构造, 规则查找, 解码器误用    | `@code api=StatusInternalServerError`, 不声明 hint           |

规则定义错误一律 500, 无 `title`, 仍输出 `code` 与 `key` (如 `RULE::INVALID_RULE_LEADER`), 便于开发者定位; 也**不输出 `detail`**: 标识层的 `detail` 取自 `codex.Errorf` / `Wrapf` 的格式化文本, 而这些错误的文本含规则标签原文与参数 (如 `got %s`), 属于内部信息. 因此规则定义错误统一以 `codex.Wrap(code, fmt.Errorf(...))` 构造, 文本进 cause (只进 `Desc` 供日志), 不进 `detail`. 涉及 `rule`, `va`, `validators` 与 `validation.go` 中规则定义错误的约 40 处调用点 (机械替换). 它们不应出现在展示层 (见 Decisions 3), 但保留流入的可能, 流入时按此处理.

归类依据逐个核对了调用点 (`文件:行` 为构造或校验发生处):

| 类型               | 错误码                      | 调用点                                                         | 类别         |
|--------------------|-----------------------------|----------------------------------------------------------------|--------------|
| `validation.Error` | `INPUT_TYPE`, `INPUT_VALUE` | `validators` 的 `Validate`                                     | 请求输入     |
|                    | `MISSING_REQUIRED`          | `arshaler.go:163` 缺必填参数, `wrap.go:64` 缺必填字段               | 请求输入     |
|                    | `STRING_LENGTH_MODE`        | `va_string.go` `New`: 规则参数不合法                           | 规则定义     |
|                    | `SLICE_PARAM`, `MAP_PARAM`  | `va_slice.go` / `va_map.go` `New`: 规则参数个数不合法          | 规则定义     |
|                    | `UNREGISTERED_RULE`         | `validation.go:137`: 规则名没有注册 (未知验证规则)             | 规则定义     |
|                    | `DEC_INVALID_INPUT`         | `decode.go:45`: `must.BeTrueWrap`, 调用方传入非指针, 本身是 panic | 规则定义 (不会成为响应) |
| `va.Error`         | `OUT_OF_LENGTH`             | `va_length.go` `Validate`                                      | 请求输入     |
|                    | `OUT_OF_VALUE_RANGE`        | `va_range.go` `Validate`                                       | 请求输入     |
|                    | `OUT_OF_ENUMERATED_VALUES`  | `va_enums.go` `Validate`                                       | 请求输入     |
|                    | `NOT_MATCH_MULTIPLE`        | `va_multiple.go` `Validate`                                    | 请求输入     |
|                    | `OUT_OF_FLOAT_SCALE`        | `va_float_scale.go` `Validate`                                 | 请求输入     |
|                    | `OUT_OF_INT_BITS`           | `va_int_scale.go` `Validate`                                   | 请求输入     |
|                    | `NOT_MATCH_REGEXP`          | `va_regexp.go` `Validate`, 以及 `NewRegexpProvider` 的校验函数 (原 `validation.ERROR__NOT_MATCH_REGEXP`, 见 Decisions 5) | 请求输入     |
|                    | `INVALID_LENGTH_RANGE`, `INVALID_VALUE_RANGE`, `INVALID_ENUM`, `INVALID_MULTIPLE`, `INVALID_FLOAT_SCALE`, `INVALID_INT_BITS` | 各 `New`: 规则参数不合法 | 规则定义 |
|                    | `INVALID_INT_VALUE`         | `va_int_scale.go:91` `Validate` 的 default 分支: 传给 validator 的值不是 `int64` / `uint64`, 是内部类型误用而不是请求内容 | 规则定义 |
| `rule.Error`       | 全部                        | `rule.Compile`: 规则标签解析错误                               | 规则定义     |

上一版把 `INVALID_INT_VALUE` 归为输入错误, 核对后更正; 并且 "不声明状态码" 改为 "显式声明 500", 理由见 Decisions 4.

### 2. offset

`code` 在 domain 内唯一, 三个类型原本都从 0 开始, 同一个 `code` 值对应三个不同的错误. 每个包新增一个不符合 `ERROR__<NAME>` 的 offset 常量 (genx 约定), `ERROR_UNDEFINED Error = iota + <Offset>`, 底层类型放宽为 `int32` 以容纳 offset:

| 包                    | offset 常量               | 取值    | 范围          |
|-----------------------|---------------------------|---------|---------------|
| `internal/validation` | `ValidationErrorOffset`   | 10000   | 10000 - 10008 |
| `internal/validation/rule` | `RuleErrorOffset`    | 11000   | 11000 - 11016 |
| `internal/validation/va`   | `VAErrorOffset`      | 12000   | 12000 - 12014 |

取值是我的选择, 项目内的 offset 登记表 (plan D6 称由项目集中管理) 目前不存在, 实现时在 plan 中补一张. `Value` 只用于内部排查, 对外标识以 `key` 为准 (plan 风险表).

### 3. 规则定义错误与展示层

规则定义错误通常在服务启动时已被 OpenAPI 构建的预编译拦截并 panic (`newmux` -> `DefaultBuildFunc` -> `rule.MustCompile` + `validation.New`), 但保证不完整 (如 `default` tag 请求期才读取, 预编译隐式依赖 OpenAPI 构建), 因此保留它们流入展示层的可能, 一律按 Decisions 1 处理. 启动期统一校验记入下一期 (plan P21).

### 4. 显式声明 500, 先于位置兜底

`UnmarshalRequest` 是请求解析的唯一入口, 其字段循环内产生的**所有**错误都会带 `location` (`split-error-location`), 包括规则定义错误. 位置兜底 (无声明状态码时带 `location` 为 400) 会把它们误判为客户端错误, 因此规则定义错误显式声明 `api=StatusInternalServerError`: 状态码解析是 "从外往里第一个声明的节点", 声明先于兜底生效.

### 5. 正则校验错误

- **合并重复的错误码**: `validation.ERROR__NOT_MATCH_REGEXP` 与 `va.ERROR__NOT_MATCH_REGEXP` 含义重复, 删除前者, 只保留 `va.ERROR__NOT_MATCH_REGEXP`. 调用点: `pkg/validation/validators/va_user_defined.go` (`NewRegexpProvider`) 改用 `va` 的错误码 (该包已依赖 `va`); `internal/validation/decode_test.go:185` 的断言同步改. 删除后 `validation.Error` 的常量顺序变化, 数值随 offset 整体重排, 已属 BREAKING 范围.
- **不暴露 pattern 与输入原文**: 响应只表明 "不匹配" (`key` 为 `VA::NOT_MATCH_REGEXP`, `title` 为 `非法请求`). `NewRegexpProvider(exp, name, hint)` 与 `va.RegexpVa` 保留的 `hint` 非空时作为 `detail`, 为空时省略 `detail`. pattern 与输入原文只放在 cause 里, 进入 `Desc` 供日志使用, 不进 `detail`.
- 构造方式: 以 `codex.Wrapf(va.ERROR__NOT_MATCH_REGEXP, cause, "%s", hint)` (hint 非空) 或 `codex.Wrap(va.ERROR__NOT_MATCH_REGEXP, cause)` (hint 为空) 构造, `cause` 为含 pattern 与输入的普通 error.
- `genx` 生成的内置 regexp provider 的 hint 目前都是空串, 是否补文案不在本期.

## Risks / Trade-offs

- 三个错误码类型的 `code` 数值整体变化, 底层类型变宽 -> 依赖数值的使用方会失效; 标 BREAKING, `key` 不变.
- 归类可能有误 -> 归类表在测试中固化: 每个输入错误断言 400 与 `非法请求`, 每个规则定义错误断言 500, 无 title, 有 key; 并断言 `StatusCode()` 的声明与归类一致.
