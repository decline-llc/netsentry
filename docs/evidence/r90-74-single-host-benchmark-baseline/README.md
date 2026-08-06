# R90-74 Repeated Single-Host Benchmark Baseline

## Evidence Boundary

- Status: `observation_only`
- Evidence class: `observation_only_local_synthetic_microbenchmark`
- Samples: 5 sequential uncached complete captures
- Commit: `b3d4f8f82e8913093be518ffe426f1d6dc8eee7f`
- Tree: `14aeceed857497b6c81a20b8cb06c9f187607a8f`
- Environment fingerprint: `da131adb45948ce4689a537121c495009182dcee985edf7c6ce24f1a9ccdddbb`
- Parameters: C iterations `100000`, Go benchtime `10s`
- Quartiles: inclusive method; variation: sample standard deviation / absolute mean.
- Threshold applied: false
- Production-derived or portable/cross-host claim: false
- Release, tag, or publication authority: false

## Samples

| File | SHA-256 | Start | End |
| --- | --- | --- | --- |
| `sample-01.json` | `cf2f1f1028cf3a25d94bac6e07c2c4ccb27d6a4c16ae6893a56f8df961f37258` | 2026-08-06T08:50:12.626279Z | 2026-08-06T08:52:43.316504Z |
| `sample-02.json` | `63d855ef5e12e94dfec67a64568cef17df3b69eba1c8a53fb1761e0592eac351` | 2026-08-06T08:52:43.834919Z | 2026-08-06T08:55:09.681060Z |
| `sample-03.json` | `e708542fe83dbc2f950d00fcdf3836d3ea527fd8c871128a0327cbed28022c30` | 2026-08-06T08:55:10.188193Z | 2026-08-06T08:57:39.408265Z |
| `sample-04.json` | `dbfe48519a8623773f150fd3c8741d8a69e0c30ce7fba1df463f45bda5bbc5f1` | 2026-08-06T08:57:39.940445Z | 2026-08-06T09:00:07.069458Z |
| `sample-05.json` | `245764decaf0c84455bb197f96e4b801c58b3674fa732ea5d7990f08ff022418` | 2026-08-06T09:00:07.643713Z | 2026-08-06T09:02:35.853678Z |

## Descriptive Results

### C Benchmarks

| Case | Metric | Median | Inclusive IQR | CV % | Min | Max |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| `bench_parser/tcp_plain` | `ns_per_packet` | 275.67 | 30.99 | 8.239373 | 265.81 | 320.98 |
| `bench_parser/tcp_plain` | `pps` | 3627550 | 384728 | 7.89701 | 3115412 | 3762130 |
| `bench_parser/tcp_qinq` | `ns_per_packet` | 278.15 | 5.19 | 7.84107 | 271.23 | 326.27 |
| `bench_parser/tcp_qinq` | `pps` | 3595129 | 66939 | 7.14775 | 3064984 | 3686877 |
| `bench_parser/tcp_vlan` | `ns_per_packet` | 268.62 | 5.48 | 9.902731 | 264.45 | 329.56 |
| `bench_parser/tcp_vlan` | `pps` | 3722691 | 75752 | 8.772971 | 3034304 | 3781456 |
| `bench_uds_sender/format_heartbeat_json` | `ns_per_op` | 21883.81 | 931.73 | 2.408047 | 21797.85 | 22849.99 |
| `bench_uds_sender/format_heartbeat_json` | `ops_per_sec` | 45696 | 1868 | 2.387247 | 43764 | 45876 |
| `bench_uds_sender/format_packet_json` | `ns_per_op` | 21584.39 | 798.58 | 3.60357 | 21509.81 | 23354.06 |
| `bench_uds_sender/format_packet_json` | `ops_per_sec` | 46330 | 1653 | 3.502192 | 42819 | 46490 |
| `bench_uds_sender/summary` | `avg_json_serialize_us` | 11.52 | 0.5 | 2.903127 | 11.51 | 12.23 |
| `bench_uds_sender/uds_send_line` | `ns_per_op` | 28042.1 | 859.97 | 3.134536 | 27712.95 | 29864.77 |
| `bench_uds_sender/uds_send_line` | `ops_per_sec` | 35661 | 1079 | 3.054113 | 33484 | 36084 |

### Go Benchmarks

