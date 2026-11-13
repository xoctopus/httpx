# 场景表

本表是 spec 的可执行版本: 每一行对应一个测试, 测试里用 `// spec: <需求> / <场景>` 标注. 表里没有的场景不做.

验收入口: `example/cmd/example/main_test.go`, 服务启动后由 confclient 发请求, 用 `Result.Into` 解码. 严格断言原始响应, 并断言解码后的 `*status.Detail`. 验收覆盖不到的由单测覆盖.

| #  | 入口                                                              | 出口与期望                                                                       | 覆盖方式                                   | spec 需求 / 场景                           |
|----|-------------------------------------------------------------------|----------------------------------------------------------------------------------|--------------------------------------------|--------------------------------------------|
| 1  | `PUT /login`, 缺 `password`                                       | 400; `errors[0].location=body`, `pointer=/password` (不带 `/body` 前缀)        | 验收                                       | 错误位置 / 缺少必填的 body 字段, pointer 不带 location 前缀 |
| 2  | `GET /user`, 缺必填的 `memberType`                                | 400; `location=query`, `pointer=/memberType`                                         | 验收                                       | 错误位置 / 缺少必填的 query 参数           |
| 3  | `GET /user?memberType=x&createdAt=yesterday`                      | 400; `location=query`, `pointer=/createdAt`                                          | 验收                                       | 错误位置 / 缺少必填的 query 参数 (类型错误) |
| 4  | `DELETE /user/abc`, `userID` 是 int64                             | 400; `location=path`, `pointer=/userID`                                              | 验收 (先验证能否触发, 否则换等价方式)      | 错误位置 / 路径参数                        |
| 5  | `PUT /login`, 请求体不是合法 JSON                                 | 400; `location=body`, **没有** `pointer`                                             | 验收 (现有用例补断言)                      | 错误位置 / 请求体不是合法 JSON             |
| 6  | `PUT /login`, 请求体 `{"username":1}` (类型错误且缺 `password`)   | 400; `pointer=/username`                                                             | 验收                                       | 校验的失败顺序 / 同一 body 内的多个错误    |
| 7  | 业务错误 (密码不匹配, 查询失败)                                   | 原始响应体没有 `location` 与 `pointer`; 状态码沿用现有的 401 / 500                   | 验收 (现有用例补断言)                      | 错误位置 / 非解析错误没有位置              |
| 8  | 客户端 `Result.Into` 解码 #1 - #4                                 | `d.Errors[0].Location` / `Pointer` 与原始响应一致                                    | 验收 (并入 #1 - #4)                        | 错误位置 / 下游错误位置被保留              |
| 9  | header 与 cookie 参数不合法                                       | `location=header`, `pointer=/X-Create-Time` (规范形式); `location=cookie`, `pointer=/session` | 单测 (example 没有校验 header / cookie 的接口) | 错误位置 / header 与 cookie            |
| 10 | body 嵌套 `/items/1/age`, 数组元素 `/ids/1`, urlencoded, multipart | `pointer` 为对应路径                                                                 | 单测                                       | 错误位置 / 嵌套的 body 字段                |
| 11 | query 多值 `ids=1&ids=x`                                          | `location=query`, `pointer=/ids/1`                                                   | 单测                                       | 错误位置 / 多值参数的某个元素              |
| 12 | header 与 body 同时不合法                                         | `location=header` (`header -> path -> query -> cookie -> body`)                  | 单测                                       | 校验的失败顺序 / 多个 location 同时有错    |
| 13 | 错误只带 `location`, 链上没有声明状态码                           | 状态码 400                                                                       | 单测                                       | 位置决定状态码兜底 / 只有 location 的错误  |
| 14 | 错误链上已声明 500, 同时带 `location`                             | 状态码 500, 不被兜底覆盖                                                         | 单测                                       | 位置决定状态码兜底 / 已声明状态码不被兜底覆盖 |
| 15 | 没声明状态码, 也不带 `location`                                   | 状态码 500                                                                       | 单测                                       | 位置决定状态码兜底 / 没有声明也没有位置    |
| 16 | 下游响应体带 `location` / `pointer` (含子条目)                        | `UnmarshalResponse` 后保留                                                       | 单测                                       | 错误位置 / 下游错误位置被保留              |

