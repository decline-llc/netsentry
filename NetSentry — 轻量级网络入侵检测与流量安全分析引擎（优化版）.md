#│  驱动选型: modernc.org/sqlite（纯 Go，无 CGO，兼容 CGO_ENABLED=0）  │
│  理由: 避免 mattn/go-sqlite3 的 CGO 依赖，Docker scratch 兼容      │
 NetSentry — 轻量级网络入侵检测与流量安全分析引擎（优化版）

> **发布版本**：v0.1.0（规划中）
> **定位**：生产级轻量 IDS 引擎，聚焦 pcap 离线分析与边缘网络检测场景
> **周期**：2026.06 — 2026.09（12 周，全部用于 v0.1.0 核心检测管道）
> **状态**：开发前设计文档 —— 代码尚未实现，本文描述目标设计与验收标准

---

## ⚠️ 项目定位与性能边界声明

**本项目的设计目标是成为小规模部署与 pcap 取证分析场景下的生产级轻量 IDS。**

NetSentry 在 v0.1.0 聚焦明确定义的边界内实现生产级质量——即边缘设备、pcap 离线分析、IDS 学习研究等场景。以下为硬性设计边界（知晓什么不做，比知晓什么能做更重要）：

| 限制项 | 说明 |
|--------|------|
| **性能上限** | ~50K PPS（单核，离线 pcap 模式），实时模式 ~30K PPS |
| **内存占用** | v0.1.0 约 30-50MB（空闲），满载约 80MB；v0.3.0 引入 Python pandas 后约 200-300MB |
| **"零拷贝"范围** | 仅指协议解析层（C 模块内部指针偏移）。跨 IPC 边界不适用——实际数据路径包含 5 次拷贝（libpcap→cJSON→UDS→Go堆→json.Unmarshal） |
| **检测覆盖** | 仅明文流量 payload 关键词 + IP 黑名单。无 TCP 流重组、无 IP 分片重组、无 TLS 解密 |
| **IPv6** | 不支持 |

NetSentry 的目标是：**在明确定义的边界内，每一个实现的功能都达到生产级质量**——包括 AAA 级别的测试覆盖（正常路径 + 异常路径 + 边界值 + 模糊测试）、完善的崩溃恢复、结构化日志、Prometheus 可观测性和安全的 API 设计。

