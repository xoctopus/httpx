# Tasks

流程 (tdd_sdd): 场景表 -> 验收测试 -> 由测试核对 spec -> Red -> 实现 (每个任务自带 red-green) -> 验收变绿 -> 归档.

场景编号 #1 - #16 见本目录的 `scenarios.md`; 验收覆盖 #1 - #8, 单测覆盖 #9 - #16. 测试必须能编译, 缺少的类型先以空实现补出 (任务 1.1), 不允许以编译失败代替 Red. 测试里用 `// spec: <需求> / <场景>` 标注可追溯.

## 1. 先写测试 (Red)

- [x] 1.1 **空实现骨架** (只让测试能编译, 行为保持不变; 骨架里新增的 `Pos` 字段后来统一为 `Pointer`, 见 2.7): `validation.WrapLocation(err, location)` 直接返回 `err`; `PointerError.Location()` 返回 `""`; `status.HasLocation` 接口; `Detail` 增加 `Location string` (`json:"location,omitzero"`) 与 `Pos jsontext.Pointer` (`json:"pointer,omitzero"`) 字段 (暂不被填充); `UnmarshalResponse` 暂不读取. 验证: `go build ./...` 与 `go vet ./...` 通过, 现有测试结果不变.
- [x] 1.2 验收测试 `example/cmd/example/main_test.go`: 新增 #1 缺 `password`, #2 缺 `memberType`, #3 `createdAt=yesterday`, #4 `DELETE /user/abc`, #6 `{"username":1}`; 为现有的 #5 与 #7 补断言; #1 - #4 同时断言解码后的 `*status.Detail` (#8). 先验证 #4 的请求确实触发 path 校验错误, 否则换等价方式并更新 `scenarios.md`.
- [x] 1.3 单测 (`internal/payload/transformer/pointer_test.go` 改写, `internal/status/*_test.go`, `internal/validation/validation_error_test.go`): #9 - #16, 全部能编译, 只因断言失败而红.
- [x] 1.4 核对 spec: 逐条检查 "错误位置", "校验的失败顺序", "位置决定状态码兜底" 的每个场景都对应 `scenarios.md` 中的测试; 对不上的场景从 spec 移到范围外, 并同步 `scenarios.md`.
- [x] 1.5 Red: 运行验收测试与单测. 已有行为的场景通过; 新行为的场景因 "实现仍是 `pointer`" 失败, 逐个抽查失败原因不是测试本身的问题 (按 `scenarios.md` 的 "Red 预期" 对照).

## 2. 实现 (每个任务 red-green-refactor)

- [x] 2.1 `internal/validation/validation_error.go`: 把 1.1 的空实现补全: 位置错误增加 `location` 与 `Location()`; `WrapLocation(err, location)` (已有 location 不覆盖; 链上无位置错误时新建; 覆盖 `errors.Join` 各分支). 验证: `go test ./internal/validation/`.
- [x] 2.2 `internal/status`: 补全 `HasLocation`, 填充 `Detail.Location` 与 `Detail.Pointer` (相对 location); `resolution` 分别记录; 状态码兜底 (无声明时带 `location` 为 400, 否则 500, 已声明优先). 验证: #13 - #15 通过.
- [x] 2.3 (随 2.2 完成: 子条目由 `Detail` 的 json tag 解码) `UnmarshalResponse` 读回 `location` / `pointer` (含子条目). 验证: #16 通过.
- [x] 2.4 `internal/payload/transformer/arshaler.go` `UnmarshalRequest`: 前缀去掉 location 段, 改为 `PrefixPointer(err, "/<参数名>")` (body 不加参数名段, header 取规范形式) 并 `WrapLocation`; `tsfm_multipart.go` 与 `UnmarshalReaders` 的前缀不变. 验证: #9 - #12 通过.
- [x] 2.5 (核对结果: 现有遍历已与 spec 一致, 无需改实现, 由 `TestPosition` 的 `HeaderBeforeBody` / `QueryBeforeBody` / `BodyDocumentOrderBeforeMissing` 固化) 失败顺序: 核对 `UnmarshalRequest` 的遍历与 spec 一致 (`header -> path -> query -> cookie -> body`, 同 location 内按声明顺序, body 内先文档顺序再缺失字段); 与 spec 不一致时修实现, 不改 spec. 验证: #6, #12 通过.
- [x] 2.6 `internal/status/doc.go`: 更新 `HasPointer` 与 `HasLocation` 的说明, `Detail` 字段含义, 状态码兜底.

- [x] 2.7 命名统一 (提交前追加): `Detail.Pos` -> `Detail.Pointer`, JSON `pos` -> `pointer`, 与 `HasPointer` / `PointerError` 一致. 先做 Go 标识符重命名 (行为不变), 再改 spec / 场景表 / 测试确认 Red (JSON 仍为 `pointer`), 最后改 JSON tag 变绿. 验证: 验收与单测全绿.

## 3. 验收与归档

- [x] 3.1 验收测试 #1 - #8 全部变绿; `go vet ./...`, `go test ./...`, `golangci-lint run`, `openspec validate split-error-location --strict` 通过.
- [x] 3.2 文档同步 (obsidian): `httpx.plan.md` 的 D17 标注 "Go 路径方案作废, 改为 `location` + wire name `pointer`", T1 / T2 关闭, P16 修复列更新, 下一期候选加入 change `normalize-validation-codes`; `httpx.specs.errors_rfc9457_problem_detail.md` "错误位置" 一节改为 `location` + wire name `pointer`.
- [ ] 3.3 提交 (一个提交): `feat(status)!: split error pointer into location and pointer`, 带 `BREAKING CHANGE` 脚注 (`pointer` 的含义变化: 原为带 location 前缀的 `/body/password`, 现为 `location: body` + `pointer: /password`, 成员名不变; 迁移 `pointer` = `"/" + location + pointer`); 提交后 `make changelog`; `openspec archive split-error-location`; 范围外清单整理为下一期候选.