## Red 预期

- 因 "实现仍把 location 拼进 pointer" 而失败 (新行为): #1, #2, #3, #4, #6, #8, #9 - #12, #16, 以及 #5, #7 新增的 "没有 `pointer`" / "没有 `location`" 断言中依赖新字段的部分.
- 应当通过 (证明基线): #5, #7 的原有断言, #14, #15; #16 的 "子条目读回 `location` / `pointer`" (子条目由 `Detail` 的 json tag 解码, 骨架已具备); #12 以及 #6 的字段与顺序本身 (现有实现已按此顺序报告, 只缺 `location` / `pointer` 的拆分).

## 更新记录

- 2026-11-13 #4 已验证: `DELETE /user/abc` (admin token) 返回 400 且带 `errors` 条目, 触发 path 校验错误, 触发方式不变. 该用例在现有测试中已存在, 补了 `location` / `pos` 断言.
- 2026-11-13 验收测试已写入 `example/cmd/example/main_test.go`: #1 - #4, #6 新增或补断言, #5, #7 补断言 (#7 并入 `ExpectProblem`, 对所有非解析错误生效); `ExpectInvalid` 签名改为 `ExpectInvalid(t, location, pos)`.
- 2026-11-13 Red 已确认 (验收部分): 失败只来自 `body: no pointer`, `entry location`, `entry pos` 三类断言, 共 6 个用例; 没有预期之外的失败.
- 2026-11-13 单测已写入 (#9 - #16): `internal/payload/transformer/position_test.go` (原 `pointer_test.go` 改名并改为断言 `location` + `pos`; 新增 `BodyInvalidJSON`, `HeaderBeforeBody`, `QueryBeforeBody`, `BodyDocumentOrderBeforeMissing`), `internal/validation/validation_error_test.go` (`TestLocation`), `internal/status/location_test.go` (#13 - #16), `response_test.go` 与 `detail_test.go` (改为断言 `Location` / `Pos`).
- 2026-11-13 Red 已确认 (单测部分): 失败集合为 `TestPosition` 全部子用例, `TestLocation` 除 `Nil` 外, `TestLocationStatus` 的 #13, `TestLocationBody` 的位置与 `pointer` 断言, `TestUnmarshalLocation` 的 "pointer 不被识别", `TestResponse`, `ExampleAsDetail`. 通过的基线: #14, #15, #16 的读回, `TestLocation/Nil`. 抽查 `HeaderBeforeBody` 等的失败值, 当前输出 `/header/X-Create-Time` 等, 字段与顺序正确, 失败原因仅是 `location` 与 `pos` 未拆分; 因此 #12 与 #6 的顺序要求预计无需改动实现 (任务 2.5 只需核对).
- 2026-11-13 spec 核对 (任务 1.4): "错误位置", "校验的失败顺序", "位置决定状态码兜底" 的每个场景都能在场景表中找到对应测试, spec 无需裁剪.
- 2026-11-13 实现完成 (任务 2.1 - 2.6): 验收 #1 - #8 与单测 #9 - #16 全部变绿; `go vet ./...`, `go test ./...`, `golangci-lint run` (0 issues), `openspec validate --strict` 通过. #12 与 #6 的顺序要求无需改动实现.
- 2026-11-13 命名统一 (提交前): `Detail.Pos` 改为 `Detail.Pointer`, JSON 成员名由 `pos` 改为 `pointer`; 因此 `pointer` 的含义由 "带 location 前缀" 变为 "相对 location", 场景表中 "没有 pointer" 的断言改为断言 `pointer` 的值不带 `/<location>` 前缀 (#1 / #2 / #3 / #4). 先做 Go 标识符重命名 (行为不变), 再改测试确认 Red, 最后改 JSON tag. 此后表中的 `pos` 一律读作 `pointer`; 上面几条历史记录保留当时的写法.
