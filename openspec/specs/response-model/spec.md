# response-model Specification

## Purpose

定义 operator 各种返回方式对应的 HTTP 响应: 成功响应的状态码, 响应头与 Content-Type; 错误响应按 RFC 9457 输出的结构, 字段来源与内部信息不外泄的边界.

## Requirements

### Requirement: Output 的返回方式
operator 的 Output SHALL 支持以下返回方式, 每种方式的响应由本 spec 的对应要求定义: `nil, nil`; `data, nil`; `httpx.WrapResponse(v, ...), nil`; `nil, codex 错误`; `nil, 普通错误` (如 `fmt.Errorf`); `httpx.ErrorResponse(err, ...), nil`.

#### Scenario: 六种返回方式都能写出响应
- **WHEN** operator 分别以上述六种方式返回
- **THEN** 每次请求都得到完整的 HTTP 响应, 连接不被中断

### Requirement: 成功响应的状态码
operator 正常返回时, httpx SHALL 按以下顺序决定状态码: 返回值声明了状态码 (如 `WrapResponse` 设置 `WithResponseStatusCode`) 时使用它; 返回值为 nil 时为 204; POST 请求返回非 nil 值时为 201; 其余为 200.

#### Scenario: 返回 nil
- **WHEN** operator 返回 `nil, nil`
- **THEN** 响应状态码为 204, 响应体为空

#### Scenario: POST 返回数据
- **WHEN** POST 请求的 operator 返回 `data, nil`
- **THEN** 响应状态码为 201, 响应体为 data 的序列化结果

#### Scenario: 其他方法返回数据
- **WHEN** GET 或 PUT 请求的 operator 返回 `data, nil`
- **THEN** 响应状态码为 200, 响应体为 data 的序列化结果

#### Scenario: WrapResponse 声明状态码
- **WHEN** operator 返回 `httpx.WrapResponse[any](nil, WithResponseStatusCode(204)), nil`
- **THEN** 响应状态码为 204

### Requirement: 响应头与 cookie
httpx SHALL 把 `WithResponseMetadata` 设置的响应头与 `WithResponseCookies` 设置的 cookie 写入响应, `WrapResponse` 与 `ErrorResponse` 一致.

#### Scenario: WrapResponse 设置 cookie
- **WHEN** operator 返回 `httpx.WrapResponse(v, WithResponseCookies(c)), nil`
- **THEN** 响应带有 cookie `c`

#### Scenario: ErrorResponse 设置响应头
- **WHEN** operator 返回 `httpx.ErrorResponse(err, WithResponseMetadata("X-Reason", "a")), nil`
- **THEN** 错误响应带有响应头 `X-Reason: a`

### Requirement: Content-Type
成功响应的 Content-Type SHALL 由返回值的序列化方式决定. `WithResponseContentType` SHALL 只改写 `Content-Type` 响应头, 不改变序列化方式.

#### Scenario: 默认 JSON
- **WHEN** operator 返回一个结构体值
- **THEN** 响应头 `Content-Type` 为 `application/json`

#### Scenario: 覆盖 Content-Type
- **WHEN** 返回 `WrapResponse(v, WithResponseContentType("application/vnd.x+json"))`
- **THEN** 响应头 `Content-Type` 为 `application/vnd.x+json`, 响应体仍为 v 的 JSON

### Requirement: 错误的统一出口
请求参数解析与校验失败, 以及任意 operator (包括前置 operator) 返回的 error, httpx SHALL 以同一套规则写出错误响应, 并中断后续 operator. 参数解析与校验失败的状态码为 400.

#### Scenario: 前置 operator 返回错误
- **GIVEN** 接口的 operator 依次为 Auth, CurrentUserInfo, ListUsers
- **WHEN** CurrentUserInfo 返回 `nil, codex 错误` (声明 404)
- **THEN** 响应状态码为 404, ListUsers 不被执行

#### Scenario: 参数校验失败
- **WHEN** 必填的 query 参数缺失
- **THEN** 响应状态码为 400, operator 不被执行

