# Spec Delta

## ADDED Requirements

### Requirement: 错误位置
请求解析与校验失败的错误条目 SHALL 用 `location` 与 `pointer` 定位出错的请求内容. `location` 为出错内容所在的 location: `header`, `path`, `query`, `cookie`, `body` 之一. `pointer` 为 RFC 6901 JSON Pointer, 相对该 location 的根, 不带 location 前缀: `body` 的根是整个请求体文档, 其余 location 的根是该 location 的参数集合, 第一段为参数名 (header 取规范形式), 多值参数后接下标. 只能定位到整个 location 时 `pointer` MUST 省略. 不是请求解析与校验产生的错误 MUST NOT 带 `location` 与 `pointer`.

#### Scenario: 缺少必填的 body 字段
- **WHEN** 请求体 `{"username":"1"}` 缺少必填的 `password`
- **THEN** 响应状态码为 400
- **AND** `errors` 条目的 `location` 为 `body`, `pointer` 为 `/password`

#### Scenario: 缺少必填的 query 参数
- **WHEN** 必填的 query 参数 `memberType` 缺失
- **THEN** `errors` 条目的 `location` 为 `query`, `pointer` 为 `/memberType`

#### Scenario: 嵌套的 body 字段
- **WHEN** 请求体 `{"items":[{"age":1},{"age":"x"}]}` 中第二个 `age` 类型错误
- **THEN** `location` 为 `body`, `pointer` 为 `/items/1/age`

#### Scenario: 多值参数的某个元素
- **WHEN** query 参数 `ids=1&ids=x` 中第二个值不合法
- **THEN** `location` 为 `query`, `pointer` 为 `/ids/1`

#### Scenario: header 与 cookie
- **WHEN** header `X-Create-Time` 的值不合法
- **THEN** `location` 为 `header`, `pointer` 为 `/X-Create-Time`
- **WHEN** cookie `session` 的值不合法
- **THEN** `location` 为 `cookie`, `pointer` 为 `/session`

#### Scenario: 请求体不是合法 JSON
- **WHEN** 请求体不是合法的 JSON 文档
- **THEN** 状态码为 400, `location` 为 `body`, 没有 `pointer`

#### Scenario: 非解析错误没有位置
- **WHEN** operator 返回 `nil, codex.New(user.ERROR__NOT_FOUND)`
- **THEN** 响应体没有 `location` 与 `pointer`

#### Scenario: pointer 不带 location 前缀
- **WHEN** 请求体 `{"username":"1"}` 缺少必填的 `password`
- **THEN** `pointer` 为 `/password`, 不是 `/body/password`

#### Scenario: 下游错误位置被保留
- **WHEN** 客户端解析一个下游错误响应, 其 `errors` 条目带有 `location` 与 `pointer`
- **THEN** 解析结果的对应条目保留相同的 `location` 与 `pointer`

### Requirement: 校验的失败顺序
请求解析与校验 SHALL 快速失败: 第一个失败即返回, 响应只报告这一个位置. "第一个" 按处理顺序确定: location 依次为 `header`, `path`, `query`, `cookie`, `body`; 同一 location 内按 operator 结构体的字段声明顺序; `body` 内先按文档中 key 出现的顺序, 再处理文档中缺失的字段.

#### Scenario: 多个 location 同时有错
- **WHEN** header 与 body 同时存在不合法的值
- **THEN** 响应的 `location` 为 `header`

#### Scenario: 同一 body 内的多个错误
- **WHEN** 请求体 `{"age":"x"}` 的 `age` 类型错误, 且必填的 `name` 缺失
- **THEN** `pointer` 为 `/age`

### Requirement: 位置决定状态码兜底
没有声明状态码的错误, 带有 `location` (发生在请求解析过程中) 时状态码 SHALL 为 400, 否则为 500 (服务端缺陷或未知情况). 已声明状态码的错误以声明为准, MUST NOT 被兜底覆盖.

#### Scenario: 只有 location 的错误
- **WHEN** 错误只携带 `location`, 链上没有声明状态码的节点
- **THEN** 响应状态码为 400

#### Scenario: 已声明状态码不被兜底覆盖
- **WHEN** 错误链上已有节点声明 500, 同时错误带有 `location`
- **THEN** 响应状态码为 500

#### Scenario: 没有声明也没有位置
- **WHEN** 错误没有声明状态码, 也不带 `location`
- **THEN** 响应状态码为 500
