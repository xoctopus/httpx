# Spec Delta

## ADDED Requirements

### Requirement: 校验错误的状态码与提示
由请求内容不符合规则引起的校验错误 (请求输入错误) SHALL 返回 400, 且 `title` 为 `非法请求`; 该规则对 `validation.Error`, `rule.Error`, `va.Error` 三个错误码类型统一适用. 由规则标签, 规则参数, validator 构造, 规则查找或解码器误用引起的错误 (规则定义错误, 包括全部 `rule.Error`) 即使流入响应, 也 MUST 返回 500 且不带 `title` 与 `detail`, `code` 与 `key` 照常输出. 哪个错误码属于哪一类 SHALL 由错误码的声明决定, 不由错误是否带有 `location` 决定. 三个类型的错误码数值 MUST 互不重叠, 且同一含义的错误码 MUST 只有一个 (正则不匹配只有 `va.ERROR__NOT_MATCH_REGEXP`).

#### Scenario: 输入超出规则范围
- **WHEN** 请求参数长度超出 `@string<[1,8]>` 的限制
- **THEN** 响应状态码为 400, `title` 为 `非法请求`, `key` 以 `VA::` 开头

#### Scenario: 缺少必填字段
- **WHEN** 必填字段缺失
- **THEN** 响应状态码为 400, `title` 为 `非法请求`, `key` 为 `VALIDATION::MISSING_REQUIRED`

#### Scenario: 规则标签写错
- **WHEN** 某个字段的规则标签无法解析, 错误流入响应 (如 `RULE::INVALID_RULE_LEADER`)
- **THEN** 响应状态码为 500, 响应体没有 `title`
- **AND** 响应体 `key` 为 `RULE::INVALID_RULE_LEADER`, 带有 `code`
- **AND** 响应体没有 `detail`, 不包含规则标签的原文与错误文本

#### Scenario: 其他规则定义错误
- **WHEN** 错误为 `va.ERROR__INVALID_LENGTH_RANGE`, `va.ERROR__INVALID_INT_VALUE`, `validation.ERROR__STRING_LENGTH_MODE`, `validation.ERROR__SLICE_PARAM`, `validation.ERROR__MAP_PARAM` 或 `validation.ERROR__UNREGISTERED_RULE`
- **THEN** 响应状态码为 500, 没有 `title` 与 `detail`, `key` 为对应的错误标识

#### Scenario: 定义错误不被位置兜底成 400
- **WHEN** 规则定义错误发生在请求参数解析过程中 (错误带有 `location`)
- **THEN** 响应状态码仍为 500

#### Scenario: 正则校验的提示
- **WHEN** 通过 `NewRegexpProvider(exp, name, hint)` 注册的规则校验失败, 且 `hint` 非空
- **THEN** 响应状态码为 400, `key` 为 `VA::NOT_MATCH_REGEXP`, `title` 为 `非法请求`, `detail` 为 `hint`

#### Scenario: 正则校验没有提示
- **WHEN** 正则规则校验失败, 且 `hint` 为空
- **THEN** 响应状态码为 400, `key` 为 `VA::NOT_MATCH_REGEXP`, `title` 为 `非法请求`, 没有 `detail`

#### Scenario: 正则校验不暴露 pattern 与输入
- **WHEN** 任意正则规则校验失败
- **THEN** 响应体中不出现正则的 pattern 与请求中的原始输入

#### Scenario: 错误码数值不重叠
- **WHEN** 比较三个类型所有常量的数值
- **THEN** 没有任何两个常量的数值相同