### Requirement: 以返回值返回错误响应
operator 返回 `httpx.ErrorResponse(err, ...), nil` 时, httpx SHALL 写出该错误响应并中断后续 operator, 前置 operator 与最后一个 operator 一致; 该错误响应同样遵守本 spec 关于错误响应的全部要求.

#### Scenario: 前置 operator 以返回值返回错误响应
- **GIVEN** 接口的 operator 依次为 Auth, CurrentUserInfo, UserPermission, PromoteUserLevel
- **WHEN** UserPermission 返回 `httpx.ErrorResponse(fmt.Errorf("PermissionDenied"), WithResponseStatusCode(403)), nil`
- **THEN** 响应状态码为 403, PromoteUserLevel 不被执行

### Requirement: 错误响应的媒体类型
错误响应的 `Content-Type` SHALL 默认为 `application/problem+json`; `WithResponseContentType` 设置时以其为准, 响应体仍为 JSON.

#### Scenario: 默认媒体类型
- **WHEN** operator 返回任意错误
- **THEN** 响应头 `Content-Type` 为 `application/problem+json`

#### Scenario: 覆盖媒体类型
- **WHEN** 返回 `ErrorResponse(err, WithResponseContentType("application/json")), nil`
- **THEN** 响应头 `Content-Type` 为 `application/json`, 响应体结构不变

### Requirement: 错误响应体结构
错误响应体 SHALL 为 RFC 9457 风格的 JSON 对象, 顶层只使用 `title`, `detail`, `errors` 与扩展字段 `code`, `key`; 没有值的字段 MUST 省略. 响应体 MUST NOT 包含 `type`, `status`, `instance`: 状态码只体现在状态行, 错误标识由 `key` 承担, 出处由 `Server` 响应头提供.

#### Scenario: 普通错误的最小响应体
- **WHEN** operator 返回 `nil, fmt.Errorf("UnknownError")`
- **THEN** 响应状态码为 500
- **AND** 响应体顶层没有 `code`, `key`, `title`, `detail`, 也没有 `type`, `status`, `instance`

#### Scenario: 顶层不含来源与调试信息
- **WHEN** 任意错误响应被写出
- **THEN** 响应体顶层没有 `src` 与 `desc`

### Requirement: 状态码解析
错误响应的状态码 SHALL 取 err 链从外往里第一个声明了状态码 (> 0) 的节点. 未声明状态码的节点 (如 codex 错误码未标注 `@code api`) SHALL 被跳过, 继续往里找. 都没有时为 500; 参数解析与校验失败为 400.

#### Scenario: codex 错误码声明状态码
- **WHEN** operator 返回 `nil, codex.New(user.ERROR__PASSWORD_UNMATCHED)`, 该错误码声明 401
- **THEN** 响应状态码为 401

#### Scenario: 隔一层包装仍保留状态码
- **WHEN** operator 返回 `nil, fmt.Errorf("svc: %w", codex.New(user.ERROR__NOT_FOUND))`, 该错误码声明 404
- **THEN** 响应状态码为 404

#### Scenario: 显式声明状态码
- **WHEN** operator 返回 `nil, httpx.STATUS__INTERNAL_SERVER_ERROR.Wrap(fmt.Errorf("some error"))`
- **THEN** 响应状态码为 500

#### Scenario: 显式覆盖状态码
- **WHEN** operator 返回 `nil, httpx.STATUS__FORBIDDEN.Wrap(codex.New(user.ERROR__NOT_FOUND))`, 该错误码声明 404
- **THEN** 响应状态码为 403

#### Scenario: 外层未声明时顺延到内层
- **WHEN** operator 返回 `nil, codex.Wrap(user.ERROR__QUERY_FAILED, codex.New(sqlx.NOTFOUND))`, 外层错误码未标注 `@code api`, 内层声明 404
- **THEN** 响应状态码为 404

#### Scenario: 没有任何声明
- **WHEN** operator 返回 `nil, fmt.Errorf("UnknownError")`
- **THEN** 响应状态码为 500

