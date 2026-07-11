---
title: NetSentry 知识管理规范
status: active
updated: 2026-07-10
---

# NetSentry 知识管理规范

## 1. 定位与边界

NetSentry 知识库是项目源码、设计决策、测试证据和 Git 演进的可检索镜像，Vault 固定为 `/home/virtual-machine/Desktop/NetSentry-Knowledge`。知识类 Markdown、MOC、模块拆解、技术点、迭代记录、问题方案和参考资料一律写入 Vault；项目仓库只保留本规范和 `.git/hooks/post-push` 这两个联动/规范文件，不把 Vault 内容复制进源码目录。

代码与配置是行为权威，知识库是解释和索引权威；若二者冲突，以当前代码、测试和 `git show` 为准，并在下一次同步中修正笔记。

## 2. Obsidian 环境

Ubuntu 24.04 已通过 Flatpak 安装 Obsidian，终端别名 `obsidian` 可直接启动。打开知识库：

```bash
obsidian /home/virtual-machine/Desktop/NetSentry-Knowledge
```

Vault 使用标准 Markdown、YAML frontmatter、`[[笔记名]]` 链接和标签；目录名保持数字前缀以稳定排序。不要在 NetSentry 仓库创建知识类 Markdown。

## 3. 目录约定

- `00-MOC`：总入口和内容地图，新增主题必须从 MOC 可达。
- `01-项目核心架构`：拓扑、数据流、协议契约、版本边界和决策。
- `02-功能模块拆解`：capture、engine、配置、storage、API、CI/CD。
- `03-技术栈知识点`：libpcap/协议、atomic、Aho-Corasick、UDS、SQLite、Make、ASan、Actions/Docker。
- `04-开发迭代记录`：里程碑、全量提交索引、push 增量记录。
- `05-问题与解决方案`：故障、调试、性能和已知限制。
- `06-测试与发布规范`：测试矩阵、证据处理和 release gate。
- `07-参考资料`：RFC、标准和同类 IDS 对照。
- `_待整理`：钩子生成的增量草稿，只能在 Vault 内存在，整理后保留可追溯链接。

## 4. 笔记规范

每篇独立笔记必须包含 YAML：分类、标签、关联模块、最后更新时间、对应代码相对路径、版本标记；正文须写清定义/作用、项目实际实现、边界或风险；末尾必须有“关联”和“代码引用”。至少一个 Obsidian 双向链接指向已有笔记，不允许孤立文件。路径必须使用仓库相对路径，敏感 corpus 绝对路径只写 `redacted`。

链接规则：模块链接技术点，技术点链接模块和迭代记录，提交记录链接受影响模块/技术点；重命名笔记时同步全库引用。YAML 日期使用 ISO-8601 的 `YYYY-MM-DD`，版本标记使用当前 commit short hash 或发布版本。

## 5. Git push 增量机制

安装/修复钩子权限：

```bash
chmod +x /home/virtual-machine/NetSentry/.git/hooks/post-push
```

Git push 会把 ref 的 old/new SHA 传给钩子。仅对项目 SSH remote `git@github.com:decline-llc/netsentry.git` 执行同步。钩子会：

- 从当前 Git 对象库重建 `04-开发迭代记录/全量提交索引.md`，为每个提交记录 hash、日期、作者、提交说明、变更文件范围、测试/证据关联和知识链接；因此不会因人工遗漏造成历史索引缺口。
- 为本次 push range 生成 `04-开发迭代记录/*自动知识同步.md` 和 `_待整理/*自动同步草稿.md`，自动提取任务状态中的 fuzz 审批、ASan 迭代/输入数/脱敏路径、标准化 pcap/pcapng 样本清单和流量压测 deferred 状态。
- 扫描本次提交的技术文档/脚本/Makefile，维护 MOC 的自动同步入口区，使用稳定映射避免入口孤立；相关模块笔记的更新时间和版本标记同步刷新。

同步是幂等的：重复回放同一个 range 会覆盖同名迭代笔记、重建全量索引和替换 MOC 自动区，不追加重复提交行。失败只输出警告并以成功退出，避免知识同步故障阻断代码推送；日志带 `[netsentry-knowledge]` 前缀。

非 push 场景手动同步：

```bash
NETSENTRY_SYNC_RANGE="$(git -C /home/virtual-machine/NetSentry rev-parse HEAD~1)..$(git -C /home/virtual-machine/NetSentry rev-parse HEAD)" \
  /home/virtual-machine/NetSentry/.git/hooks/post-push sync
```

指定任意范围：

```bash
NETSENTRY_SYNC_RANGE="oldsha..newsha" /home/virtual-machine/NetSentry/.git/hooks/post-push sync
```

钩子不会自动执行 Obsidian CLI、网络请求或修改源码；它只读 Git/项目文档并写 Vault。写入的 Markdown 可由 Obsidian 自动发现，增量草稿仍由维护者在 `_待整理` 中完成语义复核。

## 6. 版本一致性

每次代码 push 后，增量记录必须包含 commit SHA；模块笔记的版本标记至少指向最近影响该模块的 short SHA。发布笔记必须区分“实现/测试通过”“本地证据”“仍阻塞”，不得把 synthetic pressure 或 no-corpus fuzz 宣传为真实语料/外部 fuzz 证据。release tag 必须来自通过 `make rc-check` 的 commit，并在知识库记录 tag 与发布工作流结果。

## 7. 安全与证据

不把 token、私有 pcap、外部 fuzz corpus、operator home 路径或本地生成 archive 写入公开知识库。使用 `make sanitize-pcap` 后仍须人工审查。知识库本身可能包含架构敏感信息，应按项目仓库访问控制管理。

## 8. 维护检查

日常检查：

```bash
cd /home/virtual-machine/NetSentry
git status --short
make docs-check
git log -1 --oneline
```

知识库检查重点是：每篇笔记有 YAML/关联/代码路径；MOC 可达；增量记录的 SHA 存在于 Git；项目源码目录未出现知识库文件；`_待整理` 草稿已定期合并到稳定笔记。