| Case | Metric | Median | Inclusive IQR | CV % | Min | Max |
| --- | --- | ---: | ---: | ---: | ---: | ---: |
| `BenchmarkEngineMatch/multi_hit` | `B/op` | 3515 | 0 | 0 | 3515 | 3515 |
| `BenchmarkEngineMatch/multi_hit` | `MB/s` | 8.83 | 0.21 | 2.395926 | 8.4 | 8.9 |
| `BenchmarkEngineMatch/multi_hit` | `allocs/op` | 32 | 0 | 0 | 32 | 32 |
| `BenchmarkEngineMatch/multi_hit` | `ns/op` | 7247 | 174 | 2.446106 | 7187 | 7617 |
| `BenchmarkEngineMatch/no_hit` | `B/op` | 240 | 0 | 0 | 240 | 240 |
| `BenchmarkEngineMatch/no_hit` | `MB/s` | 78.97 | 0.57 | 0.627264 | 78.5 | 79.75 |
| `BenchmarkEngineMatch/no_hit` | `allocs/op` | 3 | 0 | 0 | 3 | 3 |
| `BenchmarkEngineMatch/no_hit` | `ns/op` | 823.1 | 5.9 | 0.625857 | 815 | 828 |
| `BenchmarkMatcherMatch/multi_hit` | `B/op` | 2104 | 0 | 0 | 2104 | 2104 |
| `BenchmarkMatcherMatch/multi_hit` | `MB/s` | 165.08 | 12.07 | 4.173737 | 156.7 | 171.67 |
| `BenchmarkMatcherMatch/multi_hit` | `allocs/op` | 5 | 0 | 0 | 5 | 5 |
| `BenchmarkMatcherMatch/multi_hit` | `ns/op` | 6118 | 459 | 4.192139 | 5883 | 6445 |
| `BenchmarkMatcherMatch/no_hit` | `B/op` | 1024 | 0 | 0 | 1024 | 1024 |
| `BenchmarkMatcherMatch/no_hit` | `MB/s` | 269.59 | 4.43 | 9.809213 | 214.77 | 275.51 |
| `BenchmarkMatcherMatch/no_hit` | `allocs/op` | 1 | 0 | 0 | 1 | 1 |
| `BenchmarkMatcherMatch/no_hit` | `ns/op` | 3561 | 59 | 11.252396 | 3485 | 4470 |
| `BenchmarkStoreQuery/exact_rule` | `B/op` | 26796 | 0 | 0 | 26796 | 26796 |
| `BenchmarkStoreQuery/exact_rule` | `allocs/op` | 1144 | 0 | 0 | 1144 | 1144 |
| `BenchmarkStoreQuery/exact_rule` | `ns/op` | 550888 | 7837 | 1.229917 | 539004 | 555674 |
| `BenchmarkStoreQuery/timestamp_range` | `B/op` | 29249 | 0 | 0 | 29249 | 29249 |
| `BenchmarkStoreQuery/timestamp_range` | `allocs/op` | 1157 | 0 | 0 | 1157 | 1157 |
| `BenchmarkStoreQuery/timestamp_range` | `ns/op` | 605793 | 16329 | 2.12352 | 598469 | 629831 |
| `BenchmarkStoreWriteBatch/batch_32_alerts` | `B/op` | 809847 | 47 | 0.007384 | 809810 | 809964 |
| `BenchmarkStoreWriteBatch/batch_32_alerts` | `alerts/op` | 32 | 0 | 0 | 32 | 32 |
| `BenchmarkStoreWriteBatch/batch_32_alerts` | `allocs/op` | 14900 | 0 | 0.003001 | 14900 | 14901 |
| `BenchmarkStoreWriteBatch/batch_32_alerts` | `ns/op` | 36548504 | 1136759 | 3.994116 | 35135154 | 39096913 |
| `BenchmarkStoreWriteBatch/single_alert` | `B/op` | 155251 | 9 | 0.004746 | 155242 | 155260 |
| `BenchmarkStoreWriteBatch/single_alert` | `alerts/op` | 1 | 0 | 0 | 1 | 1 |
| `BenchmarkStoreWriteBatch/single_alert` | `allocs/op` | 506 | 0 | 0 | 506 | 506 |
| `BenchmarkStoreWriteBatch/single_alert` | `ns/op` | 8088416 | 182053 | 2.799471 | 7811389 | 8421169 |

## Interpretation

These values describe five sequential observations from one exact clean commit and one unchanged local environment. Every raw sample is retained beside this summary and the aggregate is independently recomputable from their SHA-256-bound contents.

Variation is reported, not judged against a pass/fail budget. This baseline does not establish production capacity, cross-host portability, an SLO, a release gate, or publication authority.
