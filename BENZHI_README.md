# BENZHI_README

## 项目说明
- 项目：benzhi-project-432436b0-84e5-4b0b-b613-9b234fedb9ff
- 项目用途：壁画微生物污染处置准入服务，覆盖个案建档、污染评估、方案复核、小区试验、现场执行、成效复验与证据封存，并提供审计链完整性校验。
- Go 工具链：`golang:1.22`
- 前端工具链：无

## 项目描述
- 项目名称：mural-biocare
- 项目介绍：一个面向文物保护团队的壁画微生物污染处置准入服务，围绕单个处置个案完成建档、污染评估、方案复核、小区试验、现场执行、成效复验与证据封存。项目按 standard 档规划不少于 2000 行有效生产 Go 代码和 20 个生产 Go 文件，聚焦这一条闭环流程，不扩展资产管理、预约或报表业务。
- 项目概述：一个面向文物保护团队的壁画微生物污染处置准入服务，围绕单个处置个案完成建档、污染评估、方案复核、小区试验、现场执行、成效复验与证据封存。项目按 standard 档规划不少于 2000 行有效生产 Go 代码和 20 个生产 Go 文件，聚焦这一条闭环流程，不扩展资产管理、预约或报表业务。
- 核心工作流：处置负责人创建壁画污染个案并登记环境基线，检测人员提交采样评估，保护专家编制处置方案并由独立审核人复核；方案通过后完成限定小区试验，现场团队按检查点执行并闭环偏差，检测人员复验成效，负责人最终封存证据包，使个案从 DRAFT 依次进入 ASSESSED、PLAN_APPROVED、PILOT_PASSED、IN_PROGRESS、OUTCOME_VERIFIED 和 ARCHIVED；方案驳回、小区试验失败及执行偏差均返回明确的可整改状态。
- 对外接口：仅提供版本化 HTTP JSON API，调用方通过 /api/v1/cases 创建个案、提交各阶段命令、查询当前状态与审计证据；服务支持 -addr=127.0.0.1:<port>，默认监听 127.0.0.1:19081，亦可在未传 -addr 时读取 PORT 并绑定 127.0.0.1:<PORT>，禁止默认绑定 0.0.0.0 或常见低位端口。

## 标准构建、运行和测试命令
进入容器后执行：
cd '/app' && GOTOOLCHAIN=local go build ./...

cd /app && GOTOOLCHAIN=local go run ./cmd/server -self-check -addr=127.0.0.1:19081

cd '/app' && GOTOOLCHAIN=local go test ./...

## Docker 构建和进入容器
chmod +x build_benzhi_docker.sh

./build_benzhi_docker.sh benzhi-project-432436b0-84e5-4b0b-b613-9b234fedb9ff-amd64 linux/amd64

./build_benzhi_docker.sh benzhi-project-432436b0-84e5-4b0b-b613-9b234fedb9ff-arm64 linux/arm64

docker run -it benzhi-project-432436b0-84e5-4b0b-b613-9b234fedb9ff-amd64:latest

## 题目验证命令
1. 预期退出码 0：`go test ./...`
2. 预期退出码 0：`go run ./cmd/server -self-check -addr=127.0.0.1:19081`
