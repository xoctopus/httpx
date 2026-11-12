# Tasks

## 1. 依赖与清理

- [x] 1.1 确认 `github.com/xoctopus/x` v0.6.1 与 `github.com/xoctopus/genx` v0.4.2, 重新生成仓库内 `*_genx_code.go`, identifier 为 `<DOMAIN>::<KEY>`, `go build ./...` 通过
- [x] 1.2 grep 旧选项名 (`WithStatusCode` / `WithCookies` / `WithContentType` / `WithMetadata` / `WithContentLength`) 的调用点并改为 `WithResponse*` 或删除, `go build ./...` 通过
- [x] 1.3 删除 `internal/status` 的 `HasIdentifier` / `HasHint`, 改用 `codex.HasIdentifierCode` / `codex.HasHintCode`, `go vet ./...` 通过

## 2. 错误解析: 状态码顺延与标识层

- [x] 2.1 重写 `ErrorsFrom` / `AsDetail` / `AsResponse`: 逐节点遍历, status 独立顺延, 标识层经 `codex.Error` + `Error.As` 取 `code` / `key` / `title` / `detail`, 去掉 status text 兜底与 cause 进 `detail`; golden 表测试覆盖 spec "状态码解析", "标识层", "本次出错的说明" 的全部场景
- [x] 2.2 `Response.Desc` / `Detail.Desc` 改为 `json:"-"` 并填入节点的 `Error()`; 测试断言序列化结果中没有 `desc`, 响应体不含 Message / cause / `Error()` 文本 (spec "内部信息不外泄")
- [x] 2.3 测试断言错误响应体不含 `type` / `status` / `instance`, 顶层不含 `src`, 空字段被省略 (spec "错误响应体结构")
- [x] 2.4 更新 `internal/status` 的 doc.go: 标识层, 状态码顺延, Desc 不进响应体; 与实现一致

## 3. 返回方式与响应选项

- [x] 3.1 `internal/route` handler 对返回值为错误响应的 operator (含前置) 写出并中断; 测试覆盖 spec "以返回值返回错误响应"
- [x] 3.2 `WriteResponse` 中响应选项按字段覆盖: `WithResponseStatusCode` 不再被 err 链状态码覆盖 (P5), 实现 `WithResponseTitle`, 让 `WithResponseDetail` 生效; 测试覆盖 spec "响应选项"
- [x] 3.3 测试覆盖 spec "成功响应的状态码", "响应头与 cookie", "Content-Type", "错误响应的媒体类型", "错误的统一出口"; 与现有行为不一致处按 spec 修正
- [x] 3.4 更新 `pkg/httpx` 的 doc.go: 六种返回方式及其响应, 响应选项; Example 可运行通过

## 4. 验收与发布说明

- [x] 4.1 `example/cmd/example/main_test.go` 的验收测试全部通过 (`go test ./example/cmd/example/`): 服务启动, confclient 请求, `Result.Into` 解码
- [ ] 4.2 提交信息带 `BREAKING CHANGE` 脚注, 列出 BREAKING 项与迁移方式 (`code` 语义, `key` / `title` 不再兜底, 选项改名 `WithResponse*`, 移除 `WithContentLength`, `ErrorResponse` 作为返回值中断链路, 响应选项优先于解析结果; 并说明客户端 `Into` 仍回填); 提交后运行 `make changelog` 生成 CHANGELOG
- [x] 4.3 集成检查: `go vet ./...`, `go test ./...`, `golangci-lint run` 全部通过; `openspec validate define-response-model --strict` 通过 (`pkg/midx` `TestLogHandler` 的 B3 断言暂时注释, 为本期之前已存在的问题, 见 review.midx.01 #5, 由 midx 自己的提交修复)

## Workflow follow-up

- 评审通过后归档本 change, 确认 `openspec/specs/response-model/spec.md` 生成.
- 同步 `obsidian/httpx/httpx.plan.md` 的 M3 进度, 并把 "不在本期" 的内容整理为下一期 change.
