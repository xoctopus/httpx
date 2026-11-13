# Tasks

> 计划. 实施前按 tdd_sdd 流程重新出场景表并确认, 本文件只是预案.

约定: 每个任务先写测试并确认失败, 再改实现. 代码生成: `go run ./internal/cmd/gen`.

- [ ] 1 合并正则错误码: 删除 `validation.ERROR__NOT_MATCH_REGEXP`, `NewRegexpProvider` 改用 `va.ERROR__NOT_MATCH_REGEXP`, 更新 `decode_test.go:185`.
- [ ] 2 规则定义错误不输出 `detail`: `rule/scanner.go`, `rule/rule_matcher.go`, `va` 各 `New` 及 `INVALID_INT_VALUE`, `validators` 的 `STRING_LENGTH_MODE` / `SLICE_PARAM` / `MAP_PARAM`, `validation.go` 的 `UNREGISTERED_RULE`, 改为 `codex.Wrap(code, fmt.Errorf(...))`; 测试: 每个规则定义错误 `Detail` 为空, `Desc` 含原文本.
- [ ] 3 正则校验: `NewRegexpProvider` 与 `va.RegexpVa.Validate` 按 design 第 5 节构造; 测试: `detail` 为 hint / 为空, 响应不含 pattern 与输入.
- [ ] 4 offset 与注解: 三个包增加 offset 常量, 底层类型改为 `int32`, 按 design 归类表补 `@code` 注解, 运行代码生成; 测试 `codes_test.go`: 归类表固化, 数值互不重叠, 在 offset 范围内.
- [ ] 5 `go build ./...`, `go vet ./...` 通过: 修正类型不匹配.
- [ ] 6 响应级测试: 三个类型各取一个输入错误与一个规则定义错误, 断言状态码, `title`, `key`, `code`, `detail`; 规则定义错误加 `WrapLocation` 后仍为 500.
- [ ] 7 example 验收: 缺 `password` / `memberType` 的响应 `title` 为 `非法请求`.
- [ ] 8 文档: plan 补 offset 登记表; 提交带 `BREAKING CHANGE` 脚注; `go vet`, `go test ./...`, `golangci-lint run`, `openspec validate --strict`; 归档.
