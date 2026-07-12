---
title: NetSentry 知识管理规范
status: active
updated: 2026-07-11
---

# NetSentry 知识管理规范

## 1. 定位与边界

NetSentry 知识库是源码、设计决策、测试证据和 Git 演进的本地可检索镜像，固定为 `/home/virtual-machine/Desktop/NetSentry-Knowledge`。代码与配置是行为权威，Vault 是解释和索引权威；发生冲突时，以当前代码、测试和 `git show` 为准并修正笔记。

Vault 保持本地、非 Git，不配置远端知识仓库、GitHub artifact、GitHub secret 或 self-hosted runner。项目仓库只保存本规范和确定性抽取/测试脚本；本机 `.git/hooks/post-push` 被 Git 忽略，作为 push 成功后显式调用的同步 helper。

## 2. Obsidian 与目录

打开 Vault：

```bash
obsidian /home/virtual-machine/Desktop/NetSentry-Knowledge
```

- `00-MOC`：总入口；新增稳定主题必须可达。
- `01-项目核心架构`：拓扑、数据流、协议契约和版本边界。
- `02-功能模块拆解`：capture、engine、配置、storage、API、CI/CD。
- `03-技术栈知识点`：协议、并发、匹配、IPC、SQLite、测试与交付。
- `04-开发迭代记录`：里程碑、全量提交索引和 push 增量。
- `05-问题与解决方案`：故障、调试、性能和限制。
- `06-测试与发布规范`：测试矩阵、证据和 release gate。
- `07-参考资料`：RFC、标准与同类 IDS。
- `_待整理`：自动草稿，复核后合并到稳定笔记。

## 3. 稳定笔记要求

每篇笔记包含 YAML：分类、标签、关联模块、最后更新时间、对应代码相对路径、版本标记。正文写清定义、项目实现、不变量、边界/风险、关联和代码引用，并至少链接一篇已有笔记。路径使用仓库相对路径；token、sudo 密码、SSH 私钥、Docker auth、私有 pcap 和操作者敏感绝对路径不得进入 Vault。

自动增量记录只证明“哪些代码发生变化”，不能替代架构、规则引擎或 MITRE 证据语义。每次开发单元必须把可复用结论合并进相关稳定笔记。

## 4. 本地 push-success 同步

Git 原生没有客户端 `post-push` 生命周期钩子，因此普通 `git push` 不会自动运行 `.git/hooks/post-push`。该文件保留历史名称，但语义是“确认远端 push 成功后显式调用的 helper”，避免在远端失败时提前写入错误知识。

安装或修复 helper 权限：

```bash
chmod +x /home/virtual-machine/Desktop/NetSentry/.git/hooks/post-push
```

`$netsentry-next` 在 push 前记录远端 old SHA，push 成功并核验 new SHA 后显式执行 helper。它按 `old..new` 重建全量提交索引，生成本次增量笔记和 `_待整理` 草稿，并维护 MOC 自动入口。同步应幂等，重复处理同一 range 不产生重复提交行。

手动恢复任意范围：

```bash
NETSENTRY_SYNC_RANGE="oldsha..newsha" \
  /home/virtual-machine/Desktop/NetSentry/.git/hooks/post-push sync
```

也可直接验证确定性抽取器：

```bash
python3 scripts/sync_knowledge.py \
  --repo . \
  --vault /home/virtual-machine/Desktop/NetSentry-Knowledge \
  --before oldsha \
  --after newsha
```

helper 故障不能回滚已经成功的 Git push，因此 `$netsentry-next` 必须检查 note、MOC 和 SHA；缺失时立即按确切 range 重放并记录原因。用户绕过 `$netsentry-next` 直接运行 `git push` 时，也必须手动执行上述命令。

## 5. 验证与维护

本地检查：

```bash
cd /home/virtual-machine/Desktop/NetSentry
make knowledge-check
git status --short
git log -1 --oneline
```

维护时检查：YAML 完整、MOC 可达、双向链接、commit SHA 存在、路径为 repo-relative、自动草稿已复核、稳定笔记包含实质性知识。发布记录必须区分实现、合成测试、本地证据和仍阻塞事项，不得把 synthetic pressure 或无语料 fuzz 宣传成真实流量证据。
