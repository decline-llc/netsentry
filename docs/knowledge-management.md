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

`$netsentry-next` 在 push 前记录远端 old SHA，push 成功并核验 new SHA 后显式执行 helper。该本地 hook 只转发给版本化的 `scripts/post_push_sync.py` / `scripts.sync_knowledge` API；测试直接调用该 API，绝不依赖 `.git/hooks`。同步应幂等，重复处理同一 range 不产生重复笔记或 MOC 条目。

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

## 7. 本地 Vault 恢复手册

### 7.1 索引损坏或丢失

全量提交索引（`04-开发迭代记录/全量提交索引.md`）由 push-success helper 每次重建。如果文件损坏或被误删：

```bash
NETSENTRY_SYNC_RANGE="$(git -C /home/virtual-machine/Desktop/NetSentry rev-list --max-parents=0 HEAD | tail -n1)..HEAD" \
  /home/virtual-machine/Desktop/NetSentry/.git/hooks/post-push sync
```

使用项目首个 commit 到当前 HEAD 的范围即可重建完整索引。

### 7.2 从空 Vault 重建

如果 Vault 目录丢失：

```bash
mkdir -p /home/virtual-machine/Desktop/NetSentry-Knowledge
NETSENTRY_FULL_SYNC=1 \
  NETSENTRY_VAULT=/home/virtual-machine/Desktop/NetSentry-Knowledge \
  /home/virtual-machine/Desktop/NetSentry/.git/hooks/post-push sync
```

`NETSENTRY_FULL_SYNC=1` 触发从 root commit 开始的完整同步。

### 7.3 失败 range 重试

如果同步因临时问题中断（Python 异常、磁盘满、权限错误），修复问题后用相同 range 重新执行即可。helper 是幂等的：重复同步同一 range 不会产生重复的提交索引行，迭代笔记和草稿会被覆盖为相同内容。

确认上一次成功的远端 SHA：

```bash
git -C /home/virtual-machine/Desktop/NetSentry log --oneline -3 origin/main
```

然后用失败时的 range 或上次成功后的新 range 重新执行。

### 7.4 手动核验 MOC 可达性

同步后检查 MOC 入口的 wiki 链接是否指向存在的笔记：

```bash
cd /home/virtual-machine/Desktop/NetSentry-Knowledge
grep -oP '\[\[.*?\]\]' 00-MOC/NetSentry知识总览.md | while read link; do
  note="${link#\[\[}"
  note="${note%\]\]}"
  ls "$note.md" >/dev/null 2>&1 || echo "BROKEN: $link"
done
```

### 7.5 知识内容安全检查

确保 Vault 不包含凭据、token、sudo 密码或私有 pcap 路径：

```bash
cd /home/virtual-machine/Desktop/NetSentry-Knowledge
rg -li "ghp_\|sk-\|password\|sudo\|PRIVATE KEY\|secret\|token" | grep -v "_待整理"
```

匹配到的文件应手动审核；`_待整理` 目录下的草稿可安全删除。

### 7.6 helper 自身验证

Push-success helper 的完整测试套件：

```bash
cd /home/virtual-machine/Desktop/NetSentry
python3 -m unittest scripts.test_post_push_sync
```

8+ 个独立 Git fixture 测试（10 个测试方法），覆盖精确范围、幂等、非祖先、空范围、MOC 可达、失败恢复、凭据泄漏和路径泄漏。