### Requirement: 标识层
错误响应的 `code`, `key`, `title` SHALL 取自 err 链从外往里第一个 codex 错误 (标识层): `code` 为错误码数值, `key` 为错误码标识, `title` 为错误码提示. 标识层缺少的字段 MUST 省略, MUST NOT 向内层查找, 也 MUST NOT 用状态码文本兜底. 响应选项设置的字段优先 (见 "响应选项").

#### Scenario: 领域错误包装组件错误
- **WHEN** operator 返回 `nil, codex.Wrapf(user.ERROR__NOT_FOUND_BY_NAME, codex.New(sqlx.NOTFOUND), "用户名: %s 不存在", name)`
- **THEN** 响应体 `key` 为 `USER::NOT_FOUND_BY_NAME`, `code` 与 `title` 也来自该领域错误码
- **AND** 响应体不包含 sqlx 错误码的任何信息

#### Scenario: 标识层没有提示
- **WHEN** 标识层的错误码没有标注 `@code hint`
- **THEN** 响应体没有 `title`

#### Scenario: 没有标识层
- **WHEN** operator 返回 `nil, fmt.Errorf("UnknownError")` 或 `nil, httpx.STATUS__X.Wrap(fmt.Errorf(...))`
- **THEN** 响应体没有 `code`, `key`, `title`

### Requirement: 本次出错的说明
错误响应的 `detail` SHALL 取自标识层, 为构造该 codex 错误时 `codex.Errorf` / `codex.Wrapf` 的格式化文本; 使用 `codex.New` / `codex.Wrap` 构造, 或没有标识层时省略. `WithResponseDetail` 设置时以其为准.

#### Scenario: 带说明的领域错误
- **WHEN** operator 返回 `nil, codex.Wrapf(user.ERROR__NOT_FOUND_BY_ID, sqlErr, "用户ID: %s 不存在", "9")`
- **THEN** 响应体 `detail` 为 `用户ID: 9 不存在`

#### Scenario: 没有说明
- **WHEN** operator 返回 `nil, codex.New(user.ERROR__PASSWORD_UNMATCHED)`
- **THEN** 响应体没有 `detail`

### Requirement: 响应选项
`ErrorResponse` 的响应选项 SHALL 按字段覆盖 err 链的解析结果, 供未使用 codex 的开发者设置问题内容: `WithResponseStatusCode` 覆盖状态码, `WithResponseTitle` 覆盖 `title`, `WithResponseDetail` 覆盖 `detail`.

#### Scenario: 普通错误设置问题内容
- **WHEN** 返回 `ErrorResponse(fmt.Errorf("PermissionDenied"), WithResponseStatusCode(403), WithResponseTitle("无权操作"), WithResponseDetail("需要管理员权限操作")), nil`
- **THEN** 响应状态码为 403, `title` 为 `无权操作`, `detail` 为 `需要管理员权限操作`
- **AND** 响应体没有 `code` 与 `key`

#### Scenario: 覆盖 codex 声明的状态码
- **WHEN** 返回 `ErrorResponse(codex.New(user.ERROR__NOT_FOUND), WithResponseStatusCode(410)), nil`, 该错误码声明 404
- **THEN** 响应状态码为 410, `key` 仍来自该错误码

### Requirement: 内部信息不外泄
错误的完整文本 (codex 错误码的 Message, 被包装的 cause, 普通错误的 `Error()`) MUST NOT 出现在错误响应体的任何位置 (RFC 9457 §5).

#### Scenario: cause 不外泄
- **WHEN** operator 返回 `nil, codex.Wrapf(user.ERROR__NOT_FOUND_BY_NAME, codex.New(sqlx.NOTFOUND), "...")`
- **THEN** 响应体中不出现 sqlx 错误码的 Message

#### Scenario: 普通错误文本不外泄
- **WHEN** operator 返回 `nil, fmt.Errorf("UnknownError")`, 或 `ErrorResponse(fmt.Errorf("PermissionDenied"), ...), nil`
- **THEN** 响应体中不出现 `UnknownError` / `PermissionDenied`

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
