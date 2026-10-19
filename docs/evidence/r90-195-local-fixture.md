# R90-195 local fixture evidence

Evidence class: controlled same-VM synthetic **tooling smoke**, not L1/L2,
production capacity, physical NIC latency or power-loss durability. Baseline
fe4d2dbb22368325a6c636c8026e9c984a5e442f; R90-194 exact hosted success and Vault
were verified before selection. No pre-existing worktree edits were present.

## Implemented boundaries and direct regressions

New `scripts/local_fixture.py` and [runbook](../local-fixture.md) provide frozen
rules/oracle, rootless veth execution, cgroups, submitted outcomes and raw retention.
The 16 fixture methods and 12 existing local-profile methods passed together.
`make slo-check` includes both; full CI uses this same gate without live traffic.

| Acceptance boundary | Direct fixture regression/evidence |
| --- | --- |
| 100 rules; each positive/negative; 512-byte Ethernet/IP/UDP/checksums; 600 positives; exact 108000 schedule and minute/burst counts; deterministic source hashes | test_frozen_schedule_and_independent_packet_oracle; actual frozen oracle SHA checked before startup |
| No-send mode and exclusive destinations | test_dry_run_never_opens_socket_or_starts_process_and_preserves_existing |
| >=8GiB disk; cumulative 4GiB; exact boundary; private owner/links/root lock | test_budget_disk_links_and_serialized_root_ownership; actual periodic bytes plus fixed child-file envelope |
| Failed submission retains oracle denominator and actual error | test_submission_failure_retains_intent_offer_and_outcome |
| Early exit, readiness deadline and observed ready before send | test_readiness_early_failure_timeout_and_stop; actual API/UDS/capture and inbound Ethernet smoke |
| Cancellation before readiness, active cancellation, regular replacement/neighbor preservation | test_cancel_before_readiness_preserves_neighbor_and_replacement |
| Active SUT/generator reap, symlink/target preservation | test_active_cancel_reaps_all_children_and_preserves_symlink_target |
| Control send/ledger failure cannot skip child reap | test_wake_failure_still_reaps_owned_child |
| CLI cancellation forwarded to owned supervisor session | test_launcher_forwards_cancel_to_owned_session |
| Parent namespace and unowned mapping rejection | test_namespace_refuses_parent_or_unowned_map; actual fresh loopback-only socket inventory and different net inode |
| Actual child hard file cap and affinity readback | test_child_setup_enforces_file_cap_and_actual_affinity |
| Shared-clock interval, retained UTC metadata and sampled step rejection | test_clock_shared_interval_bound_and_steps; raw 100ms clock brackets/adjtimex/resolution |
| Exact lifecycle/persistence accounting; missing/extra/duplicate identities retained | test_accounting_retains_missing_extra_and_duplicate_identities; actual 20/2 smoke |
| Independent read-only SQLite observer; encoded path preservation | test_readonly_persistence_with_encoded_path |
| Schema-2 manifest acceptance and identity | test_manifest_matches_local_collector_contract |
| Budgeted copies/scratch, exact raw replay and drift refusal | test_budgeted_adapter_replays_supplied_raw_and_rejects_drift |

The adapter regression deliberately uses sparse synthetic supplied ledgers and
cannot establish full 108000-packet adapter capacity or acceptance. No full run
was sent. The actual smoke verifies cgroup max/quota/swap/affinity readback and
resource peaks; no delegated cgroup or privileged live test is required in CI.

## Retained measurements and deviations

Each smoke uses new workdir, UDS/API port, origin, namespace and state. The full
100-rule workload is frozen separately. Raw result inventories, intent/offered/
submission, lifecycle/capture close, clock/resource/API and SQLite files remain
local outside Git/Vault; the local delivery record locates them with hashes.

- Initial smoke: 20 packets reached the engine but idle capture exceeded its 5s
  graceful stop. Partial files, empty capture close and killed/incomplete result
  remain. No retry overwrote them.
- A separately recorded different-run control frame wakes idle libpcap after
  SIGTERM. It is noneligible traffic and does not increase the 20 offered count.
  Subsequent captures and engines both returned 0; namespace teardown and empty-
  directory-only cleanup preserve other objects.
- Audited smoke-04: exactly 20 offered/submitted, 40 arrival/processed rows, two
  oracle/durable/SQLite events, zero missing/extra/duplicate durable identities,
  API counts equal. Host arrival-to-WAL-return samples: 691919304/694217731ns.
  Shared interval candidate clock bound: 3014ns; raw absolute UTC state retained.
