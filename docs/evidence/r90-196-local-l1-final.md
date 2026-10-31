# R90-196: two independently reviewed local-v1 L1 runs

SUT source baseline: `2bdff2d3bf2d6da80eca3bafc0ed9e9f910d5399`.
Local synthetic owned-veth traffic on the current shared Ubuntu VM. This is L1
only, not L2, real-data accuracy, physical-NIC capacity or power-loss durability.
Prior clock-invalid and metadata-rejected attempts retain their original status.

## Frozen workload and outcomes

Each run independently used 100 fixed payload rules, 512-byte Ethernet/IPv4/DF/
UDP frames, 600 seconds at 100pps with 500pps during seconds 180–239 and 480–539,
108000 planned packets, 600 expected events and 30 seconds observation. Fresh
namespace, API port, UDS, database, run ID and output; first fully reviewed before
second. Every actual minute is checked against the frozen +/-5% load envelope.
Both measured runs have zero missing, extra or duplicate security events.

| Run | Offered / processed | Durable / SQLite mapped | Loss steady, burst1, burst2 | Overall p99 upper ms | Worst cohort p99 upper ms | Clock bound ms |
| --- | --- | --- | --- | --- | --- | --- |
| r90-196-l1-05 | 108000 / 108000 | 600 / 600 | 0.000%, 0.000%, 0.000% | 8.260370 | 9.095959 | 0.007072 |
| r90-196-l1-06 | 108000 / 108000 | 600 / 600 | 0.000%, 0.000%, 0.000% | 8.037633 | 8.987318 | 0.016962 |

Each report independently satisfies segmented <=1% loss and p99 plus clock bound
<=500ms, with every expected event retained in its denominator. Raw maximums,
cohort details and all packet identities remain in the local independent reviews.
No successful-prefix or average-of-runs substitution is used.

## Validity, storage and retained proof

Independent review checks every actual intent/offer/submission, frozen schedule,
frame hash, 100-rule keyword matching and oracle identity. It recomputes all
cohort loss and nearest-rank latency from actual unique lifecycle records.
Independent encoded immutable/read-only SQLite observation returns integrity ok;
all 600 storage evt_ identities per run are rebuilt from exact rule/arrival
fields and mapped to measurement slo_ identities. API and C/Go close counts,
raw lifecycle hashes and normal exit statuses reconcile. Writer admission
verifies WAL/synchronous FULL; the immutable observer's connection-local delete
journal diagnostic is retained and does not replace the writer evidence.

Initial/active clock samples, brackets, resolution and kernel metadata are
retained before verdict. The exact shared-clock bound includes pcap 1us
quantization; 10ms is unchanged. Actual maximum inter-probe gaps are disclosed:

- r90-196-l1-05: 5693 probes, maximum gap 135.966599ms; cgroup memory peaks SUT/generator 39448576/62291968 bytes.
- r90-196-l1-06: 5702 probes, maximum gap 126.866533ms; cgroup memory peaks SUT/generator 39530496/65155072 bytes.

Sampling cannot exclude a step and inverse step wholly between probes. Cgroup
peaks include more than RSS and remain below 3GiB/1GiB; 4/2 disjoint logical CPU
affinities/quotas, no swap, memory-event counters and actual process readbacks
are checked. Shared VM resources do not establish exclusive production headroom.
Both processes stop normally without wake traffic after the observation window.

Actual producer clock/manifest passed a no-send derive/component/bundle/adapter
reconstruction preflight before transmission; this synthetic fixture emits no
review certificate and is excluded from measured outcomes. 60s stable-clock,
namespace-aware Ethernet/cgroup, pinned tools/module/Docker frontend/base and
ordinary clean-build preflights passed. Existing 34 fixture/local-SLO and 33
knowledge methods passed before acquisition. No heavy gate ran during measurement.

Every original is sealed and rechecked. Budgeted derive retains exact ledgers;
versioned bundle reconstruction freshly derives identical observations. Explicit
post-run legacy component projections check all null submission-error members
and capture extensions before omission, bind their originals/projections in a
hash proof and preserve original bytes. No legacy sender-frame replay is claimed;
independent review checks each local frame. Local raw context binds source/tool,
clock/resource/frozen facts and projection proof, not a production hardware
context. Seven explicit review checks bind exact observations, bundle, local
context and original raw manifest; each final status is reviewed_local_acceptance.

## Review deviation

The initial private record generator incorrectly required distinct namespace
inode numbers across the two sequential runs. Linux recycled the same numeric
inode after the first namespace closed. The failed generator and exact diagnostic
are retained. Correct the reviewer assumption: the versioned launcher constructs
a fresh unshare network namespace each invocation, verifies it differs from its
caller and owns only its interfaces. Each run has separately frozen state,
process IDs, config, port, UDS and files. Namespace number reuse after teardown
does not establish state reuse. No raw input, metric, threshold or SUT code was
changed; both reviewed reports already existed before this record-only correction.

## Artifact hashes and delivery boundary

### r90-196-l1-05

- raw-manifest.json: `35da7c95957df431eb56e73565229adc90c4c74dd4a33167f08f724e80fd1b86`.
- independent-review.json: `c17124e26f230cb408062d12aba1f530ee87f973a0e5a10827c15830d8135e06`.
- bundle/bundle.json: `763b1a5395a74f7d5c2081daf798e5778c4629ad60e346ec4ec5562a250efdb7`.
- local-context.json: `e7a82207bfe6ee93911f4526a207a4c4a4cfe86b550178a51d3125f3912b7e6e`.
- reviewed-report.json: `4c8d675f70ae503f028294f32c5d822a02bfb9c6f91aa4c6495c22dd9c828fb5`.

### r90-196-l1-06

- raw-manifest.json: `56f7d07d1eef69912269b5111710e2c37914f1accd0d9d5bf51729ba3c750a99`.
- independent-review.json: `6976924fb7bbd07618c086b544d2d994e95f5e2cac88942e8ce465a711ef48c9`.
- bundle/bundle.json: `51eaf59e6e38dd1327e4015191b7a0fbbd0503ef4f49c18bde89e48b9a5abf52`.
- local-context.json: `d8ada086104f19dfd71bf4c3cf83fed7549e03a5e5d9249ef63b214824668e2e`.
- reviewed-report.json: `900d27a48320f6bf98994a4123758273b7e674cb35e986390c9aca38c95b8acf`.

Complete make ci-check passed exit0: nine freshly fetched locked assets, zero reachable vulnerabilities,34 fixture/local-SLO/42 evidence/33 knowledge methods, uncached race/coverage 82.3%, C/ASan fuzz, E2E/archive and full Docker build/content/runtime health.
Gate log SHA-256: `37f5cf2e8cc3c438b80c6dba23251736c3091fb0ed62fa843e5c3d471c555f9a`.
Final static review verifies443 preserved historical/pending files, both original
raw manifests and exact review bindings,204 complete unique roadmap identifiers,
docs-only scope, JSON/links/fences and sensitive diff. Cumulative evidence is
about1.1GiB, below4GiB; exact bytes are kept in the local delivery record.
Only final prose follows, with affected docs/static/knowledge checks before commit.
All raw data, copies, scratch and logs stay under the private cumulative 4GiB
root; original historical artifacts and both older dirty worktrees are preserved.
No host clock service, threshold, oracle, runtime or evaluator changed. No L2,
tag, Release or registry publication is included. Complete upload only after the
full gate; exact fetched SHA hosted success and Vault are tracked locally, without
closure-only upload. R90-197 stability/recovery is the next increment after L1
verified delivery. Real traffic data is not required for this local-v1 evidence.
