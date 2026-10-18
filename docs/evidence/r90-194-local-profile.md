# R90-194 local-v1 tooling evidence (2026-10-18)

Local implementation and final quality gate passed. Exact post-push SHA, hosted
CI and Vault facts are deliberately retained in the local delivery record
`.git/netsentry-r90-194-delivery.json`, without an additional closure push.
No live local-v1 L1/L2 measurement or production acceptance was executed.

## Behavior and direct boundaries

Observation/manifest schema 2 explicitly binds profile and contract_id local-v1.
The evaluator keeps every oracle event and raw anomaly identity, enforces actual
per-minute load, independent sustained/burst loss and latency, clock validity,
submission outcomes, allocations and complete evidence declarations. Numeric
failures remain visible together with validity gaps. Explicit local review
recomputes the summary from exact input bytes before accepting a single supplied
run; its evidence hashes/checks are review declarations, not authentication.

Twelve directly executed methods in `scripts/test_slo_local.py` cover:

| Boundary | Direct regression |
| --- | --- |
| Frozen schema-1 staging/prod summaries and diagnostics | test_frozen_legacy_outputs_and_diagnostics; eight golden cases from the original baseline source |
| Contract/profile/version/member/policy rejection | test_contract_identity_and_exact_members |
| Duplicate oracle and unexpected/duplicate/late records | test_duplicate_oracle_and_all_anomaly_reasons |
| Inclusive 1% loss, sustained and each isolated burst | test_each_loss_segment_and_exact_one_percent |
| Missing infinity and numeric failures alongside invalid capture/clock | test_missing_and_numeric_failure_survive_invalid_measurement |
| Exact 500ms bound, nearest rank, bound review and stale/false review rejection | test_review_boundary_and_nearest_rank |
| Isolated-burst latency, inflated counts, duplicate offered identity and retained partial output | test_isolated_burst_latency_and_counter_preservation |
| +/-5% minute load, 600/30 durations, oracle seconds, clock, allocations and incomplete declarations | test_load_clock_resources_and_incomplete_boundaries |
| Real temporary SQLite indexing, failed submissions, missing/extra/duplicate/late IDs and byte-identical raw replay | test_adapter_retains_missing_extra_duplicate_late_and_failed_submissions |
| Schema-2 VM context/digest/identity binding, unknowns, cross-contract rejection and preserved legacy diagnostic precedence | test_context_contract_binding_and_unknowns |
| Current retained bundle/reconstruction/context pair, crossing local/legacy rejection and unchanged old-profile pairing | test_bundle_replay_and_pair_context_contracts |
| Actual reporter subprocess CLI, all exit categories, reviewed recomputation, source/output preservation and local default path | test_cli_status_review_and_input_preservation |

Integration ledgers deliberately offer only ten packets across ten minutes;
they establish derivation and binding while remaining inconclusive load evidence.
Their synthetic sender envelopes exercise bundle schema/receipt checks, not
sender frame/oracle replay. The separate direct evaluator fixture has 600 oracle
events and 108000 declared offered packets; it is synthetic supplied data.
None of these fixtures establishes measured throughput, clock truth or L1/L2.

## Final local gate

Preflight verified Go 1.26.9, archive SHA-256
`42d158b4d8f7b61ac0a830567c940a86098fb7aac52e467a5ebec03ef5cc2f8d`,
actionlint v1.7.12, govulncheck v1.6.0, module root `engine/`, Docker daemon,
Dockerfile syntax frontend and exact ubuntu:24.04 manifest metadata. No pin,
Dockerfile, runtime source or existing gate was weakened.

Final `make ci-check` exited 0, with required asset fetches and Docker enabled:
all nine locked assets matched bytes/hash; no reachable vulnerabilities found;
12 local SLO methods, 42 existing evidence and 33 knowledge tests passed;
uncached all-package race and coverage passed (82.3% total); ordinary C parser/
sender tests, both ASan fuzz gates, E2E (6 packets/5 alerts/8 rules), distribution
checksum/content and Docker image/content/runtime-health passed. Native C
outputs were cleaned before the final chain; sanitizer fuzz targets use their
separate outputs. No cached Go test success substitutes for execution.

Final raw full-gate log SHA-256:
`392e73212b7dc3a88cc5e9d77d6306af166bd8d71ea8f0890fd7c6582748ef69`.
Raw logs stay outside Git/Vault. Final code/test/fixture/Makefile hashes were
frozen and rechecked before delivery. Closeout-only documentation receives
fresh affected docs/JSON/roadmap/link/diff/sensitive and knowledge checks.

## Deviations and compatibility review

The first full chain stopped on a raw-GitHub TLS handshake timeout before the
third asset completed. That exact URL subsequently matched its locked bytes/hash;
no later check from the failed chain counts. Its failed-log SHA-256 is
`4e2f3ae7d966037b30a03d24b557e1718f0a52db6b6601daeb360d405e5d6222`.
The unchanged full retry passed; final review then found a legacy context
kind/profile diagnostic-order shift. Restore the old order, add a both-invalid
and profile-only direct control, rerun focused checks, and require the final
complete chain above. The superseded pass is retained, not used for final source.

Static comparison verified the original reporter validator and summarizer ASTs
are unchanged except the explicit schema-2 dispatch; all 201 row/Definition
multisets match without duplicates, historical completed Definitions are preserved,
and required forward dependencies/windows/risk/acceptance/check/stop fields exist.
The generic local skill now includes delegated evaluator-module source hashes in
provenance; that skill change is separate from this feature commit.

R90-195 is next only after exact delivery is verified. Its namespace/veth runner,
new submitted-row supply, cumulative artifact budgets, low-rate capture-buffer
measurement and live smoke have not started. L1/L2 and old production targets
remain unvalidated. Existing envelope review flags remain compatible metadata;
local report review does not wait for an external department.