- Smoke-05 before the final clock guards: same exact 20/40/2 accounting; samples
  780693513/779954780ns, candidate clock bound 2784ns. Both exceed the 500ms L1
  target. These are 10pps smoke samples, not a frozen L1 verdict; retain the
  numbers and require R90-198 buffering repair before R90-196 measurement.
- Smoke-07 before the final wake-error cleanup repair: exact 20/40/2 accounting, 74102234/74352948ns delays,
  candidate clock bound 2263ns. This faster sample does not erase the prior
  ~616–781ms outcomes or establish L1; buffering phase variability remains the
  reason R90-198 precedes L1. All runs and the initial failure remain.
- Final-source smoke-08: exact 20/40/2 accounting; delays [238028947, 238359072]ns,
  candidate clock bound 3446ns. It retains the prior delayed samples.
- Final-source cgroups enforce 3GiB/four-CPU SUT and 1GiB/two-CPU generator,
  disjoint affinities, no swap; periodic resource/raw clock evidence is retained.
  Smoke-07 aggregate memory peaks: SUT 8253440 bytes, generator 13991936
  bytes. Per-file sizes and raw SHA inventory are in the local receipt.
- Rootless proc/1 access was denied and inherited sysfs exposed host interfaces;
  pre-isolation namespace identity plus socket enumeration fixed the inspection.
  A test mock initially intercepted provenance Git subprocesses; the correction
  now reaches the actual cancellation boundary and reaps the real owned child.

Clock bounds are candidate same-host interval evidence, not absolute UTC accuracy
or physical hardware latency. Sampling cannot exclude paired steps between probes;
L1 review must inspect operating conditions and all raw evidence. No reviewed
acceptance is generated automatically. No thresholds were relaxed.

## Delivery gate

Go 1.26.9 archive checksum matches the supply-chain lock; engine is the module
root. actionlint v1.7.12 and govulncheck v1.6.0 are present. Docker daemon 29.9.0,
exact Dockerfile frontend and ubuntu:24.04 manifest resolution passed before the
full chain. Focused 28 methods, Python/docs syntax, 33 knowledge methods and diff
review passed. Complete local ci-check and exact hosted/Vault facts are recorded
at final closeout; post-push facts remain local without a closure push.

The first complete ci-check passed, including Docker, but predates final clock
preflight/continuous rejection/monotonic-stop and hardware-headroom retention.
Its result is superseded for upload. Final focused/knowledge checks and a complete
clean ci-check rerun are required; no cached or mixed-source stage is substituted.

A later teardown review found that control-wake failure could skip waiting for
capture. The final repair always waits/kills/reaps the owned child, preserves the
wake cause and has a direct injected failure test. The second gate was deliberately
stopped during native tests after this source change; its termination is an
operator invalidation, not a reproduced storage defect. Only the next complete
final-source gate counts. No unrelated runtime code or assertion was changed.

Final-source gate deviation: coverage failed in existing
`TestRemoveOwnedSocketPreservesImmediateReplacementUnixListenerWithReusedInode`
at `uds_listener_test.go:1486`: after 10000 attempts the filesystem had not reused
the target inode, so the test's required setup was unavailable. Receiver/runtime
bytes were unchanged. Three exact uncached coverage-mode reruns all passed
(`go -C engine test -count=1 -covermode=atomic -run '^TestRemoveOwnedSocketPreservesImmediateReplacementUnixListenerWithReusedInode$' ./internal/receiver`).
No unrelated code or assertion was changed. Upload remains blocked until a clean
complete ci-check rerun; the failed full log and all three reruns remain local.

## Final complete local gate

The clean final-source rerun passed exit 0: nine fetched locked assets, no reachable
vulnerabilities, 28 SLO/fixture + 42 evidence + 33 knowledge methods, all-package
uncached race/coverage (82.3%), ordinary C and both ASan fuzz gates, E2E, archive
and unmodified full Docker build/content/health. Raw final log SHA-256:
`ead6a419231c2a379004cfd7f8156d02d30fc24abb36ca4c444b7362de91ebce`.
The previously failing receiver test passed in this full coverage suite; no
receiver or other detector/runtime source was changed. Earlier superseded,
intentionally terminated and failed gate logs and the three exact reruns remain.
Only closeout prose changed afterward; affected static/knowledge review is rerun
before the single batch commit/push. Exact hosted/Vault delivery facts remain
local; resolve the feature SHA and local delivery record before any repeat action.
