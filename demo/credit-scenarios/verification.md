# Credit Demo 提交前验证

日期：2026-11-19。Go 1.24.13，Linux/amd64。
原有组件基线为 `d3b2f81494085d5235b28333666a9f12d13b3667`；
本次仅新增 Demo 源码、启动器、文档和演示证据，不修改原有控制组件。

## 本次实际运行

| 检查 | 实际结果 |
|---|---|
| `go test ./... -count=1 -race`，在 `agent-festival/go` 执行 | 22 个包通过，退出码 0 |
| 启动器运行全部 12 场景 | 10 PASS / 1 GAP / 1 FAIL / 0 ERROR，退出码 1 |
| 启动器选择 `happy,retry,revoked,replay` | 4 PASS，退出码 0 |
| 指定已经存在的输出目录 | 退出码 2，逐文件摘要确认原报告与证据没有被改写 |
| 指定不存在的场景名 | 退出码 2 |
| 当前完整报告的相对链接 | 76 个均指向运行目录内的实际文件 |
| 选定四场景报告的相对链接 | 26 个有效 |
| 归档 11 场景报告的相对链接 | 67 个有效 |
| 原始场景文件与归档副本 | 逐文件字节一致；HTML SHA-256 与原 `verification.json` 一致 |
| Go 格式与 Bash 语法 | `gofmt -l` 无输出，`bash -n` 通过 |

归档结果是早期 11 场景的历史证据，详见
[sample-report/README.md](sample-report/README.md)，不是上述 12 场景的新运行输出。

## 两项未满足的要求

- `revision-gap`：场景要求 er-2 正式替代 er-1，当前原型没有接收此权威替代关系的
  入口，旧操作仍获派发许可并产生一笔模拟效果。作为已有系统缺口记录，关联
  [GitCode #29](https://gitcode.com/xingchao_ml/agent-hub-alpha/issues/29)。
- `grant-regroup`：权限的主体、工具、动作、效果上限不变，仅将合并条目拆成两条，
  流程准入接受但派发被误拒。关联
  [GitCode #23](https://gitcode.com/xingchao_ml/agent-hub-alpha/issues/23)。

新增场景保留原本应满足的断言，未将上述两项改成 PASS。组件测试全绿不代表这些
系统要求已实现。原有 T2 宿主问题也不由本次新 Demo 宿主的行为代替验收。

## 复跑

从仓库根目录开始，预先安装 Go 1.24+ 和 Bash，并允许首次解析现有 Go 模块依赖：

```bash
cd agent-festival/go
go test ./... -count=1 -race
cd ../..
bash demo/credit-scenarios/run.sh
```

当前完整场景退出码应为 1，报告仍会生成。不要为了使启动器返回 0 而忽略失败或放宽断言。
启动器在临时目录构建二进制并负责清理二进制，保留报告与证据；输出目录不得预先存在。

启动器使用 `-buildvcs=false`，避免 Go 1.24 工具链在 Git worktree 下自动探测
VCS 信息失败；报告仍显式读取 Git HEAD。在源码有未提交变更时，HEAD 只标识组件
检出基线，不能替代这些未提交内容的版本证明。

## 验证边界

使用真实参考组件、确定性 S1/S2 替身和文件银行替身；中断场景真实终止子进程。
未接入真实模型、真实银行、企业 IAM 或 openJiuwen。未进行浏览器视觉验收、断电
持久性或分布式故障实验。