如需万兆线速或全协议栈支持的 IDS，请使用 [Snort3](https://github.com/snort3/snort3) 或 [Suricata](https://github.com/OISF/suricata)。

---

### 50K PPS 性能可行性论证与微基准测试计划

v0.1.0 离线模式目标 ~50K PPS（单核），单包时间预算 20μs。以下为分段延迟估算（基于同类开源项目的经验数据，实际值以 W2 微基准测试为准）：

| 阶段 | 操作 | 预估单包延迟 | 说明 |
|------|------|-------------|------|
| C 端 | pcap 读取 + Eth/IP/TCP 解析 | ~3-5 μs | 协议解析层零拷贝，仅边界校验开销 |
| C 端 | cJSON 序列化 + PrintBuffered | ~5-8 μs | 12KB 栈缓冲，无堆分配；Payload Base64 编码额外 ~2 μs |
| IPC | UDS send/recv | ~2-4 μs | 本地 UDS，消息 < 8KB |
| Go 端 | json.Unmarshal + PacketInfo 分配 | ~4-6 μs | 使用 sync.Pool 复用结构体可降低到 ~2 μs |
| Go 端 | AC 自动机匹配（depth=4096） | ~2-5 μs | cloudwego/ahocorasick 双数组 Trie，匹配速度 ~200 MB/s |
| Go 端 | IP 黑名单 map 查找 | ~0.1 μs | Go map O(1) |
| **总计** | **关键路径** | **~16-28 μs** | 在 20 μs 预算的上限附近 |

**结论**：在最优情况下（缓存热、Pool 复用、小包）有可行性，但余量非常紧张。**v0.1.0 的 50K PPS 应视为 C 端抓包上限，实际检测吞吐取决于 Worker 处理速度**。以下措施保障性能目标：

**W2 微基准测试计划（含关键架构决策门）**：

W2 测试结果将决定以下核心架构选项——测试前不做预判，以数据为准：

| 决策门 | 触发条件 | 行动 |
|--------|---------|------|
| **单 Worker → 多 Worker** | 单协程 P99 延迟 > 25μs 且 actual_pps < 40K | 实施 SrcIP Hash 分流多 Worker（按 `fnv.New32a()` 取模），聚合器使用分片锁（`[]sync.Mutex`，256 片） |
| **JSON → Protobuf** | 端到端 P99 延迟 > 30μs 或 `json.Unmarshal` 占总延迟 > 40% | C 端引入 `nanopb`，Go 端引入 `google.golang.org/protobuf`，定义 `.proto` 文件。同时保留 JSON 路径（编译宏切换）用于调试 |
| **单 Worker 维持** | 单协程 P99 < 20μs 且 actual_pps >= 45K | 维持单协程，仅优化序列化路径（`sync.Pool` + pre-allocated buffers） |
| **JSON 维持** | 端到端 P99 < 25μs 且 JSON 占比 < 30% | 维持 JSON，仅优化 `sync.Pool` 和减少字段数量 |

**W2 微基准测试计划**：
```
测试工具: Go test -bench=. -benchtime=10s
测试场景:
  1. C 端独立基准: 100K 次 cJSON 序列化 + UDS send，测量 P50/P99
  2. Go 端独立基准: 100K 次 json.Unmarshal + AC 匹配，测量 P50/P99
  3. 端到端基准: C → UDS → Go → channel → Worker → /dev/null AlertWriter
  4. 压力测试: 持续 60s @ 100K PPS 输入，观察 channel_depth 和实际吞吐
验收标准: P99 延迟 < 30μs，持续 60s 不丢包（无 dropped），实际吞吐 >= 45K PPS
```

**Worker 单协程性能评估与监控（P1）**：

v0.1.0 采用单协程 Worker（`worker_count: 1`），理由是无竞态。但单协程意味着所有规则匹配串行执行。

**吞吐估算**：
- 单协程处理单包时间 ~8-12 μs（AC 匹配 + IP 查找 + 告警构造 + 聚合器入队）
- 理论最大吞吐 ~83K-125K PPS
- 考虑 channel 读写开销和 GC，实际预计 50-70K PPS
- 在实时模式下（30K PPS 目标）足够，离线模式下（50K PPS 目标）余量有限

**背压与降速透明化**：

channel 背压设计（阻塞发送传播到 C 端降速）保证了"不丢包"，但如果 Worker 处理速度跟不上输入速率，实际吞吐会静默下降到 Worker 的处理速度。为暴露降速幅度：

- 新增 Prometheus 指标 `netsentry_actual_throughput_pps`（Gauge，每秒实际处理包数，由 Worker 内部原子计数器 + 每秒采样计算）
- `netsentry_packets_received_total` 与 `netsentry_actual_throughput_pps` 的差值即为降速幅度
- `/api/health?verbose=true` 中暴露 `throughput.actual_pps` 和 `throughput.received_pps`
- 若 `actual_pps < received_pps * 0.9` 持续超过 60s → Prometheus AlertManager 告警
- 新增 `netsentry_worker_utilization_pct`（Gauge，Worker goroutine CPU 占用率估算，通过处理延迟/采样间隔计算）

**关于多 Worker 的务实立场**：7 专家审查建议 v0.1.0 即实现多 Worker（按 SrcIP Hash 分流）。但 v0.1.0 仍坚持单协程——理由：① 单协程天然无竞态，省去了多 Worker 共享聚合器 map 的锁开销（那会产生新的性能瓶颈）；② AC 自动机匹配的 CPU 开销远小于 JSON 反序列化和 channel 操作（3:7 比例），优先优化序列化路径 ROI 更高；③ 单协程的串行执行保证了告警的时序一致性（先到先匹配），这在取证分析中是有价值属性。

**v0.1.0 决策**：保持单协程，通过 `netsentry_actual_throughput_pps` 和 `netsentry_worker_utilization_pct` 监控瓶颈。若压测证实单协程不足，v0.2.0 实施 SrcIP Hash 分流多 Worker 方案。

**v0.2.0 管道化方案**（若 v0.1.0 压测证实单协程不足）：
- 将 Worker 拆为两个 stage：Stage 1（匹配 Worker）→ Stage 2（告警写入 Worker）
- Stage 1: 多协程并发匹配（RWMutex RLock 天然支持）
- Stage 2: 单协程聚合 + 写入（保持无竞态）
- 两者通过独立 channel 连接，Stage 1 阻塞写满后自然限速

**Prometheus 新增指标**：
- `netsentry_actual_throughput_pps`：Gauge，Worker 实际处理速度（PPS）
- `netsentry_worker_utilization_pct`：Gauge，Worker CPU 占用百分比估算

## 系统要求

| 依赖项 | 最低版本 | 用途 |
|--------|---------|------|
| **操作系统** | Linux 4.x+（UDS 支持）、WSL2（实验性）、macOS（实验性，UDS 兼容） | 运行环境 |
| **Go** | 1.21+ | 编译 engine 模块 |
| **gcc** | 9.0+（需 `-std=c11`） | 编译 capture 模块 |
| **libpcap-dev** | 1.9+ | 抓包库头文件与静态链接 |
| **make** | 4.0+ | 构建系统 |
| **Python 3 + Scapy** | 3.8+ | `make quickstart` 生成测试 pcap |
| **curl + jq** | 任意 | API 快速验证 |
| **磁盘空间** | ~200MB（含依赖与测试数据） | 开发环境 |

---

## 一、项目定位

### 一句话描述

基于 C/Go 双语言架构的生产级轻量网络入侵检测分析引擎，支持离线 pcap 流量分析、自定义规则匹配检测与 MITRE ATT&CK 映射告警。

**核心原则**："边界内生产级质量"。v0.1.0 核心交付：离线 pcap → AC 自动机 payload 检测 → MITRE 映射告警 → Prometheus 监控 → PSK API 认证 → 完整崩溃恢复路径。

### 技术堆栈与合理规模

| 语言 | 角色 | 合理性 |
|------|------|--------|
| **C** | libpcap 抓包 + 协议解析 | ✅ 正确：libpcap 是 C 原生库，C 模块可独立压测、独立部署为探针 |
| **Go** | 规则引擎 + HTTP API + 告警存储 | ✅ 正确：goroutine 适合并发管道，Go 模块可独立升级重启 |

Python 和 Vue3 前端推迟至 v0.3.0（如有实际需要）。

---

## 二、技术架构

### 2.1 整体架构图

```
┌──────────────────────────────────────────────────────────────────┐
│                      NetSentry v0.1.0 架构                         │
├────────────────────┬─────────────────────────────────────────────┤
│  Packet Capture    │  Rule Engine & Detection (Go)               │
│  (C)               │                                             │
│                    │  ┌──────────┐    ┌────────┐    ┌─────────┐  │
│  libpcap           │  │ Channel  │───▶│ Worker │───▶│ Alerts  │  │
│  pcap_open_offline │  │ (cap=    │    │ 单协程  │    │  频道   │  │
│  Eth→VLAN→IP→TCP   │  │  10000)  │    │        │    │         │  │
│  协议解析层零拷贝   │  │          │    │ RWMutex│    │ ┌─────┐ │  │
│  +边界校验(每字段)  │  │ 阻塞发送 │    │ 保护    │    │ │聚合器│ │  │
│  cJSON goto_cleanup│  │ (无丢弃) │    │ AC自动 │    │ │60s去 │ │  │
│  UDS 发送          │──│──────────│    │ 机+黑名 │    │ │重窗口│ │  │
│                    │  │          │    │ 单      │    │ │UPSERT│ │  │
│  心跳帧上报        │  │          │    │         │    │ └─────┘ │  │
│  (丢包/解析错误)    │  │          │    │ 更新    │    │   │     │  │
│  session_id 自证      │  │          │    │ last_   │    │   ▼     │  │
│                    │  │          │    │ packet  │    │ SQLite  │  │
│                    │  │          │    │ _at     │    │ +预写   │  │
│                    │  └──────────┘    └────────┘    │ 日志    │  │
│                    │                                │ +按天   │  │
│                    │                                │ 分库    │  │
│                    │                                └─────────┘  │
│                    │  ┌──────────────────────────────────────┐   │
│                    │  │  HTTP API (含统一错误格式 + 分页)     │   │
│                    │  │  /api/health?verbose=true            │   │
│                    │  │  /api/metrics (Prometheus, P0)       │   │
│                    │  │  /api/alerts (分页, ATT&CK筛选)      │   │
│                    │  │  /api/rules + /api/rules/batch       │   │
│                    │  │  /api/suppressions                   │   │
│                    │  └──────────────────────────────────────┘   │
├────────────────────┴─────────────────────────────────────────────┤
│              UDS + JSON 行式协议（含心跳帧 + session_id）             │
│              C → Go: PacketInfo JSON + Heartbeat JSON             │
│              C 端重连: 指数退避 1s→2s→4s→8s→max 30s              │
├──────────────────────────────────────────────────────────────────┤
│         SQLite WAL 模式（规则 + 告警，含 MITRE ATT&CK 字段）      │
│         告警按天分库: alerts_2026-06-25.db → 7天 TTL 自动清理      │
│         预写日志: alert_wal.jsonl (fsync + event_id 幂等重放)     │
└──────────────────────────────────────────────────────────────────┘
```

### 2.2 数据流全链路

```
pcap 离线文件
   │
   ▼
[1] C 模块：pcap_open_offline → Eth（含 VLAN tag 跳过）→ IP（含 fragment 检查）→ TCP 逐层解析
   │  ⚠️ 每字段解析前校验边界（IHL/total_length/data_offset 不越界）
   │  ⚠️ VLAN tag (0x8100/0x88A8) 循环跳过，定位正确 EtherType
   │  ⚠️ IPv4 fragment (MF=1 或 offset>0) → 跳过传输层解析，标记 is_fragment=true
   │  cJSON 序列化（goto cleanup 模式，零泄漏）→ UDS 逐行发送 JSON
   │  额外发送心跳帧（每 5 秒）：{"type":"heartbeat","session_id":"...","parse_errors":N,"sent":N}
   │  UDS 断开 → 指数退避重连（1s→2s→4s→8s→max 30s），重连期间包丢弃+计数
   │
   ▼ UDS + JSON 行式协议
[2] Go 模块：UDS 接收（单 goroutine 读取）
   │  ├─ PacketInfo JSON → channel 缓冲（阻塞发送，不丢弃）
   │  ├─ Heartbeat JSON → 直接更新 atomic.Value（不经过 channel）
   │  ▼
   │  Worker（单协程，持 RWMutex.RLock）：
   │  ├─ ip_blacklist (Go map O(1)) + payload_match (AC自动机, depth=4096可配)
   │  ├─ 匹配命中 → Alert (含 MITRE ATT&CK 字段)
   │  ├─ 更新 last_packet_at → atomic.Uint64
   │  └─ 流量统计 → atomic.Uint64
   │  ▼
   │  告警聚合器（单协程，内存 map + 定时清理过期窗口）：
   │  ├─ UPSERT: INSERT ... ON CONFLICT(rule_id,src_ip,dst_ip,dst_port,window) DO UPDATE
   │  ├─ 同窗口重复 → aggregated_count += 1, last_seen=now
   │  ├─ aggregated_count 达上限 → 强制 finalize
   │  └─ 定时器每 10s 清理过期窗口（last_seen < now - 60s）
   │  ▼
   │  critical 告警 → 独立 critical channel + 专用 writer goroutine（微批量 1s/10条一批）
   │  其他告警 → 批量缓冲 (100条/5秒) → SQLite + 截断预写日志
   │  ▼
   │  HTTP API：分页查询 / ATT&CK筛选 / Prometheus metrics
   │  /api/health?verbose=true 返回完整健康快照
   │  ▼
   │  规则热加载：POST /api/rules → RWMutex.Lock → 重建 AC + IP map → Unlock
   │  批量操作：PATCH /api/rules/batch, POST /api/alerts/batch-delete
   │
   ▼
[3] 退出序列（见第十二节完整 9 步序列）
```

---

## 三、IPC 通信与并发设计

### 3.1 C → Go 通信：UDS + JSON 行式协议 + 心跳帧

**消息类型**：

| 类型 | 格式 | 频率 |
|------|------|------|
| 数据帧 | `{"timestamp":..., "src_ip":"...", ...}` | 逐包 |
| 心跳帧 | `{"type":"heartbeat","session_id":"a1b2c3d4","seq":N,"sent":M,"dropped":D,"parse_errors":E,"buf_util_pct":U,"avg_json_serialize_us":Z}` | 每 5 秒 |
| 握手帧 | `{"type":"hello","version":"0.1.0","session_id":"...","pid":N,"hostname":"...","max_payload_len":4096}` | 连接建立时发送一次 |

心跳帧新增字段：
- `session_id`（C 端 connect() 成功后生成随机 UUID，进程级标识符），Go 端通过 `session_id` 变化区分"C 端重启"与"UDS 断开重连"——系统级 `boot_id`（取自 `/proc/sys/kernel/random/boot_id`）仅用于区分"整机重启"。
- `avg_json_serialize_us`（毫秒级采样平均），Go 端更新 `netsentry_json_serialize_duration_seconds` Histogram。
- `uds_write_errors`（自上次心跳以来的 UDS `send()` 错误计数，排除 EPIPE），Go 端更新 `netsentry_uds_write_errors_total` Counter。

**关键设计**：心跳帧**不经过数据 channel**。UDS reader goroutine 解析到 `type=="heartbeat"` 后直接更新 `atomic.Value` 存储的 `HeartbeatStatus` 结构体，Prometheus gauge 通过 `NewGaugeFunc` 从此读取。避免心跳帧与数据帧竞争 channel，保证 5s 间隔精确。

**v0.1.0 IPC 协议选择（务实立场）**：v0.1.0 保持逐包 JSON 行协议。理由：JSON 的可调试性（`strace`/`tcpdump` 可直接读 UDS 流量）在开发阶段价值巨大——C/Go 双语言 IPC 联调是项目最大风险点，JSON 可视化调试可在 W4 前发现 90% 的协议歧义。代价是 ~5-8μs/包的序列化开销，在离线 pcap 模式的 50K PPS 目标下处于预算边缘。

**W2 微基准测试的决策门**：若 JSON 路径的端到端 P99 > 30μs，v0.1.0 立即切换为 Protobuf（`google.golang.org/protobuf` + C 端 `nanopb`）。Protobuf 的优势：序列化速度 3-5× 快、消息体积小 30-50%、Go 端 `proto.Unmarshal` 的临时对象分配比 `json.Unmarshal` 少 ~40%。FlatBuffers（v0.2.0 路线图）进一步实现零拷贝读取。

**v0.2.0**：实时模式 + FlatBuffers/Protobuf 可选切换（通过编译宏 `-DUSE_FLATBUFFERS`），同时保留 JSON 路径用于调试模式。

**握手协议**：C 端连接成功后先发送握手帧（`{"type":"hello",...}`），Go 端校验版本兼容性后开始处理数据帧。详见 [附录 22.1](#221-uds-连接握手协议v010)。

### 3.2 C 端 UDS 重连状态机

```
状态: CONNECTED → (send() 返回 EPIPE/ECONNRESET) → DISCONNECTED
DISCONNECTED → 等待退避 → CONNECTING → connect() 成功 → CONNECTED
                                     → connect() 失败 → DISCONNECTED (重试)

退避策略: 1s → 2s → 4s → 8s → 16s → 30s (封顶)
最大重试: 无限制（C 端持续重连直到 Go 端恢复）
重连期间: 收到的包 → dropped++（不缓冲，避免 OOM）
```

### 3.3 Go 模块内部管道接口

```
// pipeline/interfaces.go — 管道接口定义

// Matcher — 规则匹配器接口
type Matcher interface {
    ID() string                              // 匹配器唯一标识，如 "ip_blacklist"
    Match(pkt *model.Packet) *model.Alert    // 返回 nil 表示未匹配
}

// AlertWriter — 告警输出接口
type AlertWriter interface {
    Write(alert *model.Alert) error
    Close() error
}
// v0.1.0 实现: SQLiteWriter
// v0.2.0 可扩展: KafkaWriter, WebhookWriter, StdoutWriter

// C 端 parser_registry（v0.2.0 完整实现，v0.1.0 仅保留接口定义和 passthrough 解析器）:
// capture/include/parser_registry.h

// 解析器返回值：决定是否继续由后续解析器处理（链式责任模式）
typedef enum {
    PARSE_PASS  =  0,    // 未识别该协议特征，交给下一个解析器
    PARSE_CLAIM =  1,    // 已识别并成功解析，停止传递
    PARSE_ERROR = -1     // 解析错误（数据异常），停止传递并记录错误
} ParseResult;

// 解析器函数签名：payload + 长度 → 解析结果 + 填充 PacketInfo
typedef ParseResult (*ParserFunc)(const uint8_t *payload, uint32_t len, PacketInfo *info);

// 解析器注册实体：通过 (protocol, port) 二元组匹配
typedef struct {
    uint16_t    port;        // 目标端口（0 = 匹配所有端口）
    uint8_t     protocol;    // IPPROTO_TCP / IPPROTO_UDP
    ParserFunc  func;        // 解析器函数指针
    const char *name;        // 解析器名称，用于日志和调试
} ParserEntry;

// 注册 API
void parser_registry_register(uint8_t protocol, uint16_t port, ParserFunc func, const char *name);

// 查找 API：按 protocol + port 精确查找，未找到返回 NULL
ParserFunc parser_registry_lookup(uint8_t protocol, uint16_t port);

// 链式调用示例（Go 端 UDS reader 调用）:
// ParseResult result = PARSE_PASS;
// for (每个注册的解析器) {
//     if (parser->protocol != pkt_protocol) continue;
//     if (parser->port != 0 && parser->port != pkt_dst_port) continue;
//     result = parser->func(payload, payload_len, &info);
//     if (result == PARSE_CLAIM) break;   // 被当前解析器"抢占"——停止链
//     if (result == PARSE_ERROR) break;   // 解析错误——停止链（记录错误）
//     // result == PARSE_PASS → 继续尝试下一个解析器
// }
// if (result == PARSE_PASS) {
//     // 所有解析器都未识别 → 标记为 generic 协议（仅保留 IP/端口信息）
// }

// v0.1.0 默认注册：
//   TCP/* + UDP/* → passthrough 解析器（仅记录 IP + 端口，不解析应用层协议）
//   所有应用层协议数据标记为 "generic"，payload 仍传输用于 AC 自动机匹配
// v0.2.0 默认注册（新增）：
//   TCP/80  → http_basic 解析器（仅提取 Content-Type 和 Host）
//   UDP/53  → dns_basic 解析器（仅提取 qname）
```

> **v0.1.0 范围说明**：parser_registry 的完整链式责任模式（`PARSE_PASS`/`PARSE_CLAIM`/`PARSE_ERROR`）属于 v0.2.0 设计，当前文档在此保留接口定义以指导后续开发。v0.1.0 仅注册一个 `passthrough` 解析器——不解析 HTTP/DNS 等应用层协议，仅记录 IP+端口，payload 传递到 Go 端由 AC 自动机做明文匹配。实际的应用层解析器（http_basic、dns_basic）将在 v0.2.0 实现。

> **v0.1.0 已知设计限制 — 规则类型硬编码**：当前 `RuleEngine.Match()` 的实现通过 `switch rule.Type` 硬编码了 `ip_blacklist` 和 `payload_match` 两条匹配路径。这不是理想的插件化设计，但 v0.1.0 仅需支持两种规则类型，接口抽象引入的复杂度（类型断言、错误传播、插件注册表）在此规模下得不偿失。v0.2.0 增加 `frequency_threshold` 和 `port_blacklist` 时，将重构为真正的 `Matcher` 注册表模式（`map[string]Matcher`），新增规则类型仅需实现接口并注册，无需修改核心引擎代码。

> **v0.1.0 已知设计限制 — 聚合器与 SQLite 耦合**：当前聚合器直接构造 SQLite UPSERT 语句。v0.2.0 将重构为"聚合器输出内存对象 → AlertWriter 接口决定落库方式"，实现聚合逻辑与存储后端的解耦。v0.1.0 保持此耦合是因为仅有一个存储后端（SQLite），解耦引入的 channel 传递和接口抽象在当前阶段属于过度设计。

**设计优势**：

- **链式责任（Chain of Responsibility）**：多个解析器逐级尝试——例如 TCP 80 端口先由 HTTP 解析器尝试，如果特征不匹配（非 HTTP 流量），返回 `PARSE_PASS` 交给通用 TCP 处理逻辑
- **抢占机制**：成功识别后 `PARSE_CLAIM` → 停止链式传递，避免资源浪费
- **错误隔离**：`PARSE_ERROR` 停止链但不会导致进程崩溃——仅增加 `parse_errors` 计数并通过心跳帧上报
- **可扩展性**：v0.2.0 支持 DNS/HTTP/MQTT/Modbus 等协议解析器时，无需修改核心调度逻辑——仅调用 `parser_registry_register()` 注册即可
- **对比原设计**：原 `void` 返回值的 `parser_fn` 无法区分"未识别"和"已处理"——新设计通过 `ParseResult` 三态返回值实现了严谨的解析器调度语义
```

### 3.4 规则热加载并发安全

**问题**：HTTP handler 和 Worker 是不同的 goroutine。`POST /api/rules` 重建 AC 自动机时，Worker 正在 `Match()` 读取 AC 自动机内部状态 → 竞态 panic。

**方案 A — v0.1.0**：`sync.RWMutex`（备选方案——高频热路径上写锁可能被读锁无限期阻塞，导致 HTTP API 超时）

```go
type RuleEngine struct {
    mu              sync.RWMutex
    acMatcher       *ahocorasick.Matcher
    ipBlacklist     map[string]*Rule          // key: IP 字符串 → *Rule
    ipCIDRTrie      *cidr.Trie                // CIDR 网段匹配（抑制规则和白名单共享）
    rulesByPriority []*Rule                   // 按 priority DESC 排序的规则列表
}

func (e *RuleEngine) Match(pkt *model.Packet) []*model.Alert {
    e.mu.RLock()
    defer e.mu.RUnlock()
    var alerts []*model.Alert

    // 按 priority 降序遍历规则列表
    for _, rule := range e.rulesByPriority {
        // early exit 条件：上一轮命中了 critical + early_exit 规则
        if e.shouldEarlyExit(alerts) {
            break
        }
        // ip_blacklist: O(1) map 查找
        // payload_match: AC 自动机匹配（仅在上游未 early exit 时才到达这里）
        if alert := rule.Match(pkt); alert != nil {
            alerts = append(alerts, alert)
        }
    }
    return alerts
}

func (e *RuleEngine) ReloadRules(rules []*Rule) error {
    newAC := buildAC(rules)
    newIP := buildIPMap(rules)
    sorted := sortByPriorityDesc(rules)  // 0–1000，高优先级在前
    e.mu.Lock()
    e.acMatcher = newAC
    e.ipBlacklist = newIP
    e.rulesByPriority = sorted
    e.mu.Unlock()
    return nil
}
```

**方案 B — v0.1.0（✅ 推荐）**：`atomic.Pointer` 实现 lock-free 规则切换。构建新 `ruleState` 快照后 `atomic.Pointer.Store()`。Worker 端 `atomic.Pointer.Load()` 获取快照指针，零锁、零阻塞。高频热路径上 RWMutex 的 RLock 仍有原子递减开销（~20ns），且在写锁等待期间所有读锁被阻塞——`atomic.Pointer` 完全消除此问题
```go
type RuleEngine struct {
    state atomic.Pointer[ruleState]  // 不可变快照
}
type ruleState struct {
    acMatcher   *ahocorasick.Matcher
    ipBlacklist map[string]*Rule
}
// ReloadRules: 构建新 ruleState → atomic.Pointer.Store()
// Match: state := atomic.Pointer.Load() → 读取快照 → 匹配
// 优势: 无锁、无阻塞、读路径零开销
```

### 3.5 Channel 背压策略

**原设计 risk**：select/default 丢弃——10K 缓冲在突发下 125ms 填满，之后每 125ms 丢弃 80000 包。

**设计**：**阻塞发送**（不使用 select/default）。当 channel 满时，UDS reader goroutine 阻塞 → UDS 接收缓冲区填满 → C 端 `send()` 阻塞 → C 端降速。形成端到端自然背压。

```go
// UDS reader goroutine
for scanner.Scan() {
    var pkt model.Packet
    if err := json.Unmarshal(scanner.Bytes(), &pkt); err != nil {
        continue  // 畸形 JSON 丢弃（不影响管道）
    }
    // 阻塞发送传播背压，但必须对 context 取消可中断——
    // 否则退出时若 channel 满，本 goroutine 会永久阻塞在发送上，造成泄漏。
    select {
    case packetCh <- &pkt:
    case <-ctx.Done():
        return  // 优雅退出：停止读取，让上层 drain 剩余 channel
    }
}
```

**设计原则**：IDS 的核心价值是"不遗漏"。丢包比慢处理更危险——丢掉的包可能恰好包含攻击特征。如果管道饱和，应当让上游减速而非静默丢弃。

**退出安全**：阻塞发送与优雅退出存在交互——退出序列先 `UDS listener.Close()`（停止新数据），再 drain channel。若此时 channel 满、Worker 卡在 SQLite 写入，reader 的 `packetCh <- &pkt` 会阻塞；故发送必须包在 `select{ case packetCh<-pkt: case <-ctx.Done(): }` 中，保证 context 取消时 reader 能退出，不留孤儿 goroutine（见第十二节退出序列与第十一节 goroutine 泄漏测试）。

**新增监控**：除 `netsentry_packets_dropped_total`（Counter，C 端丢弃）外，新增 `netsentry_channel_depth`（Gauge，当前队列深度），在 `/api/health?verbose=true` 中暴露，便于运维预警。

### 3.6 C 端协议解析边界检查（VLAN + Fragment）

**VLAN tag 处理**：

```c
// eth_parser.c — VLAN tag 跳过循环
uint16_t ether_type;
uint8_t *l3_start = pkt + ETH_HDR_LEN;
int l3_offset = ETH_HDR_LEN;

while (1) {
    if (pkt_len < l3_offset + 2) {
        parse_error_count++;
        return -1;  // 包太短，无法读取 EtherType
    }
    // 只读 EtherType 两字节，不复用 eth_header 强转（避免跨已校验边界）
    ether_type = ((uint16_t)pkt[l3_offset - 2] << 8) | pkt[l3_offset - 1];

    if (ether_type == 0x8100 || ether_type == 0x88A8 || ether_type == 0x9100) {
        // VLAN tagged (802.1Q / 802.1ad Q-in-Q)
        l3_offset += 4;
        if (pkt_len < l3_offset) {
            parse_error_count++;
            return -1;
        }
        continue;  // 继续跳过下一层 VLAN tag
    }
    break;  // 找到真正的 L3 EtherType
}
// l3_start = pkt + l3_offset
// 根据 ether_type 调用对应的 L3 解析器
```

**IPv4 fragment 处理**：

```c
// ip_parser.c — fragment 检查（在解析传输层之前）
uint16_t frag_off = ntohs(ip->frag_off);
bool is_fragment = (frag_off & 0x1FFF) != 0;         // 非首片
bool more_fragments = (frag_off & 0x2000) != 0;       // MF 标志

if (is_fragment || more_fragments) {
    pkt_info->is_fragment = true;
    if (is_fragment) {
        // 非首片：无传输层头，跳过 TCP/UDP 解析
        return 0;  // 仅记录 IP 层信息
    }
    // 首片 (MF=1, offset=0)：尝试解析传输层（但可能不完整）
}
```

**cJSON 内存管理模式**：

```c
// json_serializer.c — goto cleanup 模式，零泄漏
int serialize_packet(const PacketInfo *info, char *out_buf, size_t buf_size) {
    cJSON *root = NULL;
    char *json_str = NULL;
    int ret = -1;

    root = cJSON_CreateObject();
    if (!root) goto cleanup;

    // 逐字段添加（每个 AddXxxToObject 检查返回值）
    if (!cJSON_AddStringToObject(root, "src_ip", info->src_ip)) goto cleanup;
    if (!cJSON_AddNumberToObject(root, "src_port", info->src_port)) goto cleanup;
    // ... 其他字段 ...

    // cJSON_PrintBuffered 直接写入固定缓冲，避免额外的堆分配
    if (!cJSON_PrintBuffered(root, out_buf, buf_size, 0)) {
        goto cleanup;  // 缓冲不足
    }
    ret = 0;  // 成功

cleanup:
    if (root) cJSON_Delete(root);  // root 负责所有子节点内存
    // json_str 无需 free（PrintBuffered 使用外部缓冲）
    return ret;
}

// 栈缓冲增加到 12KB（覆盖 4096 字节 payload 全引号转义的极端情况）
// _Static_assert(sizeof(serialize_buf) >= 12288, "JSON buffer too small");
```

---

## 四、核心功能清单

### P0 — v0.1.0 必做

#### 功能 1：离线流量捕获与协议解析（C）

- `pcap_open_offline` 读取 pcap，**Ethernet（含 VLAN tag 跳过）→ IP（含 fragment 检查）→ TCP** 解析
- **每字段边界检查**：IHL、total_length、TCP data_offset、VLAN offset、fragment offset 解引用前校验
- **JSON 序列化**：cJSON `goto cleanup` 模式 + 固定栈缓冲 12KB（`cJSON_PrintBuffered` 直接写入，超出截断设 `truncated` 标志并保证 JSON 合法性——先截断 payload 再序列化）
- **心跳帧**：每 5 秒发送 `{"type":"heartbeat","session_id":"...","seq":N,...}`
- **UDS 重连**：指数退避 1s→2s→4s→8s→max 30s
- UDS 逐包发送

#### 功能 2：规则引擎（Go + RWMutex）

**v0.1.0 规则**：`ip_blacklist`（Go map O(1)）+ `payload_match`（AC 自动机，`cloudwego/ahocorasick`）

**并发安全**：`atomic.Pointer[ruleState]` 实现 lock-free 规则切换。Worker 通过 `Load()` 获取不可变快照后无锁匹配，HTTP handler 构建新快照后 `Store()` 原子替换。零竞态、零阻塞，消除了 RWMutex 在高频热路径上写锁被读锁阻塞的风险。

**payload 检测深度**：默认 4096 字节，`config.yaml` 可配置（`payload_preview_len`），规则级别可覆写 `offset`/`depth`。v0.2.0 支持不限深度。

**已知限制**：
> **致命检测盲区（请务必阅读）**：v0.1.0 的 payload_match 在每个 TCP segment 上独立运行，无 TCP 流重组。攻击者可将攻击特征拆分到两个 TCP segment 中（如 UNI + ON SELECT 分两包发送），或放在 POST body 4097 字节之后（超出默认检测深度）。此类攻击在本引擎中**完全不可检测**。部署前请确认：① 攻击面是否主要为单包攻击模式（如 ICMP 隧道、DNS 查询注入）；② 检测深度 4096 是否覆盖目标应用的典型请求体大小。完整 TCP 流重组在 v0.3.0 Roadmap。

**`payload_preview` 编解码流程（v0.1.0 明确化）**：C 端将原始 payload 字节经 Base64 编码后放入 JSON 的 `payload_preview` 字段发送。Go 端 UDS reader 接收后 `json.Unmarshal` 得到的 `payload_preview` 仍是 Base64 字符串。Worker 匹配前调用 `encoding/base64.StdEncoding.DecodeString()` 解码为明文字节串，在明文上运行 AC 自动机匹配。匹配成功后，告警中的 `payload_preview` 存储明文截断（前 200 字符），`raw_payload` 存储完整 JSON 上下文。此流程确保了 C→Go 传输的安全性（Base64 可安全嵌入 JSON）与 Go 端检测的便利性（明文匹配）。

- **Unicode/URL 编码绕过**：不检测 `%55%4E%49%4F%4E` 等编码变体，v0.3.0 Roadmap
- **注释插入绕过**：`UNION/**/SELECT` 不匹配，需正则或规范化预处理，v0.3.0 Roadmap

#### 功能 2.1：规则匹配逻辑

**多规则命中策略**：

v0.1.0 支持一个数据包同时触发多条规则——每个命中生成独立的 Alert 对象。例如：同一个 HTTP 请求包既可命中 `rule-001`（SQL 注入检测）又可命中 `rule-005`（XSS 检测），生成两条独立的告警——它们将被聚合器独立去重。

```
单包多规则匹配流程：
  for each enabled rule (sorted by priority DESC):
      match = rule.Match(packet)
      if match != nil:
          alert := buildAlert(rule, packet, match)
          alerts_channel <- alert
          if match.Severity == critical && rule.EarlyExit:
              break  // critical + early_exit: 跳过后续低优先级规则
```

**优先级处理**：

每条规则有 `priority` 字段（0–1000，默认 100）。规则匹配按优先级降序处理：

- **高优先级（priority ≥ 500）**：ip_blacklist 等"必须立即响应"的规则。命中 critical 级别的黑名单规则后，若 `early_exit: true`，**跳过后续所有 depth 检测**（payload_match AC 自动机匹配），直接生成告警。
- **中优先级（100–499）**：payload_match 等"需要深度检测"的规则。
- **低优先级（0–99）**：信息收集类规则（如端口探测统计、流量异常标记），用于补充上下文而非触发告警。

**早期退出优化（Early Exit）**：

```json
{
  "id": "rule-010",
  "name": "已知恶意 C2 IP 阻断",
  "type": "ip_blacklist",
  "severity": "critical",
  "priority": 1000,
  "early_exit": true,
  "config": {
    "ips": ["192.168.1.100", "10.0.0.0/8"]
  }
}
```

- `early_exit: true`（默认 false）：当此规则命中且 severity 为 critical 时，Worker 跳过后续所有规则匹配。这避免了已知恶意 IP 的包还进行 AC 自动机深度扫描的浪费。
- **注意**：early exit 仅在规则的 `severity == critical` 时生效——high 及以下不跳过后续规则（可能还有更高的检测价值）。
- **性能收益**：在 50K PPS 场景下，若 10% 流量来自已知黑名单 IP，early exit 可节省 ~15% CPU（避免了 5000 次/秒的 AC 自动机遍历）。

**快速匹配路径**：

```
Worker 匹配顺序（每个包）：
  1. ip_blacklist（Go map O(1) 查找 src_ip 和 dst_ip）
     ├─ critical → alert + early_exit（跳过 AC 自动机）
     └─ non-critical → alert + 继续
  2. port_blacklist（O(1) 查找 dst_port）
  3. payload_match（AC 自动机，在 payload_preview 上运行，CPU 密集）
     └─ 仅当前两步未触发 early_exit 时执行
```

#### 功能 3：告警系统（Go + SQLite + MITRE ATT&CK + UPSERT 聚合）

**告警模型（MITRE ATT&CK 字段，v0.1.0 必须）**：

```json
{
  "id": "alert-0001",
  "event_id": "evt_a1b2c3d4e5f6",
  "rule_id": "rule-001",
  "rule_name": "SQL注入特征检测",
  "timestamp": "2026-06-25T10:30:00Z",
  "src_ip": "10.0.0.99",
  "dst_ip": "192.168.1.10",
  "dst_port": 80,
  "protocol": "TCP",
  "severity": "high",
  "aggregated_count": 1,
  "first_seen": "2026-06-25T10:29:50Z",
  "last_seen": "2026-06-25T10:30:00Z",
  "mitre_tactic": "Initial Access",
  "mitre_technique_id": "T1190",
  "mitre_technique_name": "Exploit Public-Facing Application",
  "payload_preview": "GET /search?q=UNION+SELECT...",
  "matched_keyword": "UNION SELECT",
  "raw": { ... }
}
```

**告警聚合器（UPSERT 原子聚合）**：

```
聚合 Key: (rule_id, src_ip, dst_ip, dst_port, window_start)
时间窗口: 60 秒（可配置）
聚合上限: 100 条/窗口（可配置，达到后强制 finalize）

SQL（单条原子 UPSERT，消除 TOCTOU 窗口）:

INSERT INTO alerts (
    id, event_id, rule_id, rule_name, timestamp,
    src_ip, dst_ip, dst_port, protocol, severity,
    aggregated_count, first_seen, last_seen,
    mitre_tactic, mitre_technique_id, mitre_technique_name,
    payload_preview, matched_keyword, raw, created_at
) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, 1, ?, ?, ?, ?, ?, ?, ?, ?, datetime('now'))
ON CONFLICT(rule_id, src_ip, dst_ip, dst_port, window_start)
DO UPDATE SET
    aggregated_count = aggregated_count + 1,
    last_seen = excluded.last_seen,
    payload_preview = excluded.payload_preview;

-- 需要先建唯一索引：
CREATE UNIQUE INDEX idx_alert_aggregation
ON alerts(rule_id, src_ip, dst_ip, dst_port, window_start);

-- 窗口起始时间（截断到分钟边界）：
-- window_start = datetime(first_seen, 'start of minute')
-- 或按 60s 取整: window_start = (first_seen_unix / 60) * 60
```

**过期窗口清理**：
- 聚合器内存 map 每 10s 扫描，删除 `last_seen < now - 60s` 的条目
- 清理前触发 finalize：将聚合结果写入 SQLite，从 map 移除
- 使用 `time.Timer` 而非 ticker，避免清理与插入竞争

**分级写入**：
- critical 告警：**必须聚合**（10 秒短窗口），走独立 critical channel + 专用 writer goroutine（微批量 1s/10条一批）。若不聚合，内网单台机器被植入木马后每秒 1000 次 C2 连接将产生 6 万条/分钟的 critical 告警，直接击穿 SQLite。10 秒窗口可将告警量压缩 100-1000 倍且通过 `aggregated_count` 保留实际触发次数
- 其他 severity：聚合后经批量缓冲（100条/5秒）INSERT

**预写日志**：

```
格式: data/alert_wal_2026-06-25.jsonl
每条记录: {"event_id":"evt_xxx","alert":{...},"written_at":"..."}

写入顺序（关键！）:
  1. 构造告警 JSON（含 event_id UUID）
  2. fwrite(jsonl_fd, line) + fflush(jsonl_fd) + fdatasync(fileno(jsonl_fd))
  3. 执行 SQLite UPSERT

崩溃恢复:
  1. 扫描 jsonl 文件，逐行解析（截断行容错：读到换行符才算完整行）
  2. 对每条记录，以 event_id 查 SQLite：
     - 存在 → 跳过（已持久化，幂等）
     - 不存在 → UPSERT（崩溃时丢失的数据）
  3. 重放完毕后，轮转 jsonl（重命名为 .replayed）

轮转策略: 每日轮转，保留最近 3 天
```

**抑制规则/白名单**：

```json
{
  "id": "suppress-001",
  "rule_id": "rule-001",
  "src_ip": "10.0.0.0/24",
  "dst_ip": null,
  "reason": "Internal vulnerability scanner — Nessus weekly scan",
  "expires_at": "2026-06-26T00:00:00Z",
  "enabled": true
}
```

支持 CIDR 网段、到期自动恢复、带操作审计日志（创建/删除抑制规则时记录日志）。

**数据保留与清理策略**：

```
告警数据库:
  - 按天分库: data/alerts_2026-06-25.db
  - 每日 00:00 UTC 自动切换新库
  - 7 天 TTL: 超过 7 天的 .db 文件自动删除
  - 按天分库的优势: 无需 VACUUM，直接 rm 回收磁盘

SQLite 维护:
  - PRAGMA auto_vacuum = INCREMENTAL
  - 每日凌晨执行 PRAGMA incremental_vacuum（低峰期）
  - WAL checkpoint: 每次批量写入后检查 WAL 大小，超过 10MB 触发被动 checkpoint

规则/抑制规则: 存储在 netsentry.db（不轮转，数据量小）
```

**跨天分库查询方案（P0 级设计）**：

v0.1.0 采用简化方案 A：API 默认只查当天库。

```
GET /api/alerts?start_time=2026-06-24T00:00:00Z&end_time=2026-06-26T23:59:59Z
```

**行为规则**：
- 若 `start_time` 和 `end_time` 均在同一天（UTC）→ 仅打开 `alerts_YYYY-MM-DD.db` 查询
- 若时间范围跨天 → v0.1.0 返回 HTTP 400 错误，`error.code: "CROSS_DAY_QUERY_UNSUPPORTED"`，提示调用方分次查询
- API handler 在查询前通过 `start_time` 和 `end_time` 所在的 UTC 日期判断是否跨天
- v0.2.0 方案：支持跨天查询，使用 SQLite ATTACH 机制同时 ATTACH 最多 7 个历史库，UNION ALL 合并结果。分页通过应用层游标实现（先在各库分别 LIMIT+OFFSET，再应用层合并排序）。全局排序以 `first_seen DESC` 为准

**单库日期查询参数**（v0.1.0）：
- `date=2026-06-25`：显式指定查询某天的库，忽略 `start_time`/`end_time` 参数
- 未指定 `date` 时默认查当天（UTC），`start_time`/`end_time` 在当日内筛选

**告警写入归属规则**：按告警的 `first_seen` 时间戳决定写入哪天的库，而非处理时刻。确保包处理延迟不导致跨天错库。

**日库切换并发安全设计（P0 级设计）**：

每日 00:00 UTC 切换新库时的并发安全保障：

```
切换流程（由独立 goroutine 每 30s 检查触发）:

1. 检测到 UTC 日期变更 → 触发切换流程
2. 调用 aggregator.FinalizeAll() → 强制 finalize 所有内存中的聚合窗口
   ├─ 聚合窗口按 first_seen 归属（跨午夜窗口归属于 first_seen 所在日期）
   └─ 例如: 22:59 开始的窗口, 01:00 finalize → first_seen 在昨天 → 写入昨天的库
3. 创建新日期的 SQLite 数据库文件 + 执行建表语句
4. atomic.Pointer[sql.DB].Store(newDB) → 原子切换当前库指针
5. 旧库执行 WAL checkpoint → 关闭连接（不阻塞新库写入）
6. 清理超过 7 天 TTL 的旧 .db 文件
```

**并发安全关键设计**：

- **库指针**：使用 `atomic.Pointer[sql.DB]` 存储当前写入目标库。Worker 写入前 `Load()` 获取库指针，写入期间指针可能已切换到新库但不影响当前写入（写入操作作用于 Load 返回的旧指针）
- **写入时确定库**：聚合器 `Finalize()` 时根据 `alert.first_seen` 计算目标库日期，而非切换时的"当前库"。这意味着切换后仍可能有少量告警写入旧库（跨午夜延迟 finalize），这符合设计意图
- **切换原子性**：`atomic.Pointer.Store()` 的单个 CPU 指令保证 Go 端所有 goroutine 在同一时刻之后 Load 到新库指针
- **API 查询安全**：API handler 查询时显式指定 `date` 参数或通过当前 UTC 日期决定查询哪个库。查询打开只读连接，与写入连接独立
- **切换失败回退**：若新库创建失败（如磁盘满），保留旧库指针不变，记录 ERROR 日志，下一个检查周期（30s 后）重试
- **交叉写入窗口**：切换后 60s 内（聚合窗口最大值），旧库仍可能接收 finalize 写入。旧库执行 WAL checkpoint 后**立即关闭**（不等待延迟窗口）。当日 00:00:00 UTC 后 finalize 的告警按 `first_seen` 归属日期库——若 `first_seen` 在昨天则写入旧库（旧库已关闭则丢弃该告警，记录 WARN 日志），若在今天则写入新库。接受极少量的跨天边界数据丢失（~60s 聚合窗口内的尾量）以换取简洁的切换逻辑，避免 WaitGroup 追踪带来的死锁风险

**跨午夜数据包归属规则**：
- `first_seen` 时间戳决定归属日期
- 聚合窗口 `window_start` 始终截断到 60s 边界，归属于 `first_seen` 所在日期
- 例如: 23:59:50 的包 → first_seen = 23:59:50 → window_start = 23:59:00 → 归属当天库
- 同窗口内后续包即使跨午夜（00:00:10）→ first_seen 不变 → 仍归属当天库

**SQLite 磁盘满降级策略（P0）**：

磁盘满不是 IDS 引擎应该"优雅处理"的场景——它是运维失职，唯一合理的响应是**停止并告警**。

```
监控指标:
  netsentry_db_dir_free_bytes  Gauge  磁盘剩余空间（字节）——每 30s 由 Go runtime 更新

检查: 每次批量写入前 checkDiskFree()（syscall.Statfs），每 30s 一次
阈值: 剩余 < 50MB 或 SQLITE_FULL 错误

处理:
  1. 打印 ERROR 日志 "disk critically low (X MB free), stopping SQLite writes"
  2. Prometheus 指标 netsentry_db_emergency_mode 置 1
  3. /api/health 返回 status: "degraded"，component.sqlite: "disk_full"
  4. 停止所有 SQLite 写入（含预写日志）——后续告警直接丢弃
  5. 不做内存缓冲、不做自动恢复、不做环形队列

恢复:
  - 运维清理磁盘后**手动重启进程**
  - 不做自动恢复——磁盘满意味着 7 天 TTL 清理机制也已停止工作，人工介入是唯一正确路径
```

**Prometheus 指标**：
- `netsentry_db_dir_free_bytes`：Gauge，每 30s 更新
- `netsentry_db_emergency_mode`：Gauge，0/1

不实现内存环形缓冲的理由：v0.1.0 内存上限 50-80MB，缓冲 10000 条告警额外消耗 ~20MB。磁盘满时堆积告警毫无意义——运维必须清理磁盘后重启。代码量 ~30 行。
#### 功能 4：RESTful API + Prometheus（Go，v0.1.0）

**API 端点总览**：

| 路径 | 方法 | 功能 |
|------|------|------|
| `/api/health` | GET | 健康检查（轻量，给负载均衡器） |
| `/api/health?verbose=true` | GET | **完整健康快照**（含 channel 深度、SQLite 延迟、C 端心跳数据） |
| `/api/metrics` | GET | **Prometheus 格式指标** |
| `/api/alerts` | GET | 告警列表（分页、时间、severity、ATT&CK、IP、端口筛选） |
| `/api/alerts/:id` | GET | 单条告警详情 |
| `/api/alerts/batch-delete` | POST | **批量删除告警**（按 ID 数组） |
| `/api/stats` | GET | 流量统计快照 |
| `/api/rules` | GET/POST | 规则列表 / 新增（含校验） |
| `/api/rules/batch` | POST/PATCH | **批量导入 / 批量启用禁用** |
| `/api/rules/:id` | GET/PUT/PATCH/DELETE | 查看/全量更新/部分更新/删除规则 |
| `/api/suppressions` | GET/POST | 抑制规则列表 / 新增 |
| `/api/suppressions/:id` | GET/PUT/PATCH/DELETE | 抑制规则 CRUD |
| `/debug/pprof/*any` | GET | **Go pprof 性能剖析端点**（仅监听 127.0.0.1，需 PSK 认证） |
| `/debug/pprof/profile` | GET | CPU profile（?seconds=30 抓取 30s） |
| `/debug/pprof/heap` | GET | 堆内存 profile |
| `/debug/pprof/goroutine` | GET | Goroutine 堆栈 dump |

**分页设计（所有列表端点统一）**：

```
请求参数:
  page      int  页码（1-based，默认 1）
  per_page  int  每页条数（默认 20，上限 100）

响应 envelope:
{
  "data": [...],
  "pagination": {
    "page": 1,
    "per_page": 20,
    "total": 234
  }
}
```

**统一错误响应格式**：

```json
{
  "error": {
    "code": "VALIDATION_ERROR",
    "message": "Request validation failed",
    "details": [
      {"field": "rule_id", "reason": "must match pattern ^rule-[0-9]{3}$"},
      {"field": "severity", "reason": "must be one of: low, medium, high, critical"}
    ],
    "request_id": "req_a1b2c3d4"
  }
}
```

**HTTP 状态码规范**：

| 状态码 | 场景 |
|--------|------|
| 200 | 查询成功 |
| 201 | 创建成功（POST /api/rules） |
| 204 | 删除成功（DELETE），无响应体 |
| 400 | 请求参数校验失败 |
| 404 | 资源不存在 |
| 409 | 资源冲突（如 rule_id 重复） |
| 429 | 触发 rate limit |
| 500 | 内部错误 |

**告警筛选参数**：

```
GET /api/alerts?start_time=2026-06-25T00:00:00Z&end_time=2026-06-25T23:59:59Z
    &severity=high,critical
    &src_ip=10.0.0.0/24&dst_ip=192.168.1.0/24
    &dst_port=80,443
    &rule_id=rule-001
    &mitre_tactic=Discovery
    &mitre_technique_id=T1046
    &aggregated_count_min=10
    &sort_by=timestamp&sort_order=desc
    &page=1&per_page=20
```

所有筛选参数为 **AND 组合**（多值逗号分隔为 OR）。`sort_by` 支持 `timestamp`/`severity`/`aggregated_count`。`sort_order` 支持 `asc`/`desc`。

**API 输入校验**：字段级严格校验 + 参数化 SQL（所有 SQL 构造使用 `?` 占位符，绝不通过 `fmt.Sprintf` 拼接值）+ 规则变更 rate limiting（5次/秒）+ 查询端点 rate limiting（30次/秒）+ ATT&CK tactic/technique 值白名单校验。

**API 认证机制（v0.1.0，P0）**：

NetSentry 的 API 暴露告警数据、规则配置和系统健康信息——这些信息若被未授权方获取将直接威胁网络安全。v0.1.0 引入 **Pre-Shared Key (PSK)** 认证。

```
实现方式：HTTP 中间件（middleware）
  - 检查 Authorization: Bearer <token> 头
  - 校验 token 是否等于 config.yaml 中的 engine.api_auth_token
  - 匹配失败 → 401 Unauthorized（含统一错误格式响应体）
  - 未配置 token（默认空字符串）→ 打印 WARN 日志，允许所有请求（开发模式）

排除路径（无需认证）：
  - /api/health（K8s liveness/readiness probe 需要无鉴权访问）
  - /api/metrics（Prometheus scrape 通常不支持 Bearer token，可改为 Basic auth 或仅监听 127.0.0.1）
  - v0.2.0: /api/metrics 增加独立的 metrics_token 配置

配置项：
  engine:
    api_auth_token: "${NETSENTRY_API_TOKEN}"   # 支持环境变量展开，默认空 = 开发模式
    api_auth_enabled: true                     # 显式开关
```

**PSK 认证机制的已知局限性（v0.1.0 诚实标注）**：

- PSK 是静态共享密钥，无 token 轮换、无 token 撤销机制
- 多人协作时，所有人共享同一个 token，无法区分操作者身份
- 审计日志的 `user_identity` 字段在单一 PSK 模式下固定为 `"psk-default"`，不提供用户级审计
- v0.2.0 Roadmap 增加 JWT/OAuth2 支持，实现真正的用户身份认证

**安全最佳实践**：
- Token 不得硬编码在 `config.yaml` 中，必须通过环境变量注入（`${ENV_VAR}` 语法）
- Token 长度最低要求 32 字符（建议 `openssl rand -hex 32` 生成）
- 生产环境强制启用认证：若 `api_auth_enabled: true` 且 token 为空，启动时 FATAL 退出

**敏感数据脱敏**：

API 返回的 `payload_preview` 字段可能包含 HTTP 头中的敏感信息（Cookie、Authorization、POST 表单密码等）。v0.1.0 引入可配置的自动脱敏：

```yaml
engine:
  redact_sensitive_fields: true  # 默认启用
  redact_patterns:               # 可自定义正则模式
    - 'Cookie:\s*[^\r\n]+'       # Cookie: xxx → Cookie: [REDACTED]
    - 'Authorization:\s*[^\r\n]+' # Authorization: xxx → Authorization: [REDACTED]
    - 'Set-Cookie:\s*[^\r\n]+'   # Set-Cookie: xxx → Set-Cookie: [REDACTED]
    - 'password=[^&\s]+'         # password=xxx → password=[REDACTED]
    - 'passwd=[^&\s]+'           # passwd=xxx → passwd=[REDACTED]
    - 'token=[^&\s]+'            # token=xxx → token=[REDACTED]
```

**实现**：Go 端在告警写入 SQLite 之前（`alert/aggregator.go` 的 `Finalize()` 方法中），执行 `regexp.ReplaceAllString(payload_preview, "[REDACTED]")` 替换。原始 payload 仍保留在 `raw_payload` 字段中（该字段仅通过 `/api/alerts/:id` 详情端点返回，且同样执行脱敏）。

**CORS 配置**：`config.yaml` 数组配置：
```yaml
cors_allowed_origins: ["http://localhost:3000"]   # dev
# cors_allowed_origins: ["https://dashboard.internal.example.com"]  # prod
```
设 `Access-Control-Allow-Methods: GET, POST, PUT, PATCH, DELETE`，不设 `*`。PSK 认证与 CORS 是正交的安全层级——CORS 防止浏览器端 CSRF，PSK 认证防止未授权 API 直接调用。

**Prometheus 指标**：

| 指标名 | 类型 | 说明 |
|--------|------|------|
| `netsentry_packets_received_total` | Counter | 总包数 |
| `netsentry_packets_dropped_total` | Counter | C 端丢包数（来自心跳帧） |
| `netsentry_alerts_total` | CounterVec (severity) | 告警按严重级别 |
| `netsentry_rules_loaded` | Gauge | 当前加载规则数 |
| `netsentry_rules_enabled` | Gauge | 当前启用规则数 |
| `netsentry_match_duration_seconds` | Histogram | 匹配耗时分布 |
| `netsentry_channel_depth` | Gauge | Worker channel 当前队列深度 |
| `netsentry_sqlite_write_seconds` | Histogram | SQLite 写入延迟 |
| `netsentry_capture_connected` | Gauge | C 模块连接状态（0/1） |
| `netsentry_capture_restarts_total` | Counter | C 进程重启次数（来自 session_id 变化） |
| `netsentry_capture_parse_errors_total` | Counter | C 端解析错误数（来自心跳帧） |
| `netsentry_capture_dropped_total` | Counter | C 端丢包数（来自心跳帧） |
| `netsentry_json_serialize_duration_seconds` | Histogram | **C 端 JSON 序列化耗时**（来自心跳帧中的性能采样字段），buckets: 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0 |
| `netsentry_json_unmarshal_duration_seconds` | Histogram | **Go 端 JSON 反序列化耗时**（Go 内部测量），buckets: 0.001, 0.005, 0.01, 0.05, 0.1, 0.5, 1.0 |
| `netsentry_uds_write_errors_total` | Counter | **UDS 写错误次数**（C 端 `send()` 返回 -1 的次数，不含 EPIPE 重连），用于检测 UDS 缓冲区满 |
| `netsentry_uds_bytes_sent_total` | Counter | UDS 发送总字节数（间接测量吞吐量） |
| `netsentry_early_exit_total` | Counter | **Early exit 触发次数**——衡量 ip_blacklist 对 AC 自动机 CPU 的节省效果 |
| `netsentry_actual_throughput_pps` | Gauge | **Worker 实际处理速度（PPS）**——每秒实际处理的包数，与 `netsentry_packets_received_total` 对比暴露降速幅度 |
| `netsentry_worker_utilization_pct` | Gauge | **Worker CPU 占用百分比估算**——通过处理延迟/采样间隔计算，预警单协程瓶颈 |
| `netsentry_db_dir_free_bytes` | Gauge | **磁盘剩余空间**（字节），每 30s 更新——触发紧急模式的关键指标 |
| `netsentry_db_emergency_mode` | Gauge | **磁盘降级状态**（0/1）——1 表示磁盘空间不足（ < 50MB），所有 SQLite 写入已暂停，后续告警直接丢弃 |

**Histogram Buckets 说明**：
- JSON 序列化/反序列化的 buckets 集中在 1ms–1s 范围——正常情况应在 0.01ms（10μs）级别完成，若持续高于 0.5ms 则表明 JSON 开销成为瓶颈，触发 v0.2.0 FlatBuffers 迁移决策。
- UDS 写错误 `netsentry_uds_write_errors_total` 排除 EPIPE（那是正常的重连触发），聚焦于真正的缓冲区满（`EWOULDBLOCK` / `ENOBUFS`）——这些错误表明 Go 端消费速度跟不上 C 端生产速度。

**Go pprof 性能剖析端点（P0）**：

生产环境中，性能瓶颈（CPU 热点、内存泄漏、Goroutine 阻塞）往往难以仅通过 Prometheus 指标直接定位。v0.1.0 引入 Go pprof HTTP 端点用于深度故障诊断：

```go
// engine/internal/api/router.go — 调试路由组
import _ "net/http/pprof"

// 在 debug 路由组中注册（受 PSK 保护或仅监听 localhost）：
debugGroup := r.Group("/debug")
debugGroup.GET("/pprof/*any", gin.WrapH(http.DefaultServeMux))
```

**安全配置**：
- **生产环境**：pprof 端点仅监听 `127.0.0.1`，使用独立的 HTTP server 监听独立端口（6060），与 API 端口（8080）分离。注意：在 K8s Pod 内所有容器共享 network namespace，`127.0.0.1` 不提供跨容器隔离——生产集群应通过 NetworkPolicy 限制 6060 端口访问
- **开发环境**：可配置监听所有接口，但必须有 PSK 保护
- **启动校验**：若 pprof 监听地址非 `127.0.0.1` 且无 PSK 保护，打印 WARN 日志
- pprof 暴露的 goroutine 堆栈和堆内存数据可能包含敏感业务信息，务必限制访问范围
- **访问示例**：
  - 30s CPU profile：`go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30`
  - Goroutine dump：`curl http://localhost:8080/debug/pprof/goroutine?debug=2`
  - Heap profile：`go tool pprof http://localhost:8080/debug/pprof/heap`
  - 所有 profile 列表：`curl http://localhost:8080/debug/pprof/`

**价值**：运维人员可在服务运行期间实时抓取性能剖析数据进行分析，无需重启服务或另行编译 debug 版本。

**OpenTelemetry Trace Context 兼容性设计**：

文档当前的 `trace_id` 用于日志关联，但为了适应现代可观测性栈（Jaeger、Grafana Tempo），v0.1.0 使 trace_id 兼容 W3C Trace Context 标准（`https://www.w3.org/TR/trace-context/`）：

```go
// engine/internal/pipeline/trace.go — Trace ID 生成兼容 W3C Trace Context
import (
    "crypto/rand"
    "encoding/hex"
)

// GenerateTraceID 生成符合 W3C Trace Context 的 16 字节 trace ID
// 格式: 32 字符 HEX 字符串（16 字节随机数）
func GenerateTraceID() string {
    b := make([]byte, 16)
    rand.Read(b) // 使用 crypto/rand，非 math/rand（安全性要求）
    return hex.EncodeToString(b)
}
```

**关键差异**：
- **W3C 标准**：16 字节（32 位十六进制），如 `"4bf92f3577b34da6a3ce929d0e0e4736"`
- **原 UUID v7**：36 字符（含连字符），如 `"a1b2c3d4-e5f6-7890-abcd-ef1234567890"`
- **v0.1.0 采用**：16 字节 W3C 格式——兼容 Jaeger/Grafana Tempo 的原生 trace 查询，同时仍然满足日志关联需求
- **向后兼容**：Go 端同时接受 UUID 格式和 W3C 格式的 `trace_id`（通过长度检测：36 字符 = UUID，32 字符 = W3C）

**全链路追踪价值**：
- 在 Jaeger 中直接追踪一个数据包从 C 端解析 → Go 端匹配 → 写入 SQLite 的全链路耗时
- 无需额外的 trace collector——只需在日志中嵌入 `trace_id`，Grafana Loki 的 Derived Fields 即可关联到 Tempo
- v0.2.0 引入真正的 OpenTelemetry SDK（`go.opentelemetry.io/otel`），将 trace_id 自动注入 Span Context

**`/api/health?verbose=true` 响应**：

```json
{
  "status": "ok",
  "uptime_seconds": 3600,
  "version": "0.1.0",
  "data_freshness_seconds": 2.3,
  "components": {
    "capture": {
      "status": "connected",
      "last_heartbeat_ago_ms": 1200,
      "session_id": "a1b2c3d4",
      "restarts_total": 0,
      "parse_errors_total": 3,
      "packets_sent_total": 98765
    },
    "engine": {
      "status": "ok",
      "goroutines": 8,
      "heap_alloc_mb": 45,
      "channel_depth": 342,
      "channel_capacity": 10000,
      "channel_dropped_total": 0
    },
    "sqlite": {
      "status": "ok",
      "wal_size_mb": 2,
      "write_latency_p99_ms": 5,
      "current_db": "alerts_2026-06-25.db"
    }
  },
  "throughput": {
    "packets_total": 98765,
    "packets_dropped": 0,
    "alerts_total": 67,
    "alerts_by_severity": {"critical": 2, "high": 15, "medium": 30, "low": 20}
  }
}
```

**结构化日志规范（JSON Structured Logging）**：

所有 Go 端日志使用统一的 JSON 格式，便于接入 ELK/Loki/Datadog 等日志系统：

```json
{
  "level": "info",
  "ts": "2026-06-25T10:00:00.123Z",
  "caller": "engine/pipeline/worker.go:89",
  "msg": "Packet processed",
  "latency_ms": 0.5,
  "alert_fired": true,
  "rule_id": "rule-001",
  "trace_id": "evt_a1b2c3d4e5f6"
}
```

**字段规范**：

| 字段 | 类型 | 说明 |
|------|------|------|
| `level` | string | `debug` / `info` / `warn` / `error` / `fatal` |
| `ts` | string | ISO 8601 UTC 时间戳，毫秒精度 |
| `caller` | string | `包/文件.go:行号` 格式，由 `zap`/`zerolog` 自动填充 |
| `msg` | string | 人类可读的日志消息（英文） |
| `trace_id` | string | 关联同一数据包的 event_id，贯穿解析→匹配→告警→写入全链路。v0.1.0 采用 W3C Trace Context 兼容格式（16 字节随机数的 32 字符 HEX 字符串），便于直接接入 Jaeger/Tempo |
| `latency_ms` | float | 可选，操作耗时（毫秒） |
| `error` | string | 仅 `level=error` 时出现，包含错误详情 |
| `stack` | string | 仅 `level=error/fatal` 时出现，goroutine 堆栈跟踪（可配置关闭） |

**全链路追踪**：
- `trace_id` 贯穿一次数据包处理的完整生命周期：C 端初始化 → UDS 传输 → Go 端接收 → Worker 匹配 → 告警聚合 → SQLite 写入
- C 端在 PacketInfo JSON 中附带 `trace_id` 字段（W3C Trace Context 兼容的 16 字节随机数 HEX 编码），Go 端所有日志沿袭此 ID
- 告警写入 SQLite 后，`trace_id` 存储在 `alerts.event_id` 字段中——通过 API 查询告警后可直接用其 `event_id` 在 Jaeger/Grafana Tempo 中回溯全链路耗时（而不仅是日志关联）

**实现**：Go 端使用 `go.uber.org/zap` 的 `zap.NewProduction()` 配置，输出 JSON 到 `stdout`（可重定向到文件）。C 端日志保持在 `stderr`（v0.1.0 不引入 C 端 JSON 日志依赖，保持依赖最小化——`make quickstart` 将 C 端 stderr 重定向到 `logs/capture.log`）。

---

## 五、自定义规则格式设计

### 5.1 完整规则 JSON Schema

```json
{
  "id": "rule-001",
  "name": "SQL注入特征检测",
  "type": "payload_match",
  "severity": "high",
  "priority": 150,
  "enabled": true,
  "early_exit": false,
  "config": {
    "keywords": ["UNION SELECT", "SELECT FROM", "DROP TABLE", "OR 1=1", "' OR '1'='1"],
    "case_insensitive": false,
    "protocols": ["TCP"],
    "ports": [80, 8080, 443],
    "direction": "dest",
    "depth": 4096,
    "offset": 0
  },
  "mitre_techniques": [
    {
      "tactic": "Initial Access",
      "technique_id": "T1190",
      "technique_name": "Exploit Public-Facing Application"
    }
  ],
  "description": "检测 HTTP 流量中的 SQL 注入攻击特征"
}
```

**字段说明**：
- `priority`：规则优先级（0–1000，默认 100）。数值越高越先匹配。ip_blacklist 建议 500–1000，payload_match 建议 100–499，信息收集类建议 0–99
- `early_exit`：若命中且 severity 为 critical，是否跳过后序规则匹配（默认 false）。仅 ip_blacklist 等"确认即阻断"的规则应设为 true
- `case_insensitive`：是否忽略大小写匹配（默认 `true`，构建 AC 自动机时将 pattern 和输入均 `ToLower` 折叠）。v0.1.0 实现——复杂度极低（构建和匹配各一次 strings.ToLower），防止最低级的 `Union Select` 绕过
- `config`：嵌套的类型特定配置对象（替代扁平字段），与数据库 `config_json` 字段直接对应

### 5.2 各类型规则 ATT&CK 映射参考

| 规则类型 | 攻击场景 | 对应 ATT&CK |
|----------|---------|------------|
| `payload_match` SQL注入 | Web应用攻击 | TA0001 Initial Access → T1190 |
| `payload_match` XSS | 钓鱼/水坑 | TA0001 Initial Access → T1189 |
| `ip_blacklist` 已知C2 IP | 命令控制 | TA0011 C2 → T1571 / T1090 |
| `port_blacklist` 非标端口 | C2/数据外传 | TA0011 C2 → T1571 / TA0010 Exfil → T1048 |
| `frequency_threshold` 端口扫描 | 资产发现 | TA0007 Discovery → T1046 |
| `frequency_threshold` 暴力破解 | 凭证攻击 | TA0006 Credential Access → T1110 |
| 流量突增 | DDoS | TA0040 Impact → T1498 |

### 5.3 API ATT&CK 筛选

```
GET /api/alerts?mitre_tactic=Discovery
GET /api/alerts?mitre_technique_id=T1046
GET /api/alerts?mitre_tactic=Initial+Access&severity=critical
```

### 5.4 告警查询响应示例

```json
// GET /api/alerts?severity=high&page=1&per_page=2
{
  "data": [
    {
      "id": "alert-0001",
      "rule_id": "rule-001",
      "rule_name": "SQL注入特征检测",
      "timestamp": "2026-06-25T10:30:00Z",
      "src_ip": "10.0.0.99",
      "dst_ip": "192.168.1.10",
      "dst_port": 80,
      "severity": "high",
      "aggregated_count": 12,
      "mitre_tactic": "Initial Access",
      "mitre_technique_id": "T1190",
      "matched_keyword": "UNION SELECT"
    }
  ],
  "pagination": {
    "page": 1,
    "per_page": 2,
    "total": 67
  }
}
```

---

## 六、项目目录结构

```
netsentry/
├── README.md                    # 含竞品对比表 + 5行curl快速开始 + 系统要求
├── CONTRIBUTING.md              # 贡献指南
├── CHANGELOG.md                 # 版本变更记录
├── SECURITY.md                  # 安全漏洞报告流程
├── CODE_OF_CONDUCT.md           # 社区行为准则
├── SUPPORT.md                   # 支持渠道说明
├── LICENSE                      # MIT
├── config.yaml
├── Makefile                     # make quickstart + make bench + make test
├── .github/
│   ├── workflows/
│   │   ├── ci.yml               # 编译+测试+lint+ASan+race detector+许可证扫描
│   │   └── release.yml          # 静态二进制Release构建
│   ├── PULL_REQUEST_TEMPLATE.md
│   └── ISSUE_TEMPLATE/
│       ├── bug_report.md
│       └── feature_request.md
├── .devcontainer/
│   └── devcontainer.json
├── docs/
│   ├── architecture.md
│   ├── rule-guide.md
│   ├── api-reference.md         # 含分页、错误格式、所有筛选参数文档
│   ├── development.md
│   ├── troubleshooting.md       # 故障排查指南
│   └── interview-prep.md        # 面试准备（含系统设计扩展讨论）
├── configs/
│   ├── rules.json               # 首次种子导入（10-15条精选规则）
│   ├── rules.example.json
│   └── suppressions.json
├── capture/                     # C 模块
│   ├── Makefile
│   ├── deps/ (cJSON, unity)
│   ├── src/
│   │   ├── main.c
│   │   ├── pcap_offline.c
│   │   ├── eth_parser.c         # 含 VLAN tag 跳过循环
│   │   ├── ip_parser.c          # 含 fragment 检查
│   │   ├── tcp_parser.c         # 每字段含边界检查
│   │   ├── json_serializer.c    # goto cleanup 模式, cJSON_PrintBuffered
│   │   ├── uds_client.c         # 数据帧+心跳帧发送, 指数退避重连
│   │   ├── heartbeat.c          # timerfd_create + poll 集成
│   │   └── signal_handler.c     # volatile sig_atomic_t + 主循环检查
│   ├── include/
│   │   ├── packet_types.h
│   │   ├── parser.h
│   │   ├── parser_registry.h    # (protocol, port) 二元组 key
│   │   ├── uds_client.h
│   │   └── config.h
│   └── tests/
│       ├── test_parser.c        # 含畸形包 + VLAN + fragment 测试用例
│       └── test_serializer.c    # 含全引号转义边界测试
├── engine/                      # Go 模块
│   ├── main.go
│   ├── internal/
│   │   ├── config/
│   │   ├── receiver/            # uds_listener.go, heartbeat.go (atomic.Value)
│   │   ├── pipeline/
│   │   │   └── interfaces.go    # Matcher + AlertWriter 接口定义
│   │   ├── rule/
│   │   │   ├── engine.go        # RuleEngine (含 sync.RWMutex)
│   │   │   ├── rule.go, loader.go
│   │   │   ├── ip_blacklist.go
│   │   │   ├── payload_match.go # depth/offset/case_insensitive 支持
│   │   │   └── ac_builder.go
│   │   ├── alert/
│   │   │   ├── alert.go         # 含 MITRE ATT&CK 字段 + event_id
│   │   │   ├── aggregator.go    # UPSERT + 过期窗口定时清理
│   │   │   ├── suppressor.go    # CIDR Trie 匹配
│   │   │   ├── buffer.go, store.go
│   │   │   └── walog.go         # fsync + event_id 幂等 + 轮转
│   │   ├── stats/               # atomic.Uint64, Prometheus NewCounterFunc
│   │   ├── health/              # monitor.go (含 verbose 模式)
│   │   ├── api/
│   │   │   ├── router.go, auth.go
│   │   │   ├── metrics.go       # Prometheus 端点
│   │   │   ├── validator.go     # 输入校验（CIDR/ATT&CK白名单）
│   │   │   ├── errors.go        # 统一错误响应格式
│   │   │   ├── pagination.go    # 分页 envelope
│   │   │   ├── audit.go         # 审计日志中间件（记录所有非 GET 操作）
│   │   │   ├── alert_handler.go, rule_handler.go
│   │   │   └── suppression_handler.go
│   │   └── signal/              # signal.go (context.WithCancel + WaitGroup)
│   ├── pkg/model/               # packet.go, alert.go (MITRE字段), rule.go
│   └── tests/
│       ├── test_matcher.go
│       ├── test_concurrency.go  # 规则热加载 + 匹配并发测试
│       ├── test_aggregator.go   # UPSERT 幂等测试
│       ├── test_api.go
│       └── test_walog.go        # 截断行容错 + 幂等重放测试
├── testdata/                    # sample.pcap, attack_traffic.pcap, vlan_tagged.pcap
│   ├── raw/                     # 原始生产 pcap（.gitignore，需脱敏后方可使用）
│   └── sanitized/               # 脱敏后的 pcap（可安全提交）
├── data/                        # netsentry.db, alerts_YYYY-MM-DD.db, alert_wal_YYYY-MM-DD.jsonl, uds_socket.sock
└── scripts/                     # setup.sh, gen_test_pcap.py, quickstart.sh, sanitize_pcap.py, encrypt_config.go
```

---

## 七、配置管理设计

```yaml
capture:
  mode: "offline"
  offline_file: "testdata/attack_traffic.pcap"
  payload_preview_len: 4096
  uds_socket_path: "/tmp/netsentry.sock"
  uds_socket_mode: "0600"           # UDS socket 文件权限（八进制）。Go 端 listen() 前通过 syscall.Umask(0o077) 设置 umask，确保 socket 创建时即为 0600，消除 listen() 与 chmod() 之间的权限竞态窗口。listen() 后恢复 umask 原值
  heartbeat_interval: 5

engine:
  uds_socket_path: "/tmp/netsentry.sock"
  channel_buffer_size: 10000
  worker_count: 1                  # v0.1.0 单协程（无竞态）；v0.2.0 拆多 stage
  db_dir: "data/"                  # 告警按天分库存放目录
  db_path: "data/netsentry.db"     # 规则/抑制规则主库
  db_journal_mode: "WAL"
  db_busy_timeout: 5000
  db_wal_autocheckpoint: 1000
  db_checkpoint_on_shutdown: true
  rules_seed_file: "configs/rules.json"
  suppressions_file: "configs/suppressions.json"
  api_port: 8080
  cors_allowed_origins: ["http://localhost:3000"]
  alert_batch_size: 100
  alert_batch_interval: 5
  alert_aggregation_window: 60    # 告警聚合窗口（秒）
  alert_aggregation_max_count: 100  # 单窗口最大聚合条数
  alert_retention_days: 7         # 告警数据保留天数
  wal_retention_days: 3           # 预写日志保留天数
  health_freshness_limit_seconds: 30

logging:
  level: "info"
  engine_log: "logs/engine.log"
  format: "json"                     # json (生产) / text (开发)
  redact_sensitive_in_logs: true     # 日志中脱敏 IP/mac/payload？(默认 true，原始数据仅在 API 和 DB 中可见)
```

### 7.1 环境变量展开

配置值支持 `${ENV_VAR}` 和 `${ENV_VAR:default}` 两种语法：

```yaml
engine:
  api_auth_token: "${NETSENTRY_API_TOKEN}"            # 必须从环境变量读取，未设置 → 启动时 FATAL
  api_auth_token: "${NETSENTRY_API_TOKEN:dev-token}"  # 优先读环境变量，未设置时使用默认值 "dev-token"
  db_dir: "${NETSENTRY_DATA_DIR:/var/lib/netsentry}"  # 有默认值的路径配置
```

**解析规则**：
- `${ENV_VAR}`：必须存在，否则启动 FATAL 并列出所有缺失的环境变量
- `${ENV_VAR:default}`：可选，未设置时使用 `default` 值
- 默认值可以为空：`${ENV_VAR:}` → 空字符串
- 嵌套不支持（避免复杂度），如 `${A_${B}}` 不解析

### 7.2 敏感配置加密（v0.1.1，已从 v0.1.0 降级）

生产环境中，部分敏感配置（数据库连接字符串、云服务 API Key）即使通过环境变量注入也可能残留于配置文件。v0.1.0 提供一个简单的配置加密工具：

**加密工具**（`scripts/encrypt_config.go`）：

```bash
# 加密敏感字段
go run scripts/encrypt_config.go --key "$(openssl rand -hex 32)" \
    --in config.yaml --out config.enc.yaml

# 加密后的值在原文件中被替换为 ENC[...] 标记
# api_auth_token: "ENC[AES256:9f8a7b6c...]"

# 启动时通过 --decrypt-key 参数解密
./netsentry-engine --config config.enc.yaml --decrypt-key "${NETSENTRY_CONFIG_KEY}"
```

**实现**：
- 加密算法：AES-256-GCM（`crypto/aes` + `crypto/cipher`），提供认证加密（防篡改）
- 加密范围：仅加密值（如 `"my-secret-token"`），保留 YAML 结构可读性
- 密钥管理：生产环境密钥通过 K8s Secret / Vault / 环境变量注入，仅在内存中存在
- 降级方案：未配置 `--decrypt-key` 时，`ENC[...]` 标记的值按原样传递（Go 端检测到 `ENC[...]` 前缀但无解密密钥时 FATAL）

**安全边界说明**：
- 这不是完整密钥管理系统（KMS）的替代品——v0.3.0 引入 HashiCorp Vault 集成
- 核心目标是防止配置文件被提交到 Git 或日志中泄露敏感信息（如 `config.yaml` 可以被安全地提交到仓库，因为敏感字段已加密）

### 7.3 SIGHUP 配置热加载（v0.1.1，已从 v0.1.0 降级）

文档已实现规则热加载（POST /api/rules），但系统配置（如 `channel_buffer_size`、`api_port`）的热加载同样重要——避免因修改日志级别等非关键配置而重启服务导致数据丢失。

**SIGHUP 信号处理机制**：

```go
// engine/internal/signal/signal.go — SIGHUP 处理器
func HandleSIGHUP(ctx context.Context, cfg *config.Config, reloadFn func(*config.Config) error) {
    sigCh := make(chan os.Signal, 1)
    signal.Notify(sigCh, syscall.SIGHUP)
    
    go func() {
        for {
            select {
            case <-sigCh:
                log.Info("Received SIGHUP, reloading configuration")
                newCfg, err := config.Load(cfg.Path) // 重新读取 config.yaml
                if err != nil {
                    log.Error("Failed to reload config", zap.Error(err))
                    continue
                }
                
                // 动态可调整参数 — 立即应用
                cfg.Logging.Level = newCfg.Logging.Level
                cfg.AlertAggregationWindow = newCfg.AlertAggregationWindow
                
                // 不可动态调整参数 — 仅记录 WARN
                if cfg.APIPort != newCfg.APIPort {
                    log.Warn("api_port changed, requires restart",
                        zap.Int("current", cfg.APIPort),
                        zap.Int("new", newCfg.APIPort))
                }
                if cfg.ChannelBufferSize != newCfg.ChannelBufferSize {
                    log.Warn("channel_buffer_size changed, requires restart",
                        zap.Int("current", cfg.ChannelBufferSize),
                        zap.Int("new", newCfg.ChannelBufferSize))
                }
                
                if reloadFn != nil {
                    reloadFn(cfg) // 回调通知组件更新
                }
                
            case <-ctx.Done():
                signal.Stop(sigCh)
                return
            }
        }
    }()
}
```

**动态 vs 静态参数分类**：

| 可动态调整（SIGHUP 立即生效） | 需重启（仅打印 WARN） |
|------------------------------|----------------------|
| `logging.level` | `api_port` |
| `logging.format` | `channel_buffer_size` |
| `alert_aggregation_window` | `uds_socket_path` |
| `alert_batch_size` | `worker_count` |
| `alert_batch_interval` | `db_path` / `db_dir` |
| `health_freshness_limit_seconds` | `db_journal_mode` |
| `redact_sensitive_in_logs` | `api_auth_token` |

**验收标准**：
- `kill -SIGHUP $(pidof netsentry-engine)` → 日志级别从 `info` 变为 `debug`，无需重启
- 动态调整参数在 1 秒内生效（config.Load 是纯文件 I/O，无网络依赖）
- 不可动态调整的参数变更后，日志中明确提示需重启，且当前运行不中断

---

## 八、开发环境搭建

### 8.1 系统要求

在运行 `make quickstart` 前，请确保已安装以下依赖：

```bash
# Ubuntu/Debian
sudo apt update
sudo apt install -y build-essential gcc make libpcap-dev golang-go python3 python3-pip curl jq
pip3 install scapy

# 验证
gcc --version        # >= 9.0
go version           # >= 1.21
pcap-config --version # >= 1.9
```

### 8.2 CLI 输出格式（pcap 取证场景首选）

对于 pcap 快速取证场景，NetSentry 的 CLI 模式（`--output table`）比 REST API 更友好——直接在终端打印带 ANSI 颜色和 MITRE ATT&CK 标签的格式化告警表格：

```bash
# 默认 CLI 表格输出（无需 jq，无需 curl API）
docker run --rm -v ./suspicious.pcap:/data/input.pcap:ro \
    ghcr.io/yourusername/netsentry:v0.1.0 \
    --pcap /data/input.pcap --output table

# 输出示例：
# ┌──────────┬──────────┬──────────────┬─────────────────────┬────────────────────┐
# │ SEVERITY │ RULE ID  │ SRC IP       │ MITRE TECHNIQUE     │ PAYLOAD PREVIEW    │
# ├──────────┼──────────┼──────────────┼─────────────────────┼────────────────────┤
# │ HIGH     │ rule-001 │ 10.0.0.99    │ T1190 (Exploit...)  │ GET /search?q=UN...│
# │ CRITICAL │ rule-010 │ 192.168.1.50 │ T1571 (Non-Std Po...│ [IP BLACKLIST]     │
# └──────────┴──────────┴──────────────┴─────────────────────┴────────────────────┘
# Total: 2 alerts | Pcap: 15234 packets | Time: 2.3s
```

`--output` 参数支持：
- `table`（默认）：ANSI 彩色表格，适合终端直接阅读
- `json`：每行一个 JSON 对象，适合 `| jq` 管道处理
- `jsonl`：同 json，每行一条记录
- REST API 模式（`--serve`）：启动 HTTP server 在 8080 端口，适合持久化部署

CLI 模式不启动 SQLite（告警仅输出到 stdout），适用于 pcap 文件的一次性快速分析。REST API 模式才启用完整的 SQLite 持久化 + 查询功能。

### 8.3 快速开始（make quickstart）

```bash
git clone https://github.com/yourusername/netsentry.git
cd netsentry
make quickstart
# 自动完成：编译 C + Go → 生成测试 pcap → 启动 Go → 启动 C → curl 验证
# 输出:
#   === NetSentry v0.1.0 Quickstart ===
#   [1/5] Building C capture module... ok (gcc -std=c11 -O2)
#   [2/5] Building Go engine... ok (go build -race)
#   [3/5] Generating test pcap (scapy)... ok (25 packets)
#   [4/5] Starting engine + capture...
#   [5/5] Verifying API... ok
#   === Ready! ===
#   curl http://localhost:8080/api/alerts | jq .
```

### 8.3 二进制 Release 安装

```bash
curl -L https://github.com/yourusername/netsentry/releases/download/v0.1.0/netsentry-linux-amd64.tar.gz | tar xz
./netsentry --quickstart
curl http://localhost:8080/api/alerts | jq .
```

### 8.4 Docker 一键运行（✅ 推荐分发方式）

Docker 镜像是 NetSentry 的**主要分发方式**——用户无需安装任何编译依赖（无 gcc、Go、libpcap-dev），`docker run` 即可在 10 秒内看到 ATT&CK 告警输出。静态编译的二进制打入最小化容器（`scratch` 或 `alpine`），镜像体积目标 < 20MB。

```bash
# 仅需 Docker，10 秒跑通
docker run --rm -v $(pwd)/mypcap.pcap:/data/input.pcap:ro \
    ghcr.io/yourusername/netsentry:v0.1.0 \
    --pcap /data/input.pcap --output json

# 预期输出（stdout）:
# {"level":"info","msg":"Scanning pcap...","packets_total":15234}
# {"event_id":"evt_xxx","rule_id":"rule-001","severity":"high",
#  "mitre_technique_id":"T1190","payload_preview":"GET /search?q=UNION SELECT..."}
```

Dockerfile 关键特征：
- 多阶段构建：stage 1 编译 C（gcc -static）+ Go（CGO_ENABLED=0），stage 2 从 `scratch` 或 `alpine:3.20` 复制二进制
- 静态链接：C 端静态链接 `libpcap.a`，Go 端 `CGO_ENABLED=0`，无运行时动态库依赖
- CI 自动构建：GitHub Actions 在 tag push 时构建 `linux/amd64` 和 `linux/arm64` 双架构镜像
- 构建命令参考：
  - Go 模块：`CGO_ENABLED=0 GOOS=linux go build -race -ldflags="-s -w" -o netsentry-engine`
  - C 模块：`gcc -std=c11 -O2 -Wall -Wextra -static -o netsentry-capture`（需 `libpcap.a`）
  - ASan 构建：`gcc -std=c11 -fsanitize=address -g -o netsentry-capture-asan`（测试用）

### 8.5 流量脱敏工具（`make sanitize-pcap`）

在企业开发流程中，开发人员往往需要使用生产环境的真实流量进行调试和规则测试，这带来了严重的隐私泄露风险。v0.1.0 集成流量脱敏脚本，强制要求使用生产流量调试时必须经过脱敏处理。

**工具集成**：

```makefile
# Makefile — sanitize-pcap 目标
.PHONY: sanitize-pcap
sanitize-pcap: ## 脱敏原始 pcap 文件（IP 掩码 + Payload 替换）
	@echo "=== Sanitizing pcap files ==="
	@mkdir -p testdata/sanitized
	@for f in testdata/raw/*.pcap; do \
		echo "  Processing $$f..."; \
		python3 scripts/sanitize_pcap.py \
			--input "$$f" \
			--output "testdata/sanitized/$$(basename $$f)" \
			--mask-ip-last-octet \
			--replace-payload "0x41" \
			--preserve-headers; \
	done
	@echo "=== Done. Sanitized files in testdata/sanitized/ ==="
```

**脱敏脚本**（`scripts/sanitize_pcap.py`）：

```python
#!/usr/bin/env python3
"""
流量脱敏工具 — 使用 Scapy 处理 pcap 文件。
在保留协议结构和攻击特征的同时，匿名化敏感信息。
"""
import argparse
from scapy.all import rdpcap, wrpcap, IP, TCP, UDP, Raw, Ether
import random

def sanitize_pcap(input_path, output_path, mask_ip=True, replace_payload=True, preserve_headers=True):
    packets = rdpcap(input_path)
    sanitized = []
    
    for pkt in packets:
        # 1. IP 地址掩码：将源/目 IP 的最后一段替换为随机值
        if mask_ip and IP in pkt:
            pkt[IP].src = anonymize_ip(pkt[IP].src)
            pkt[IP].dst = anonymize_ip(pkt[IP].dst)
            # 清除 IP 校验和（让 Scapy 重新计算）
            del pkt[IP].chksum
        
        # 2. 传输层端口随机化（保留端口范围特征）
        if TCP in pkt:
            pkt[TCP].sport = randomize_port(pkt[TCP].sport)
            pkt[TCP].dport = randomize_port(pkt[TCP].dport)
            del pkt[TCP].chksum
        elif UDP in pkt:
            pkt[UDP].sport = randomize_port(pkt[UDP].sport)
            pkt[UDP].dport = randomize_port(pkt[UDP].dport)
            del pkt[UDP].chksum
        
        # 3. Payload 替换：保留协议头，将敏感 Payload 替换为全 'A'
        if replace_payload and Raw in pkt:
            original_len = len(pkt[Raw].load)
            pkt[Raw].load = b'A' * original_len  # 保留长度特征，但内容匿名化
        
        sanitized.append(pkt)
    
    wrpcap(output_path, sanitized)
    print(f"Sanitized {len(sanitized)} packets → {output_path}")

def anonymize_ip(ip_str):
    """保留前 3 个八位组，随机化最后一个八位组"""
    parts = ip_str.split('.')
    if len(parts) == 4:
        parts[3] = str(random.randint(1, 254))
    return '.'.join(parts)

def randomize_port(port):
    """保留端口范围特征：<1024 → 随机 <1024，>=1024 → 随机 >=1024"""
    if port < 1024:
        return random.randint(1, 1023)
    else:
        return random.randint(1024, 65535)
```

**开发规范（强制）**：

> ⚠️ **使用生产环境 pcap 进行本地调试时，必须经过 `make sanitize-pcap` 脱敏处理。**
> 
> - 原始生产 pcap 文件不得提交到 Git 仓库
> - `testdata/raw/` 目录已在 `.gitignore` 中
> - 脱敏后的 pcap 存放在 `testdata/sanitized/`，可安全提交
> - CI 中的模糊测试语料库同样使用脱敏数据

**替代方案**（无需 Python 依赖）：

```bash
# 使用 tcprewrite（tcpreplay 工具集）进行 IP 掩码
sudo apt install tcpreplay
tcprewrite --srcipmap=0.0.0.0/0:10.0.0.0/8 --dstipmap=0.0.0.0/0:10.0.0.0/8 \
    --infile=raw.pcap --outfile=sanitized.pcap

# 使用 tcpdump 截断 payload（仅保留头）
tcpdump -r raw.pcap -s 96 -w sanitized.pcap  # -s 96 = 仅保留前 96 字节（含 IP+TCP 头）
```

**价值**：
- 保护用户隐私，满足 GDPR/个人信息保护合规要求
- 开发者可在本地使用接近生产的真实流量进行调试，同时不泄露敏感信息
- 脱敏后的 pcap 可作为开源社区的测试语料库（在 CC0 许可证下发布）

---

## 九、12 周开发计划（v0.1.0 核心检测管道）

> **周期说明**：2026.06–2026.09（12 周）全部聚焦 v0.1.0 核心检测管道的开发和测试。v0.2.0（实时模式、IP 分片重组、慢速扫描检测等）和 v0.3.0（Web 面板、Docker、TCP 流重组）单独立项，不包含在当前周期内。v0.2.0 预计另需 6–8 周，v0.3.0 预计另需 8–10 周。

| 周 | 任务 | 交付物 |
|----|------|-------------|
| W1 | 项目初始化 + Makefile + CI + Release CI + Unity 集成 + 开源标配文件 | 可构建骨架 + CHANGELOG/SECURITY/CODE_OF_CONDUCT |
| W2 | C：Eth（含 VLAN tag 跳过）+ IP（含 fragment 检查）+ TCP 解析 + **每字段边界检查** + **微基准测试** | 协议解析器 + 畸形包测试 + 性能基准数据 |
| W3 | C：cJSON `goto cleanup` 模式 + `cJSON_PrintBuffered` 12KB + 心跳帧（timerfd） + UDS 重连状态机 | 序列化零泄漏 + 重连状态机 |
| W4 | IPC：UDS 数据帧+心跳帧 → Go 接收 + 健康监控 + C 重启检测（session_id） | 端到端 UDS 链路打通 |
| W5 | Go：channel + Worker 框架 + **atomic.Pointer 规则引擎（lock-free）** + pipeline 接口定义 | 并发安全的检测管道骨架 |
| W6 | Go：ip_blacklist + AC 自动机（depth=4096 可配）+ Prometheus 端点 | 可匹配规则 + 指标可见 |
| W7 | Go：告警聚合器（**UPSERT**） + 抑制规则（**CIDR Trie**） + SQLite **按天分库** + MITRE ATT&CK 告警模型 | 告警去重 + 落库 + ATT&CK 映射 |
| W8 | API 契约（**分页 + 统一错误格式**） + 输入校验 + 跨天查询方案 | API 完整实现 |
| W9 | `/api/health?verbose=true` + 结构化日志 + 审计日志 + PSK 认证 + pprof 端点 | 可观测性 + 安全基础 |
| W10 | 端到端联调 + 优雅退出完整测试 + goroutine 泄漏测试 + 竞态测试 | 集成验证 |
| W11-12 | **缓冲 + 压力测试 + 性能调优 + 文档完善** | v0.1.0 发布候选 |

**v0.1.1 降级功能**（从 v0.1.0 移除，推迟至后续 minor 版本）：
- 配置加密工具（`scripts/encrypt_config.go`）— 非核心运行时组件，推迟至 v0.1.1
- SIGHUP 配置热加载 — 推迟至 v0.1.1（手动重启即可实现同等效果）
- 流量脱敏工具（`make sanitize-pcap`）— 开发辅助工具，推迟至 v0.1.1
- OpenTelemetry Trace Context 兼容 — v0.1.0 使用简单 UUID，推迟至 v0.2.0

**v0.2.0 独立周期（另立项目，6–8 周）**：
- 实时抓包 + 微批次 + FlatBuffers 序列化优化
- port/freq 规则 + IP 分片重组 + 慢速扫描检测
- HTTP/DNS 应用层解析器（parser_registry 完整实现）
- JWT/OAuth2 认证 + 批量操作 API

**v0.3.0 独立周期（另立项目，8–10 周）**：
- Web 面板 或 Python 日报 + Docker + TCP 流重组
- 跨天查询完整方案（SQLite ATTACH 多库并行查询）
- 审计日志外部 syslog 同步

---

### 社区贡献规划（Good First Issues）

以下任务被设计为外部贡献者的入门切入点，每个任务独立、有明确验收标准、不依赖其他模块：

| 难度 | 任务 | 涉及模块 | 预计工时 |
|------|------|---------|---------|
| easy | 新增 HTTP Basic 解析器（提取 Host + Content-Type） | `capture/src/` C 端 | 4h |
| easy | 新增 `payload_match` 规则 10 条（常见攻击特征） | `configs/rules.json` | 2h |
| easy | 补充 MITRE ATT&CK 映射表（现有规则的 technique 补充） | `configs/rules.json` | 2h |
| medium | 实现 Elasticsearch AlertWriter（告警直写 ES） | `engine/internal/alert/` Go 端 | 8h |
| medium | 新增 DNS 域名黑名单检测规则类型 | `engine/internal/rule/` + C 端 | 12h |
| medium | CLI `--output html` 格式（生成静态 HTML 报告） | `engine/internal/cli/` | 8h |
| hard | 新增 Modbus 协议解析器 | `capture/src/` C 端 | 16h |

所有任务在 GitHub Issues 中标记 `good first issue` 或 `help wanted`，附带详细的设计约束和接口契约。社区贡献流程在 `CONTRIBUTING.md` 中定义。

## 十、Git 工作流与 CI

**Release CI**：Tag push → 构建 C/Go 静态二进制 + ASan 构建 → 打包 tarball → GitHub Release。

**CI 检查项**：
- Go：`go vet` + `golangci-lint` + **`go test -race`（竞态检测）** + `go test -cover`
- C：`gcc -Wall -Wextra -Werror` + **ASan 构建 + 畸形包测试** + Valgrind 内存泄漏检测
- 许可证扫描：`go-licenses check ./...` + `pip-licenses`（Python 阶段）+ `license-checker`（Node.js 阶段）

---

## 十一、测试策略

### 11.1 分层测试体系

| 测试层 | 目标 | 工具 | 频率 |
|--------|------|------|------|
| **单元测试** | 每个解析函数、规则匹配器、告警聚合逻辑 | C: Unity, Go: `testing` + `-race` | 每次 commit |
| **集成测试** | UDS 通信、端到端 pcap→告警、API CRUD | Go: `httptest` + temp dir | 每次 push |
| **模糊测试** | 协议解析器鲁棒性 | AFL++ / LibFuzzer / Go fuzz | CI 每日定时 |
| **竞态测试** | 规则热加载 + Worker 匹配并发 | `go test -race -count=100` | 每次 commit |
| **内存测试** | C 端零泄漏、Go 端无 goroutine 泄漏 | Valgrind, ASan, `runtime.NumGoroutine()` | 每次 push |
| **性能测试** | PPS 吞吐、匹配延迟 p99 | `go test -bench=.` + `perf stat` | 每次 release |

### 11.2 专项测试用例

| 新增测试 | 内容 |
|----------|------|
| **VLAN tag 测试** | 单 tag / Q-in-Q 双 tag → IP 头偏移正确 |
| **IPv4 fragment 测试** | 首片/非首片/无分片 → is_fragment 标志正确 |
| **cJSON 泄漏测试** | Valgrind 10000 包序列化 → 0 泄漏 |
| **全引号转义测试** | payload 4096 字节全 `"` → 12KB 缓冲不溢出，JSON 合法 |
| **规则热加载竞态测试** | `go test -race`：Worker 匹配 + 并发 POST /api/rules × 1000 → 0 race |
| **UPSERT 幂等测试** | 同 key 100 次写入 → 1 行 aggregated_count=100 |
| **预写日志容错测试** | 截断行 → 解析跳过；event_id 重复 → 幂等跳过 |
| **UDS 重连测试** | kill Go → C 端指数退避重连 → Go 重启后恢复 |
| **优雅退出测试** | SIGINT → 9 步序列完整执行 → 无 goroutine 泄漏 |

### 11.3 协议解析器模糊测试（P0）

作为网络安全工具，协议解析器是攻击面的第一道防线——畸形包不应导致解析器崩溃。v0.1.0 引入模糊测试确保解析器在任何输入下均安全。

**测试目标**：`eth_parser.c`、`ip_parser.c`、`tcp_parser.c`、`json_serializer.c`

**工具与配置**：

```
方案 A — LibFuzzer + clang（推荐，CI 友好）:
  - 编译: CC=clang CFLAGS="-fsanitize=fuzzer,address -g" make fuzz
  - CI: 每日定时运行 1 小时，crash 文件自动归档为 GitHub Artifact
  - 适配: 每个解析器独立 fuzz target，避免状态耦合

方案 B — AFL++（深度探索，开发环境）:
  - 安装: sudo apt install afl++
  - 编译: CC=afl-clang-fast make fuzz-afl
  - 语料库: testdata/fuzz_corpus/ (合法 pcap 片段作为种子)
  - 运行: afl-fuzz -i corpus -o findings ./fuzz_parser @@
```

**Fuzz Target 接口**：

```c
// capture/tests/fuzz_parser.c — 协议解析器模糊测试入口

#include <stdint.h>
#include <stddef.h>
#include "packet_types.h"
#include "parser.h"

// LibFuzzer 入口：LLVMFuzzerTestOneInput
int LLVMFuzzerTestOneInput(const uint8_t *data, size_t size) {
    PacketInfo info;
    // 即使 data 是完全随机的垃圾字节流，parse_packet 也不应 crash
    // 仅允许的失败模式：返回错误码 !0 或静默丢弃
    int ret = parse_packet(data, size, &info);
    (void)ret;  // 不关心返回值，只关心不 crash/无 ASan 报告

    return 0;  // 非零返回值 = LibFuzzer 认为测试失败
}

// 独立 target — JSON 序列化器（输入为已解析的 PacketInfo 字段边界值）
int LLVMFuzzerTestOneInput_Serializer(const uint8_t *data, size_t size) {
    if (size < sizeof(PacketInfo)) return 0;

    const PacketInfo *info = (const PacketInfo *)data;
    char buf[12288];
    // 任意字段值组合 → 序列化不 crash，不写越界
    serialize_packet(info, buf, sizeof(buf));
    return 0;
}
```

**畸形输入覆盖清单**：

| 畸形类型 | 示例输入 | 预期行为 |
|----------|---------|---------|
| 空包 | `size=0` | 立即返回 -1 |
| 超短 Ethernet | `size=13`（不足 14 字节 ETH_HDR_LEN） | 返回 -1 |
| VLAN 标签链溢出 | 连续 50 层 VLAN tag（`0x8100` 重复） | 循环计数器上限保护，返回 -1 |
| IP IHL 越界 | `IHL=15`（60 字节）但 `total_length=20` | 返回 -1 |
| TCP data_offset 越界 | `data_offset=15`（60 字节）但 IP payload 仅 20 字节 | 返回 -1 |
| 全零字节流 | `size=65535` 全 `0x00` | 不崩溃，正确识别为无效包 |
| payload 全引号 | payload 和字段值全部为 `"` 字符 | JSON 正确转义，不截断 JSON 结构 |

**验收标准**：
- LibFuzzer 运行 1 小时 0 crash（ASan 模式）
- AFL++ 运行 24 小时 0 unique crash
- CI 每日自动运行，crash 报告通过 GitHub Issue 自动创建
- 所有已知 CVE 相关 pcap（如 CVE-2023-XXXX 格式的 fuzzing 样本）加入回归语料库

---

## 十二、优雅退出

**完整 9 步退出序列**：

```
Signal (SIGINT/SIGTERM) → context.Cancel()

1. HTTP server.Shutdown(ctx)               ← 停止接受新 API 请求（超时 5s）
2. UDS listener.Close()                     ← 停止 accept，C 端 send() 将收到 EPIPE
3. Drain packet channel                     ← 排空剩余包（超时 30s）
   ├─ Worker 处理剩余包 → 产生告警 → 进入聚合器
   └─ 超时后记录剩余未处理数
4. ticker.Stop()                            ← 停止批量写入定时器
5. flush batch buffer                       ← 写入批量缓冲中的告警
6. aggregator.FinalizeAll()                 ← finalize 所有过期窗口
   ├─ finalize 可能产生新告警
   └─ 递归排空：finalize → 新告警 → 再次 finalize（最多 3 轮，每轮超时 5s）
       第 4 轮仍有新告警 → 放弃剩余告警，ERROR 日志 "finalize drain exceeded max rounds (3), discarding N pending alerts"
       此保护防止聚合器逻辑 Bug 导致退出序列无限循环，确保进程在 30s 内完成退出
7. flush final alerts → SQLite              ← 最后一轮写入
8. SQLite WAL checkpoint + db.Close()       ← 清理 WAL，关闭数据库
9. wg.Wait()                                ← 等待所有 goroutine 退出（超时 5s）
→ exit
```

**关键原则**：
- 每个等待步骤加超时（`select { case <-done: case <-time.After(timeout): log.Warn("timeout") }`），防止无限阻塞
- 使用 `sync.WaitGroup` 追踪所有 goroutine，退出时 `wg.Done()`
- 所有 `defer` 中包含资源释放（`ticker.Stop()`, `db.Close()`, `conn.Close()`）

---

## 十三、竞品对比表（诚实版）

| 维度 | NetSentry v0.1.0 | Zeek | Suricata |
|------|:---:|:---:|:---:|
| 定位 | 生产级轻量 pcap 分析 / 边缘 IDS | 生产级 NSM | 生产级 IDS/IPS |
| 协议解析器 | 3（Eth/IP/TCP，含 VLAN） | 50+（脚本可扩展） | 30+（C 插件） |
| 规则语言 | JSON（自定义） | Zeek 脚本 | Snort 规则（ET Open 3万+） |
| 告警输出 | SQLite + Prometheus | Kafka/ES/Syslog/S3 | Kafka/ES/Redis/JSON |
| MITRE ATT&CK | ✅ 告警/规则均有 | ✅ | ✅ |
| 告警聚合去重 | ✅ UPSERT 原子操作 | ✅ | ✅ |
| TCP 流重组 | ❌（v0.2.0 Roadmap） | ✅ | ✅ |
| IP 分片重组 | ❌（v0.2.0 Roadmap） | ✅ | ✅ |
| TLS 解密 | ❌（需部署 SSL 卸载设备后方） | ✅（需密钥） | ✅（需密钥） |
| 性能 | ~50K PPS（离线） | 10Gbps 线速 | 10Gbps+ |
| 内存 | ~50MB | ~300MB | ~500MB+ |
| 启动命令 | `make quickstart` | `zeek -i eth0` | `suricata -i eth0` |
| 适用场景 | pcap 取证分析、边缘 IDS、IDS 学习 | 企业 SOC | 企业 SOC/ISP |
| 部署复杂度 | 单二进制 + 1 配置文件 | 需配置框架 | 需配置 + 规则集 |

**NetSentry 的真实护城河（v0.1.0 重新定位）**：

在 v0.1.0 聚焦"极简"而非"双语言架构"：

- **零依赖运行**：Docker 镜像 < 20MB，`docker run` 一行命令 10 秒出 ATT&CK 告警。对比 Zeek/Suricata 需要配置框架、规则集管理、ES/Kafka 依赖
- **单二进制默认输出到 stdout**：无需配置数据库即可看到结果（`--output json`），适合 `curl | jq` 工作流
- **内置 MITRE ATT&CK 映射**：告警自带 `mitre_tactic`/`mitre_technique_id`。Zeek/Suricata 需额外插件或外部映射表
- **5MB 内存空闲**：在边缘设备和树莓派上运行，Suricata 需要 500MB+
- **真正的"5 分钟取证"**：拿到 pcap → `docker run` → 终端打印 ATT&CK 告警。不需要 ELK 也不需要配置 IDS 规则集

C/Go 双语言架构对**用户**不是卖点——它是实现上述体验的工程手段（C 做 libpcap 高性能解析，Go 做并发服务）。对用户应该只看到"一个 Docker 镜像，一个命令，一堆 ATT&CK 告警输出"。① 比 Suricata 轻一个数量级（50MB vs 500MB+），适合边缘设备和资源受限环境——在这些场景中 NetSentry 是生产级选择；② 规则用 JSON 而非 Snort 语法（DevOps 更友好）；③ 适合 pcap 离线快速取证分析——拿到 pcap 文件 5 分钟内出 ATT&CK 映射结果；④ C/Go 双语言架构天然适合作为系统编程与分布式系统的简历展示项目。

---

## 十四、已知检测盲区（诚实标注）

以下攻击场景**不在 v0.1.0 覆盖范围内**，在 README 中应明确标注：

| 盲区 | 原因 | 计划 |
|------|------|------|
| 暴力破解（SSH/RDP/HTTP表单） | 需 TCP 流重组 + 认证失败协议解析 | v0.2.0 freq 规则部分覆盖 |
| SYN Flood / DDoS | 需 SYN/SYN-ACK 不对称计数器 | Roadmap |
| DNS 隧道 / DNS Amplification | 需 DNS 协议解析 | Roadmap |
| ICMP 隧道 | ICMP 解析器已实现但未检测 payload | v0.2.0 |
| C2 Beaconing | 需周期性连接间隔分析 | Roadmap |
| 横向移动（SMB/PsExec/WMI） | 需应用层协议深度解析 | Roadmap |
| Slowloris / Slow HTTP | 需 TCP 流状态跟踪 | Roadmap |
| 低速端口扫描（>60s间隔） | 60s 频率窗口盲区 | v0.2.0 增加慢速模式 |
| IP 分片绕过 | 无分片重组（首片基本信息可用） | v0.2.0 |
| TCP 分段绕过 | 无流重组 → payload_match 在 TCP 上弱 | v0.3.0 |
| **大小写变体绕过** | ✅ v0.1.0 已覆盖——AC 自动机默认 case_insensitive: true，构建和匹配时均 ToLower | — |
| **Unicode/URL 编码绕过** | `%55%4E%49%4F%4E`、UTF-16 等变体不检测 | v0.3.0 |
| **SQL 注释插入绕过** | `UNION/**/SELECT` 不匹配 | v0.3.0 |
| TLS 加密流量 | payload_match 无法检查密文 | 部署在 SSL 卸载设备后方 |
| IPv6 攻击 | 仅支持 IPv4 | Roadmap |

---

## 十五、TLS 部署指南

**NetSentry 无法检测 TLS 加密流量中的 payload 特征。** 推荐部署位置：

- **位置 A（最佳）**：SSL 卸载设备（如 Nginx/HAProxy）后方，检测解密后的东西向 HTTP 流量
- **位置 B（可用）**：内网非加密服务边界（数据库协议、遗留 HTTP 服务）
- **位置 C（有限）**：南北向流量的 TLS handshake 元数据检测（JA3/JA4 指纹，Roadmap）

---

## 十六、面试高频问题

> **文档结构说明**：本节和[第二十节](#二十系统设计面试扩展讨论)（系统设计面试扩展讨论）属于面试准备材料。在正式设计文档中保留这些内容是为了保持完整性，但最终版本建议将面试相关内容移至独立的 `docs/interview-prep.md` 文件中，使设计文档更聚焦于"如何实现"。

### Q1: 为什么是 C+Go 而非全 Go？

C 做 libpcap 调用和协议解析，Go 做并发服务和 HTTP API。具体工程理由：
- **libpcap 是 C 原生库**：虽然 Go 有 gopacket（纯 Go pcap 库），但 gopacket 在高速场景下的内存分配模式会导致显著的 GC 压力
- **进程边界为未来分布式部署预留空间**：C 模块可独立压测、独立部署为网络探针（v0.3.0 场景），Go 模块可独立升级和重启而不中断抓包
- **可调试性**：UDS + JSON 方案使 C/Go 可以独立编译、独立调试

代价是 IPC 开销（5 次拷贝），但在 50K PPS 目标性能下可接受。

### Q2: 如何控制误报？

三层机制：① 告警聚合器 `(rule_id, src_ip, dst_ip, dst_port, 60s窗口)` 去重（sqlmap 5分钟 15000 → 5条），使用 SQLite UPSERT 原子操作保证正确性；② 抑制规则/白名单（CIDR + 时间段），使用 Trie 加速 CIDR 匹配；③ critical 告警不聚合（保留完整攻击时间线）。

### Q3: 为什么不用 Hyperscan 做 AC 匹配？

v0.1.0 用 Go 原生 `cloudwego/ahocorasick`。理由：
- 项目已有 C 模块但无 CGo 调用路径——引入 CGo（非 UDS，而是通过 `import "C"` 链接 Hyperscan）会引入跨越 cgo 调用栈的额外延迟（~40-80ns/call），且 Hyperscan 需要 Intel SIMD + Vectorscan 依赖
- 对于 50K PPS 的目标性能，Go 双数组 Trie AC 自动机完全够用——Hyperscan 的优势在万兆线速场景（10Gbps+），不在本项目的设计目标内
- Hyperscan 已在 Roadmap 作为性能升级选项（届时走独立 C 模块通过 Hyperscan 匹配，结果通过 UDS 回传）

### Q4: 性能上限是多少？

v0.1.0 离线模式 ~50K PPS（单核），内存 ~50MB。一个典型的中小企业网络出口约 200Mbps，混合流量约 30K-80K PPS——v0.1.0 的离线模式可以处理这个量级的 pcap 取证分析。实时模式（v0.2.0）~30K PPS，勉强覆盖 100Mbps 链路。这不是万兆线速系统，也从不是设计目标。

### Q5: 如果一个攻击跨越了多个 TCP segment，你的 payload_match 能检测到吗？

**坦诚回答**：v0.1.0 不能。核心原因是 TCP 流重组尚未实现——payload_match 在每个 TCP segment 上独立运行，如果攻击特征被拆分到两个 segment 中（如 `UNION` 在第 N 个 segment，`SELECT` 在第 N+1 个 segment），每个 segment 单独看都不匹配，会导致漏报。

**但是**：① HTTP 请求通常在一个 TCP segment 中（MSS 1460 字节足够容纳大多数字段）；② v0.2.0 Roadmap 中 TCP 流重组是 P0 项——实现后 payload_match 将在重组后的流上运行；③ 这恰恰说明了为什么 IDS 需要流重组——这是所有 IDS 的核心难点，我理解其重要性和实现原理。

### Q6: 你的 AC 自动机如何处理大小写变体和编码绕过？

**坦诚回答**：v0.1.0 已覆盖大小写变体——AC 自动机默认 `case_insensitive: true`，构建和匹配时均执行 `strings.ToLower` 折叠，`union select` 与 `UNION SELECT` 均可命中同一条规则。

- **大小写**：✅ v0.1.0 已覆盖——`case_insensitive` 默认启用，实现成本极低（构建+匹配各一次 `ToLower`）
- **Unicode/URL 编码**：❌ `%55%4E%49%4F%4E`、UTF-16 编码等属于数据规范化问题，v0.3.0 Roadmap 中将引入可插拔的规范化预处理层
- **SQL 注释插入**：❌ `UNION/**/SELECT` 需要正则预处理或 SQL 语法感知的规范化，这是更深层次的检测问题——当前设计不对攻击者构造输入的方式做假设

总结：我理解这些绕过技术的存在，并在 Roadmap 中为每一类都规划了应对方案。v0.1.0 的核心价值是建立端到端检测管道——规范化层是在此管道上的增量增强。

### Q7: 如果 C 进程 segfault 了，Go 端的恢复机制是什么？

**回答**：
1. **检测**：Go 端 UDS listener 检测到 EOF/连接断开 → 标记 capture_status = disconnected
2. **数据保全**：Go 端继续排空 channel 中的剩余包（正常处理，不丢弃），同时记录 `netsentry_capture_restarts_total` 递增
3. **自动恢复**：Go 端以子进程方式管理 C 进程（`os/exec`），检测到退出后自动重启（最多 3 次，防止 crash loop）
4. **数据丢失**：C 进程崩溃瞬间，UDS 发送缓冲区中未发送的帧丢失（这是双进程架构的固有代价，无法完全避免）。但已到达 Go 端的包不受影响
5. **根因定位**：C 端心跳帧包含 `session_id`，Go 端通过 `session_id` 变化区分"正常重启"和"异常崩溃"；C 端编译 ASan 版本用于崩溃后分析

**为什么不会崩溃**：设计上已在每个协议字段解引用前校验边界 + VLAN tag 处理 + fragment 跳过 + cJSON goto cleanup 模式 + Valgrind/ASan CI。但如果这些措施仍不足，上述恢复机制确保 Go 端不降级为孤儿进程。

---

## 十七、风险与应对

| 风险 | 影响 | 应对 | 优先级 |
|------|------|------|------|
| **C 指针无边界检查导致 segfault** | 畸形包击穿 C 进程 | 所有字段解析前校验长度；VLAN tag 循环检查；fragment 跳过；解析失败计数心跳上报 | P0 |
| **C 进程崩溃 → Go 端孤儿** | 检测管道停摆 | Go 端检测 UDS 断开 → 自动重启 C 进程（最多 3 次）→ `netsentry_capture_restarts_total` 告警 | P1 |
| **规则热加载竞态 → Worker panic** | 检测管道崩溃 | `sync.RWMutex` 保护 AC 自动机读写；CI `go test -race` | P0 |
| **Channel select/default 静默丢包** | 攻击包被丢弃，漏检 | 阻塞发送传播背压到 C 端；channel_depth Gauge 预警 | P0 |
| **告警刷屏导致告警疲劳** | sqlmap 5分钟15000条 | 告警聚合器 UPSERT + 60s 窗口去重 | P0 |
| **256字节检测深度漏掉攻击** | SQL注入在 POST body 500+字节处完全盲区 | 默认提升至 4096，规则级 offset/depth | P0 |
| **告警数据无限增长 → 磁盘耗尽** | 30天 ~75GB SQLite | 按天分库 + 7天 TTL 自动清理 + WAL 轮转 | P1 |
| **预写日志 crash 时截断** | 恢复丢数据 | fdatasync + event_id 幂等重放 + 截断行容错 | P1 |
| **Prometheus 一再推迟** | 运维不可见 | 回归 v0.1.0（<30行代码） | P0 |
| **"零拷贝"虚假宣传** | 面试时被追问无法自圆其说 | 诚实标注"协议解析层零拷贝"，列出完整拷贝路径 | P0 |
| **Worker 单协程 panic → 管道停摆** | nil pointer dereference 处理畸形 PacketInfo | Worker 循环内 `recover()` + 记录错误包；Go `-race` 编译 | P1 |
| **无二进制 Release + 多终端启动** | 用户 5 分钟内无法跑通 | make quickstart + Docker + CI Release 静态二进制 | P1 |
| **SQLite 查询与写入并发竞争** | API 被频繁轮询时写入阻塞 | WAL 模式读写不互斥；busy_timeout 5s 兜底；按天分库减小单库体积 | P1 |
| **磁盘空间耗尽 → SQLite 写入失败** | Worker 阻塞甚至 panic，告警数据丢失 | 三级降级机制（预警→紧急写入→只读模式）；`netsentry_db_dir_free_bytes` Gauge 监控；磁盘恢复后自动回写 | P0 |
| **误修改配置需重启生效** | 修改日志级别等非关键配置导致服务中断 | SIGHUP 热加载（动态参数立即生效，静态参数 WARN 提示重启） | P1 |

---

## 十八、术语表

| 术语 | 解释 |
|------|------|
| MITRE ATT&CK | 对手战术、技术和通用知识库 |
| Tactic | ATT&CK 战术（如 Initial Access、Discovery） |
| Technique | ATT&CK 技术（如 T1190、T1046） |
| 告警聚合 | 按 (rule_id, src_ip, dst_ip, 时间窗口) 合并重复告警 |
| UPSERT | SQLite INSERT ... ON CONFLICT ... DO UPDATE，原子性插入或更新 |
| 抑制规则 | 在特定条件下（IP 段、时间段）静默告警的规则 |
| 心跳帧 | C 端定期发送的状态消息，含丢包/解析错误计数及 session_id |
| 优雅退出 | 从信号捕获到资源释放的完整多步退出序列 |
| 预写日志 (WAL) | 先写 JSONL 日志再写 SQLite 的崩溃恢复机制 |
| session_id | C 端进程启动时生成的随机 UUID，用于区分 C 进程重启。区别于系统级 `boot_id`（`/proc/sys/kernel/random/boot_id`），`boot_id` 仅用于区分整机重启 |
| TOCTOU | Time-of-check to time-of-use，检查与使用之间的竞态窗口 |

---

## 十九、故障排查指南

### 问题 1：`make quickstart` 报 `libpcap not found`

```
$ make quickstart
/usr/bin/ld: cannot find -lpcap
```

**解决**：`sudo apt install libpcap-dev`（Ubuntu/Debian）或 `sudo yum install libpcap-devel`（RHEL/CentOS）。

### 问题 2：端口 8080 被占用

```
$ make quickstart
listen tcp :8080: bind: address already in use
```

**解决**：`lsof -i :8080` 查看占用进程，或修改 `config.yaml` 中 `api_port: 8081`。

### 问题 3：UDS socket 文件权限

```
$ make quickstart
connect: Permission denied (uds: /tmp/netsentry.sock)
```

**解决**：检查 `/tmp/netsentry.sock` 是否存在残留文件（上次未正常退出）→ `rm /tmp/netsentry.sock` 后重试。

### 问题 4：SQLite database locked

```
sqlite3: database is locked
```

**解决**：SQLite WAL 模式下读写不互斥，此错误通常因残留的 WAL 文件导致。删除 `data/netsentry.db-wal` 和 `data/netsentry.db-shm` 后重启。如果频繁出现，增大 `db_busy_timeout`。

### 问题 5：Go 模块编译报依赖下载失败

**解决**：设置 Go 模块代理 `GOPROXY=https://goproxy.cn,direct`（中国大陆）或 `GOPROXY=https://proxy.golang.org,direct`（海外）。

### 问题 6：C 模块 segfault（尽管有边界检查）

**解决**：
1. 使用 ASan 构建运行：`make build-asan && ./netsentry-capture-asan --pcap crash.pcap`
2. 检查 `dmesg | tail` 查看 segfault 地址
3. 提交 Issue，附上导致崩溃的 pcap 文件（如有敏感信息可用 `tcprewrite` 匿名化）

---

## 二十、系统设计面试扩展讨论

> 以下为面试中可能遇到的"单机 → 分布式"系统设计问题准备。

### 场景：将 NetSentry 扩展为监控 100 个节点的企业级 IDS

**核心架构决策**：

```
┌─────────────────────────────────────────────────────┐
│                   管理平面                            │
│  Web Dashboard + Alert Management + Rule Management │
└──────────────────────┬──────────────────────────────┘
                       │ gRPC
      ┌────────────────┼────────────────┐
      ▼                ▼                ▼
┌──────────┐    ┌──────────┐    ┌──────────┐
│ 探针 #1   │    │ 探针 #2   │    │ 探针 #N   │   ← C 模块独立部署
│ (Agent)  │    │ (Agent)  │    │ (Agent)  │
│ pcap 抓包 │    │ pcap 抓包 │    │ pcap 抓包 │
│ + 协议解析 │    │ + 协议解析 │    │ + 协议解析 │
└────┬─────┘    └────┬─────┘    └────┬─────┘
     │ gRPC/protobuf│               │
     ▼               ▼               ▼
┌─────────────────────────────────────────────────┐
│              Kafka / Redpanda                     │  ← 缓冲层
│          (解耦探测与检测，削峰填谷)                  │
└──────────────────────┬──────────────────────────┘
                       │
     ┌─────────────────┼─────────────────┐
     ▼                 ▼                 ▼
┌──────────┐    ┌──────────┐    ┌──────────┐
│ 检测 #1   │    │ 检测 #2   │    │ 检测 #N   │   ← Go 模块水平扩展
│ 规则引擎  │    │ 规则引擎  │    │ 规则引擎  │
│ + 聚合器  │    │ + 聚合器  │    │ + 聚合器  │
└────┬─────┘    └────┬─────┘    └────┬─────┘
     │                │               │
     └────────────────┼───────────────┘
                      ▼
┌─────────────────────────────────────────────────┐
│           Elasticsearch / ClickHouse              │  ← 存储层
│        (时序告警存储 + 全文搜索 + 聚合分析)          │
└──────────────────────┬──────────────────────────┘
                       │
                      ▼
               Grafana / Kibana                     ← 可视化
```

**每个决策的 tradeoff**：

| 决策 | 选型 | 替代方案 | 权衡 |
|------|------|---------|------|
| RPC 协议 | gRPC + protobuf | REST/HTTP、Thrift | protobuf 紧凑（比 JSON 少 3-5× 带宽），gRPC 内置流控和重连，但调试不如 REST 直观 |
| 消息队列 | Kafka | NATS、RabbitMQ、Redis Streams | Kafka 持久化和回溯能力最强，但运维复杂；NATS 更轻量，适合中小规模 |
| 存储 | ClickHouse | Elasticsearch、TimescaleDB | ClickHouse 压缩率极高（10:1）且查询快，但无全文搜索；ES 支持告警 payload 全文搜索但存储成本高 |
| 探针端检测 | 部分检测前移 | 全量转发的中心检测 | 探针端做 IP 黑名单匹配（几乎无开销），复杂规则由中心引擎处理——减少网络传输和中心压力 |
| 探针管理 | 中心配置推送 | Ansible/Agent 自更新 | gRPC 双向流实现实时规则推送和心跳监控，比定时拉取更实时 |

**在面试中说出这些 tradeoff，比给出"正确"答案更重要。**

---

## 二十一、附录 — 核心数据结构定义（Schema First）

> 在开发前明确 Schema，消除 C 端与 Go 端的歧义，减少联调成本。以下定义为 v0.1.0 的契约——两端实现必须严格遵循。

### 21.1 C → Go PacketInfo JSON 结构

C 端通过 UDS 逐行发送的 JSON 数据帧，每行一个完整的 JSON 对象：

```json
{
  "timestamp_sec": 1719300000,
  "timestamp_usec": 123456,
  "src_ip": "192.168.1.100",
  "dst_ip": "10.0.0.1",
  "src_port": 54321,
  "dst_port": 80,
  "protocol": 6,
  "tcp_flags": "0x18",
  "payload_len": 1400,
  "payload_preview": "R0VUIC8...",
  "is_fragment": false,
  "truncated": false
}
```

**字段定义与约束**：

| 字段 | 类型 | 必填 | 约束与说明 |
|------|------|------|-----------|
| `timestamp_sec` | int64 | ✅ | Unix timestamp（秒），整数类型避免 float 精度丢失（IEEE 754 double 在 1e9 量级精度仅 ~1μs，不够精确表达微秒） |
| `timestamp_usec` | int32 | ✅ | 微秒部分（0–999999），与 `timestamp_sec` 组合得到完整时间戳 |
| `src_ip` | string | ✅ | IPv4 点分十进制字符串（v0.1.0 仅 IPv4），Go 端用 `net.ParseIP` 校验 |
| `dst_ip` | string | ✅ | 同上 |
| `src_port` | uint16 | ✅ | 0–65535 |
| `dst_port` | uint16 | ✅ | 0–65535 |
| `protocol` | uint8 | ✅ | IP 协议号：`6` = TCP, `17` = UDP（IANA 分配值） |
| `tcp_flags` | string | ❌ | TCP flags HEX 字符串（如 `"0x18"` = ACK+PSH），仅 protocol=6 时有意义。**格式**：`"0x"` 前缀 + 两位 HEX 大写 |
| `payload_len` | uint16 | ✅ | 实际 payload 字节数（0–65535）。v0.1.0 实际最大值受 `payload_preview_len` 配置限制（默认 4096） |
| `payload_preview` | string | ❌ | **Base64 编码**的 payload 前 N 字节。选择 Base64 而非 Hex 的理由：3 字节 → 4 字符（Base64）vs 3 字节 → 6 字符（Hex），传输体积减少 33%。Go 端用 `encoding/base64.StdEncoding.DecodeString()` 解码 |
| `is_fragment` | bool | ✅ | IPv4 分片标志。为 `true` 时 Go 端仅记录 IP 层信息，不进行传输层深度检测 |
| `truncated` | bool | ✅ | payload 是否因缓冲区不足被截断。为 `true` 时 `payload_preview` 不完整 |

**设计决策记录**：
- **为什么时间戳分两段而非 float？** IEEE 754 double 在 Unix timestamp 秒级（~1e9）有效精度约 1μs，但 JSON 序列化 `1.7193e9` 这种科学计数法在不同语言间解析行为不一致。两段整数值消除了所有歧义。
- **为什么 payload_preview 用 Base64 而非 Hex？** 4096 字节 → Base64 约 5461 字符，Hex 约 8192 字符——Base64 节省 33% 传输和解析开销。
- **为什么 tcp_flags 用 string 而非 int？** HEX 字符串 `"0x18"` 在日志和调试中可读性优于裸 int `24`，且避免 JSON 数字解析的符号歧义。

### 21.2 Go 端 PacketInfo 结构体

```go
// pkg/model/packet.go

type PacketInfo struct {
    TimestampSec  int64  `json:"timestamp_sec"`
    TimestampUsec int32  `json:"timestamp_usec"`
    SrcIP         string `json:"src_ip"`
    DstIP         string `json:"dst_ip"`
    SrcPort       uint16 `json:"src_port"`
    DstPort       uint16 `json:"dst_port"`
    Protocol      uint8  `json:"protocol"`
    TCPFlags      string `json:"tcp_flags,omitempty"`
    PayloadLen    uint16 `json:"payload_len"`
    PayloadPreview string `json:"payload_preview,omitempty"`
    IsFragment    bool   `json:"is_fragment"`
    Truncated     bool   `json:"truncated"`
}

// Timestamp returns the full timestamp as time.Time.
func (p *PacketInfo) Timestamp() time.Time {
    return time.Unix(p.TimestampSec, int64(p.TimestampUsec)*1000)
}
```

### 21.3 Alert 数据库 Schema（SQLite）

v0.1.0 告警按天分库（`alerts_YYYY-MM-DD.db`），每天一个独立 SQLite 文件。以下是每个分库的建表语句：

> **时间戳格式一致性说明**：SQLite Schema 中 `first_seen`、`last_seen`、`window_start` 存储为 Unix timestamp 整数（`INTEGER`）。API handler 层在序列化告警时统一将整数 timestamp 转换为 ISO 8601 UTC 字符串（如 `"2026-06-25T10:30:00Z"`），确保 API 响应格式与文档示例一致。数据库侧整数存储的优势是范围查询和排序性能更高，API 侧 ISO 8601 字符串对调用方更友好。

```sql
-- alerts_YYYY-MM-DD.db 建表语句（每个按天分库独立执行）

CREATE TABLE IF NOT EXISTS alerts (
    id              TEXT PRIMARY KEY,                          -- UUID v4，告警唯一标识
    event_id        TEXT UNIQUE NOT NULL,                     -- 事件 ID（UUID v7），幂等重放去重的核心字段
    rule_id         TEXT NOT NULL,                            -- 触发规则 ID，关联 netsentry.db 中的 rules 表
    rule_name       TEXT NOT NULL DEFAULT '',                  -- 规则名称（冗余存储，避免 JOIN）
    src_ip          TEXT NOT NULL,                            -- 源 IP
    dst_ip          TEXT NOT NULL,                            -- 目标 IP
    dst_port        INTEGER NOT NULL DEFAULT 0,               -- 目标端口
    protocol        TEXT NOT NULL DEFAULT 'TCP',               -- 协议名（TCP/UDP/ICMP）
    severity        TEXT NOT NULL CHECK(severity IN ('low', 'medium', 'high', 'critical')),
    aggregated_count INTEGER NOT NULL DEFAULT 1,              -- 聚合窗口内重复次数
    first_seen      INTEGER NOT NULL,                         -- 首次发现时间（Unix timestamp 秒）
    last_seen       INTEGER NOT NULL,                         -- 最后发现时间（Unix timestamp 秒）
    window_start    INTEGER NOT NULL,                         -- 聚合窗口起始时间（截断到 60s 边界）
    mitre_tactic        TEXT NOT NULL DEFAULT '',              -- MITRE ATT&CK 战术（如 "Initial Access"）
    mitre_technique_id  TEXT NOT NULL DEFAULT '',              -- MITRE 技术 ID（如 "T1190"）
    mitre_technique_name TEXT NOT NULL DEFAULT '',             -- MITRE 技术名称
    payload_preview     TEXT NOT NULL DEFAULT '',              -- 匹配时 payload 的文本预览（明文截断）
    matched_keyword     TEXT NOT NULL DEFAULT '',              -- 命中的规则关键词
    raw_payload         TEXT NOT NULL DEFAULT '{}',            -- 完整上下文 JSON（供详情 API 查询）
    created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))  -- 记录创建时间
);

-- 核心聚合唯一索引（与 UPSERT 的 ON CONFLICT 子句精确对应）
-- 这是告警去重的基石——任何对该索引列的变更必须同步修改 UPSERT 语句
CREATE UNIQUE INDEX IF NOT EXISTS idx_alert_aggregation
ON alerts(rule_id, src_ip, dst_ip, dst_port, window_start);

-- 高频查询优化索引
CREATE INDEX IF NOT EXISTS idx_alert_last_seen ON alerts(last_seen DESC);
CREATE INDEX IF NOT EXISTS idx_alert_severity ON alerts(severity);
CREATE INDEX IF NOT EXISTS idx_alert_rule_id ON alerts(rule_id);
CREATE INDEX IF NOT EXISTS idx_alert_src_ip ON alerts(src_ip);
CREATE INDEX IF NOT EXISTS idx_alert_mitre_technique ON alerts(mitre_technique_id);

-- 时间范围查询加速（覆盖 /api/alerts 最常见的时间筛选模式）
CREATE INDEX IF NOT EXISTS idx_alert_first_seen ON alerts(first_seen);
```

**索引策略说明**：
- `idx_alert_aggregation`：UPSERT 的唯一性约束——所有聚合写操作依赖此索引。这是最重要的索引，缺失会导致重复告警。
- `idx_alert_last_seen DESC`：降序索引加速"最近告警"查询（API 默认排序）。
- `idx_alert_severity`：加速按严重级别筛选。
- `idx_alert_first_seen`：加速时间范围查询（`WHERE first_seen BETWEEN ? AND ?`）。
- v0.1.0 索引总数 6 个，写入开销 ≈ 20-30%——在接受范围内。v0.2.0 可按访问频率做冷热分离。

### 21.4 Rule 数据库 Schema（SQLite）

规则和抑制规则存储在 `netsentry.db`（不按天轮转）：

```sql
-- netsentry.db 建表语句

CREATE TABLE IF NOT EXISTS rules (
    id              TEXT PRIMARY KEY,                          -- 规则 ID，格式: rule-NNN
    name            TEXT NOT NULL,                             -- 规则名称
    type            TEXT NOT NULL CHECK(type IN ('ip_blacklist', 'payload_match', 'port_blacklist', 'frequency_threshold')),
    severity        TEXT NOT NULL CHECK(severity IN ('low', 'medium', 'high', 'critical')),
    priority        INTEGER NOT NULL DEFAULT 100,              -- 优先级（0–1000，越高越先匹配）
    enabled         INTEGER NOT NULL DEFAULT 1,                -- 是否启用（0/1）
    config_json     TEXT NOT NULL DEFAULT '{}',                -- 规则配置 JSON（keywords/ports/depth/offset/case_insensitive/protocols/direction）
    mitre_json      TEXT NOT NULL DEFAULT '[]',                 -- MITRE ATT&CK 映射 JSON 数组
    description     TEXT NOT NULL DEFAULT '',                  -- 规则描述
    created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    updated_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now'))
);

CREATE INDEX IF NOT EXISTS idx_rules_type ON rules(type);
CREATE INDEX IF NOT EXISTS idx_rules_enabled ON rules(enabled);
CREATE INDEX IF NOT EXISTS idx_rules_priority ON rules(priority DESC);

CREATE TABLE IF NOT EXISTS suppressions (
    id              TEXT PRIMARY KEY,
    rule_id         TEXT,                                      -- NULL = 抑制所有规则匹配该 IP
    src_ip          TEXT,                                      -- CIDR 或单个 IP
    dst_ip          TEXT,                                      -- CIDR 或单个 IP
    reason          TEXT NOT NULL DEFAULT '',
    expires_at      INTEGER,                                  -- Unix timestamp，NULL = 永不过期
    enabled         INTEGER NOT NULL DEFAULT 1,
    created_at      INTEGER NOT NULL DEFAULT (strftime('%s', 'now')),
    FOREIGN KEY (rule_id) REFERENCES rules(id) ON DELETE CASCADE
);

CREATE INDEX IF NOT EXISTS idx_suppressions_enabled ON suppressions(enabled);
CREATE INDEX IF NOT EXISTS idx_suppressions_expires ON suppressions(expires_at);
```

### 21.5 规则配置 JSON 子结构（`config_json` 字段）

`rules` 表的 `config_json` 字段存储类型特定的配置，Go 端在加载时反序列化为对应的结构体：

```go
// pkg/model/rule.go

// PayloadMatchConfig — payload_match 类型规则的配置
type PayloadMatchConfig struct {
    Keywords        []string `json:"keywords"`                   // 匹配关键词列表
    CaseInsensitive bool     `json:"case_insensitive"`           // 是否忽略大小写（v0.2.0）
    Protocols       []string `json:"protocols"`                  // 适用协议 ["TCP", "UDP"]
    Ports           []int    `json:"ports"`                      // 适用端口，空数组 = 所有端口
    Direction       string   `json:"direction"`                  // "src" | "dest" | "any"
    Depth           int      `json:"depth"`                      // 检测深度（字节），0 = 不限（实际受 payload_preview_len 限制）
    Offset          int      `json:"offset"`                     // 检测起始偏移（字节）
}

// IPBlacklistConfig — ip_blacklist 类型规则的配置
type IPBlacklistConfig struct {
    IPs         []string `json:"ips"`                           // IP 或 CIDR 列表
    Direction   string   `json:"direction"`                     // "src" | "dst" | "any"
}

// PortBlacklistConfig — port_blacklist 类型规则的配置
type PortBlacklistConfig struct {
    Ports       []int    `json:"ports"`                         // 黑名单端口列表
    Protocols   []string `json:"protocols"`                     // 适用协议
}

// FrequencyThresholdConfig — frequency_threshold 类型规则的配置
type FrequencyThresholdConfig struct {
    Threshold   int      `json:"threshold"`                     // 阈值（如 100 次/窗口）
    WindowSecs  int      `json:"window_secs"`                   // 统计窗口（秒）
    GroupBy     string   `json:"group_by"`                      // 聚合维度 "src_ip" | "dst_ip" | "dst_port"
}
```

### 21.6 审计日志 Schema（`netsentry.db`，v0.1.0 P0）

对于生产级安全工具，合规性要求必须详细记录"谁在什么时候做了什么"。v0.1.0 引入审计日志表，记录所有规则变更、抑制规则创建/删除等操作：

```sql
-- netsentry.db — audit_logs 表（记录所有非 GET API 操作）

CREATE TABLE IF NOT EXISTS audit_logs (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    timestamp       INTEGER NOT NULL,                              -- Unix timestamp 秒
    action          TEXT NOT NULL,                                  -- "CREATE_RULE", "DELETE_RULE", "UPDATE_RULE",
                                                                    -- "CREATE_SUPPRESSION", "DELETE_SUPPRESSION",
                                                                    -- "BATCH_DELETE_ALERTS", "BATCH_IMPORT_RULES"
    operator_ip     TEXT NOT NULL,                                  -- API 请求来源 IP
    user_identity   TEXT,                                           -- PSK 对应的标识（如配置了多个 token 时为 token 别名）
    resource_type   TEXT NOT NULL,                                  -- "rule", "suppression", "alert"
    resource_id     TEXT,                                           -- 受影响的资源 ID
    changes_json    TEXT,                                           -- 变更前后的 Diff JSON（CREATE 时为完整内容，UPDATE 时为 before/after）
    status          TEXT NOT NULL DEFAULT 'success',               -- "success" / "failed"
    error_message   TEXT                                            -- 仅 status='failed' 时出现
);

-- 按时间排序查询审计日志（最常用）
CREATE INDEX IF NOT EXISTS idx_audit_timestamp ON audit_logs(timestamp DESC);

-- 按操作类型筛选（如查找所有 DELETE_RULE 操作）
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_logs(action);

-- 按资源关联查询（如查找某个规则的所有变更历史）
CREATE INDEX IF NOT EXISTS idx_audit_resource ON audit_logs(resource_type, resource_id);
```

**Go 端实现要点**：

```go
// engine/internal/api/audit.go — 审计中间件
func AuditMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        // 仅拦截非 GET 请求
        if c.Request.Method == "GET" || c.Request.Method == "HEAD" {
            c.Next()
            return
        }
        
        // 记录操作前状态（用于 UPDATE/DELETE 场景的 before diff）
        beforeSnapshot := captureBeforeState(c)
        
        c.Next() // 执行实际 handler
        
        // 异步写入审计日志（不阻塞 API 响应）
        go func() {
            entry := &AuditEntry{
                Timestamp:    time.Now().Unix(),
                Action:       deriveAction(c.Request.Method, c.FullPath()),
                OperatorIP:   c.ClientIP(),
                UserIdentity: resolveTokenIdentity(c.GetHeader("Authorization")),
                ResourceType: extractResourceType(c.FullPath()),
                ResourceID:   c.Param("id"),
                ChangesJSON:  buildDiff(beforeSnapshot, c),
                Status:       deriveStatus(c.Writer.Status()),
            }
            auditStore.Write(entry) // 写入 audit_logs 表
        }()
    }
}
```

> **防篡改限制**：v0.1.0 审计日志存储在 SQLite 的 `audit_logs` 表中，理论上具有数据库写入权限的攻击者可修改/删除审计记录。v0.1.0 接受此风险，但应用层通过以下措施缓解：① 审计日志在 POST/PUT/DELETE handler 的 goroutine 中异步写入，攻击者难以在 API 响应返回前完成数据篡改；② 日志同时写入 stdout（JSON 格式），可被外部 syslog 采集。长期方案（v0.3.0 Roadmap）：审计日志同步发送到外部 syslog 服务器实现 append-only 保证。

**合规价值**：
- 企业安全审计（SOC 2、ISO 27001）要求保留所有配置变更的完整记录
- 安全事件调查中可追溯"谁在攻击发生前 5 分钟修改了规则"
- 多人协作时提供变更问责机制——防止误操作和恶意篡改

---

## 二十二、附录 — IPC 通信协议规范

### 22.1 UDS 连接握手协议（v0.1.0）

C 端连接成功后，立即发送握手帧。Go 端必须在收到握手帧并通过校验后才开始处理数据帧——未完成握手的连接上的数据帧将被静默丢弃。

**握手帧（C → Go）**：

```json
{
  "type": "hello",
  "version": "0.1.0",
  "session_id": "a1b2c3d4-e5f6-7890-abcd-ef1234567890",
  "pid": 12345,
  "hostname": "edge-sensor-01",
  "max_payload_len": 4096
}
```

| 字段 | 类型 | 说明 |
|------|------|------|
| `type` | string | 固定值 `"hello"` |
| `version` | string | C 端协议版本号（语义化版本） |
| `session_id` | string | C 端启动时生成的随机 UUID（进程级标识符），Go 端用于区分"断开重连"与"C 进程重启" |
| `pid` | int | C 进程 PID（用于日志关联和问题排查） |
| `hostname` | string | 主机名（多探针部署时区分来源） |
| `max_payload_len` | int | C 端实际配置的 payload 截断长度（Go 端可据此判断深度检测上限） |

**握手响应（Go → C，可选，v0.2.0 实现）**：

```json
{
  "type": "hello_ack",
  "version_ok": true,
  "server_version": "0.1.0",
  "session_id": "sess_a1b2c3d4"
}
```

**版本兼容性检查**：
- Go 端校验 C 端 `version`：主版本号必须完全一致（`0.x` vs `0.y`），次版本号向前兼容（Go `0.2.0` 接受 C `0.1.0` 和 `0.2.0`）
- 版本不兼容：Go 端记录 WARN 日志并主动 `Close()` 连接。C 端 `send()` 返回 `EPIPE` 后进入重连状态机（指数退避重试）
- v0.1.0 简化：Go 端不发送 `hello_ack`（仅记录握手日志）。v0.2.0 实现双向确认

### 22.2 二进制序列化路线图（v0.2.0）

**当前状态（v0.1.0）**：UDS + JSON 行式协议。简单可调试，但在高吞吐下 CPU 开销不容忽视。

**v0.2.0 优化方案**：引入 FlatBuffers 作为可选序列化格式：

| 方案 | 序列化速度 | 反序列化速度 | 消息体积 | 可调试性 | 适用场景 |
|------|-----------|-------------|---------|---------|---------|
| JSON（当前） | 基准 | 基准 | 基准 | ✅ 最佳 | 开发、低吞吐 |
| **FlatBuffers** | 5–10× 快 | **零拷贝**（直接在接收缓冲区上读取） | 比 JSON 小 30–50% | ❌ 需工具 | 高吞吐生产 |
| Protobuf | 3–5× 快 | 需完整反序列化到堆对象 | 比 JSON 小 60–70% | ❌ 需工具 | RPC 场景 |
| MessagePack | 2–4× 快 | 需完整反序列化 | 比 JSON 小 30–40% | ❌ 需工具 | 通用二进制 JSON 替代 |

**选择 FlatBuffers 而非 Protobuf 的理由**：
- FlatBuffers 支持 **零拷贝读取**——Go 端可以直接在接收缓冲区的 `[]byte` 上读取字段，无需二次分配堆内存。这对减少 GC 压力至关重要（50K PPS × 每次 JSON unmarshal 产生 ~200 临时对象 → 每秒 1000 万次 GC 扫描）
- C 端 `flatcc`（纯 C FlatBuffers 编译器）生成代码无堆分配，适合 C 模块的内存管理模式
- v0.2.0 实现方式：C 端同时编译 JSON 和 FlatBuffers 两个发送路径，通过编译宏 `-DUSE_FLATBUFFERS` 切换；Go 端自动检测消息格式（通过首字节魔数 0x4E53 = "NS" 区分 FlatBuffers/JSON）

**v0.1.0 折中方案**：保持 JSON，C 端使用 `cJSON_PrintBuffered` 将序列化结果直接写入 12KB 栈缓冲区（零堆分配），Go 端使用 `json.Decoder` 的 `DisallowUnknownFields()` 和预分配的 `sync.Pool` 复用 `PacketInfo` 结构体，减少 GC 压力。

---

**文档结束**

最后更新：2026-06-25  
发布版本：v0.1.0（规划中）
