# Proposal

> 状态: 计划, 在 `split-error-location` 完成之后实施.

## Why

`validation.Error`, `validation.rule.Error`, `validation.va.Error` 三个错误码类型从 0 开始各自编号, 响应里的 `code: 7` 无法区分来源; 它们没有状态码与提示的声明, 请求输入错误靠位置兜底得到 400, 且没有 `title`; 规则定义错误 (请求期构造 validator 时产生) 会被同一兜底误判为 400; 正则校验失败会把 pattern 与输入放进 `detail`; `validation` 与 `va` 各有一个含义相同的 `ERROR__NOT_MATCH_REGEXP`.

目标: 让 validation 错误码的数值, 状态码与提示语义明确, 并且不暴露规则内部信息.

## What Changes

- 聚合三个错误码类型: 请求输入错误声明 400 与提示 `非法请求`, 规则定义错误 (含全部 `rule.Error`) 显式声明 500, 无 `title` 与 `detail`, 都输出 `key`; 归类以调用点为依据 (见 design).
- 三者的数值用不同的 offset 错开 (`validation` 10000, `rule` 11000, `va` 12000), 底层类型由 `int8` 放宽为 `int32`.
- 删除与 `va` 重复的 `validation.ERROR__NOT_MATCH_REGEXP`, 只保留 `va.ERROR__NOT_MATCH_REGEXP`; 正则校验失败只表明不匹配, 不暴露 pattern 与输入, `hint` 非空时作为 `detail`.
- 规则定义错误统一以 `codex.Wrap(code, fmt.Errorf(...))` 构造, 文本进 cause, 不进 `detail`.
- **BREAKING** 错误码数值与底层类型变化 (`key` 不变); 校验类 400 响应新增 `title: 非法请求`; 规则定义错误为 500 且无 `title` 与 `detail`; 删除 `validation.ERROR__NOT_MATCH_REGEXP`.

## Capabilities

### Modified Capabilities

- `response-model`: 新增 "校验错误的状态码与提示" 一条要求.

## Impact

- 代码: `internal/validation`, `internal/validation/rule`, `internal/validation/va` 的 `error.go` 与 `*_genx_code.go`; `rule/scanner.go`, `rule/rule_matcher.go`, `va/*`, `pkg/validation/validators/*`, `validation.go` 中规则定义错误的构造方式 (约 40 处); `va_regexp.go`, `va_user_defined.go`.
- 依赖: `split-error-location` (状态码兜底, 位置字段) 先完成.
- 范围外: 启动期规则校验 (plan P21); 校验错误 `detail` 收紧 (plan P22); 内置 regexp provider 的 hint 文案.
