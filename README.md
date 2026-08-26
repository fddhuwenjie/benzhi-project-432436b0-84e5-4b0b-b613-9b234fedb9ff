# mural-biocare

这是一个面向文物保护团队的壁画微生物污染处置准入服务。系统围绕单个处置个案，依次完成建档、污染评估、方案复核、小区试验、现场执行、成效复验和证据封存，并提供版本化 HTTP JSON API 与审计链校验。

标准构建：

```text
go build ./...
```

运行服务（默认监听 `127.0.0.1:19081`，可使用 `-addr=127.0.0.1:<port>` 或 `PORT` 配置；环境复测最小间隔默认 24 小时，可通过 `-baseline-retest-hours` 调整）：

```text
go run ./cmd/server -addr=127.0.0.1:19081
```

运行测试：

```text
go test ./...
```

阶段接口沿用 `/api/v1/cases/{case_id}` 聚合：`profile-correction` 仅在草稿期更正档案或交接负责人；`baseline`/`retest` 登记环境复测；`assessment` 校验采样时序并保存只读版本，`GET assessment-diff` 对比任意两个版本；`plan`/`review` 保存方案版本和四项审核清单；`pilot` 对完整逐日观察序列执行门禁；`start` 锁定检查点计划，`checkpoint` 原子核销计划项并支持偏差复发、整改任务关联，`resolve` 按升级后的严重度核验证据；`outcome` 将复验失败项转换为可核销整改任务；`GET archive-readiness` 返回分阶段封存缺口，`archive` 复检后生成六类确定性证据索引。`GET timeline` 与 `GET verify` 提供只读审计和摘要核验。阶段写请求使用 `X-Expected-Revision`、`X-Request-ID` 和 `X-Actor`。

启动自检会通过真实领域流程并主动退出：

```text
go run ./cmd/server -self-check -addr=127.0.0.1:19081
```
