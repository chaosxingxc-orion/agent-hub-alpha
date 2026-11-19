# 早期 11 场景验证记录

此目录保留用户指定的 `/tmp/agent-hub-credit-verification.e3kTLd` 中的原始场景报告、
状态文件、验证摘要和日志。报告文件逐字保留，未将后续场景混入这次历史记录。

- 生成时间：2026-11-19 04:29:19 UTC。
- 原有控制组件基线：`d3b2f81494085d5235b28333666a9f12d13b3667`。
- 11 个场景：10 个 PASS、1 个 GAP、0 个 FAIL、0 个 ERROR。
- `revision-gap` 有两条未满足断言，整个场景运行退出码为 1，不能视作全绿验收。
- 原始 Go 回归：22 个包通过，含当时新增的 Demo 包，启用 `-race`。
- 模型和银行均为替身；没有 openJiuwen 集成、真实模型或真实银行测试。
- 当时 Demo 源码尚未提交；报告中的 `reference_commit` 标识原有组件基线，
  不表示该 SHA 已包含 Demo。当前提交包含可运行源码。

下载仓库后，用浏览器打开 [report.html](report.html)，保留整个目录以便访问每个
场景的证据链接。[report.json](report.json) 是结构化结果；
[verification.json](verification.json) 是当时的验证摘要；
[scenarios.log](scenarios.log) 和 [regression-go.log](regression-go.log) 是原始运行记录。
HTML 的 SHA-256 与 `verification.json` 一致；未进行浏览器视觉验收。

报告内所有客户、业务编号、银行回执均为 Demo 生成的虚构数据。未收录编译后的
`credit-demo` 二进制、压缩包、重复的启动器检查结果，或依赖本机绝对路径的启动脚本。

当前默认演示额外包含 `grant-regroup` 权限拆分场景，其期望/实际结果详见
[使用说明](../README.md)。如需用当前源码重跑本目录所覆盖的 11 个场景：

```bash
# 在仓库根目录执行
bash demo/credit-scenarios/run.sh \
  -scenario happy,retry,conflict,revoked,lost-reply,restart,evidence,uncovered,history-veto,revision-gap,replay
```

新运行会产生独立目录；时间和回执文件应以新报告为准，不要求与本目录逐字相同。
