# Departmental correctness validation handoff

This is the source-linked handoff for the Go correctness repairs delivered as
R90-144 through R90-172. All 29 implementations are delivered; agent behavioral,
race, full-suite and knowledge validation is **not run; delegated by user**.
Compile-only and static checks recorded in their plans do not establish runtime
behavior, byte preservation, durability, absence of races or SLO compliance.

The inventory contains 38 distinct changed Go test files and 106 distinct top-level
`Test` declarations currently in those files. One test file is shared
by two repairs, so the per-increment entries contain 39 file references and 112
function references. It includes existing functions in files amended by a repair; these are file-level regression companions, not 106
new tests or 106 executed cases. Subtests, table cases, benchmarks, other test
files and the earlier SLO/Python adapter work are outside this bounded inventory.
Use the linked plans for each repair's exact assertions and acceptance boundary.
The [SLO runbook](slo-runbook.md) retains the broader adapter/measurement handoff.

## Execution and evidence ownership

The specialist department owns execution and interpretation. This document does
not schedule runs, contact the department or authorize traffic, private inputs,
release publication or a profile acceptance claim. A delivered implementation
status and a successful compile do not clear its regression execution debt.

For each row, compare the persisted acceptance criteria with the actual test
body before execution. Confirm that public calls reach the promised boundary:

- Rule reload and control/suppression updates must observe the published state,
  matching and caller-input preservation. Join concurrent writers before
  comparing counters; synchronize on observable readiness for lifecycle tests.
- Packet completion must distinguish successful terminal processing from
  admission, failure and export boundaries. Nil entries and negative durations
  must reach the public metrics/snapshot interfaces named in their plans.
- Daily list/pagination tests must exceed the claimed truncation threshold and
  assert counts and complete contents; calendar checks must observe real files.
- Storage rejection must compare persistent bytes/modes/membership and retained
  rows through the independently encoded read-only observer before writable
  reopen. Accepted-mode controls must query the real live-connection PRAGMA.
- Redaction must assert exact output and downstream persistence, covering escaped
  values and preview truncation; legacy decoding must retain the earlier error.

A named nearby test is insufficient when it misses an acceptance boundary.
Record missing assertions or invalid fixtures as gaps; retain full failure logs
and fixture diagnostics. Keep synthetic/injected faults, real-store tests,
compile-only checks and production-derived evidence distinguishable.

## Commands for the department

Run these only under the department's execution authority. Start from a clean
checkout of the exact revision under review and retain `git rev-parse HEAD`,
`git status --short`, OS/architecture, exact tool versions, commands, exit status,
full logs, selector and any skipped native-width cases. The commands below are
instructions, not recorded successful runs.

Preflight the repository pins before the longer sequence. The engine module
currently requests Go 1.26.8; compare it with `engine/go.mod` and
`.github/supply-chain-lock.json` in the actual reviewed revision. Preflight C
compiler/libpcap prerequisites for `make test` and all scanner pins before any
additional release/security gate. Do not install unpinned replacements.

```bash
git rev-parse HEAD
git status --short
(cd engine && GOTOOLCHAIN=go1.26.8 go version)
```

Focused correctness covers the owning packages for every inventory entry,
including existing package companions. Run from the owning Go module; `-count=1`
disables cached test success. The package run uses no name filter, so log which
functions actually execute and any skips.

```bash
(cd engine && GOTOOLCHAIN=go1.26.8 go test -count=1 \
    ./internal/alert ./internal/api ./internal/config ./internal/pipeline ./internal/receiver ./internal/rule ./internal/rule/ahocorasick ./internal/stats)
```

After the focused run succeeds, use the established full C/Go race and knowledge
checks serially, stopping on the first failure. Record each result separately.
These correctness/knowledge commands do not execute formal SLO acceptance.

```bash
GOTOOLCHAIN=go1.26.8 make test && make knowledge-check
```

For repeated reliability evidence, retain `-count=1` on every focused rerun. If
an unrelated full-suite test fails, keep validation unresolved, retain its exact
name/log and rerun it uncached to assess reproducibility; require a clean full
suite before clearing that debt. Do not edit unrelated code from one ambiguous
failure. Do not run benchmarks concurrently with correctness/race tests.

A departmental result should identify the reviewed full SHA and the feature
SHAs below, list exact executed package/function/subtest boundaries and counts,
report every skip/failure/deviation, retain unfiltered logs and durable artifact
observations, and state which acceptance criteria remain unmet. Merge supplied
results into task evidence in a separately authorized increment; do not relabel
these historical agent records as executed passes.

## Source-linked inventory

Feature identifiers below are immutable delivery history. The department must
also record the actual checkout SHA because later fixes can affect execution.
Names are current source declarations at this handoff's creation, not historical
execution evidence. Each entry links the original plan and state for scope,
non-goals, diagnostics, fixture limitations and compile/static provenance.

### R90-144: isolate reload snapshots from caller mutations

Feature `87c6a92976d0e5f24782d765665b31ed74b89a3a` · [plan](plans/task-20261002-rule-snapshot-isolation.md) · [state](tasks/task-state-20261002-rule-snapshot-isolation.json).

[engine/internal/rule/engine_snapshot_test.go](../engine/internal/rule/engine_snapshot_test.go):

- `TestReloadSnapshotIsolatesCallerMutations`
- `TestReloadSnapshotPreservesInputAndDefensiveOutput`
- `TestReloadSnapshotRetainsStateOnRejection`
- `TestReloadSnapshotClearsNilAndEmptySets`
- `TestReloadSnapshotConcurrentPostReturnCallerMutation`

### R90-145: count successful terminal packet processing

Feature `1027b5a2ca7553e62c037c3e76c191ad140bf31f` · [plan](plans/task-20261002-packet-completion-counter.md) · [state](tasks/task-state-20261002-packet-completion-counter.json).

[engine/internal/api/completion_metrics_test.go](../engine/internal/api/completion_metrics_test.go):

- `TestCompletionMetricsEndpointPreservesHealthJSON`
- `TestCompletionMetricsEndpointWithNilStats`

[engine/internal/pipeline/completion_test.go](../engine/internal/pipeline/completion_test.go):

- `TestWorkerCompletionTerminalAndFailureBoundaries`
- `TestWorkerCompletionWaitsForProcessedObserverReturn`
- `TestWorkerCompletionSkipsNilAndPreCancelledEmptyInput`
- `TestWorkerCompletionCountsConcurrentWorkersAfterJoin`

[engine/internal/stats/completion_test.go](../engine/internal/stats/completion_test.go):

- `TestPacketCompletionCounterNilZeroAndIndependence`
- `TestPacketCompletionCounterConcurrentIncrements`

### R90-146: serialize suppression reload with mutations

Feature `5e6d9211a14a31622f589d0a8be18a6afd30768a` · [plan](plans/task-20261002-suppression-reload-serialization.md) · [state](tasks/task-state-20261002-suppression-reload-serialization.json).

[engine/internal/alert/suppression_reload_test.go](../engine/internal/alert/suppression_reload_test.go):

- `TestSuppressionReloadOwnsExclusiveLockAtAuthoritativeRead`
- `TestSuppressionReloadOverlapsMutationsWithoutLosingCommittedState`
- `TestSuppressionReloadErrorsPreserveStateReleaseLockAndAllowMutation`
- `TestSuppressionReloadGuardsAndMissingFile`
- `TestSuppressionReloadRetainsDefensiveListSlices`

### R90-147: preserve concurrent control state updates

Feature `8b8b7703f447391e2296813dc7b95023afd143e6` · [plan](plans/task-20261002-control-state-serialization.md) · [state](tasks/task-state-20261002-control-state-serialization.json).

[engine/internal/receiver/heartbeat_test.go](../engine/internal/receiver/heartbeat_test.go):

- `TestControlStateSequentialUpdatesPreserveOtherFrame`
- `TestControlStateConcurrentSettersRetainBothFramesAfterJoin`
- `TestControlStateConcurrentReadersObserveWholePublishedFrames`
- `TestReceiverConcurrentControlFramesRetainHelloAndHeartbeat`
- `TestControlStateSnapshotIsIndependentValue`

### R90-148: redact escaped JSON credential values completely

Feature `d5addcb22226404e2d6a43c3e0df800fde3a646b` · [plan](plans/task-20261003-json-value-redaction.md) · [state](tasks/task-state-20261003-json-value-redaction.json).

[engine/internal/alert/redactor_json_test.go](../engine/internal/alert/redactor_json_test.go):

- `TestRedactJSONCredentialValuesWithEscapes`
- `TestRedactJSONValueQuoteParityFormattingAndMultipleFields`
- `TestRedactEscapedJSONBatchPreservesAlertMetadataAndNilEntries`
- `TestJSONRedactionPreservesExistingHeaderPairAndUnrelatedValues`
- `TestRedactJSONCredentialValuesAtPreviewEnd`
- `TestRedactTruncatedJSONBatchPreservesMetadata`

[engine/internal/pipeline/redaction_json_test.go](../engine/internal/pipeline/redaction_json_test.go):

- `TestWorkerRealJSONRedactorRemovesEscapedSuffixBeforeWriter`

### R90-149: guard alert pagination integer overflow

Feature `3e792e53739b8ca9d1f4bfdcd1ce39f7e480ea08` · [plan](plans/task-20261003-pagination-overflow.md) · [state](tasks/task-state-20261003-pagination-overflow.json).

[engine/internal/api/pagination_overflow_test.go](../engine/internal/api/pagination_overflow_test.go):

- `TestParsePaginationOffsetBoundaries`
- `TestPageBoundsClampsBeforeAdding`
- `TestAlertsRejectOverflowBeforeStore`
- `TestAlertsPaginationKeepsRepresentableOffsets`

### R90-150: prevent daily query slice-bound overflow

Feature `da7b5cc8ecf33ce60ad4a7a28a63ea2ab5aa6a6a` · [plan](plans/task-20261003-shard-query-bounds.md) · [state](tasks/task-state-20261003-shard-query-bounds.json).

[engine/internal/alert/query_bounds_test.go](../engine/internal/alert/query_bounds_test.go):

- `TestSliceBoundsClampsLargeLimitBeforeAdding`
- `TestStoreQueryLargeLimitsAcrossPrimaryAndDailyShards`

### R90-151: honor uncapped limits in daily shard queries

Feature `e55c784f5dee3fe45b42b9e802608f5fc2a99760` · [plan](plans/task-20261003-shard-query-limit.md) · [state](tasks/task-state-20261003-shard-query-limit.json).

[engine/internal/alert/query_limit_test.go](../engine/internal/alert/query_limit_test.go):

- `TestStoreQueryNegativeLimitsDoNotTruncateAcrossModes`
- `TestStoreQueryLimitSemanticsForEmptyResults`

### R90-152: include historical shards in capped alert lists

Feature `492d6a285c2ff7669db68ec66ccf3623d5b6db05` · [plan](plans/task-20261003-shard-list.md) · [state](tasks/task-state-20261003-shard-list.json).

[engine/internal/alert/list_shards_test.go](../engine/internal/alert/list_shards_test.go):

- `TestStoreListAcrossModesKeepsGlobalOrderAndCap`
- `TestStoreListHistoricalRowsWithEmptyCurrentShard`
- `TestStoreListEmptyCanceledAndClosedAcrossModes`
- `TestStoreListRejectsCorruptHistoricalShard`

### R90-153: preserve committed response status in audit logs

Feature `6f443423ea3d5eb83ec4319acb84774eaba65241` · [plan](plans/task-20261003-audit-status.md) · [state](tasks/task-state-20261003-audit-status.json).

[engine/internal/api/audit_status_test.go](../engine/internal/api/audit_status_test.go):

- `TestAuditStatusMatchesCommittedHTTPResponse`
- `TestAuditResponseWriterHeaderForwardingAndSwitchingProtocols`
- `TestAuditResponseWriterDoesNotRecordRejectedHeader`

### R90-154: exclude nil entries from generated alert totals

Feature `dffafd3b02739d8d9b2b748c72255bf2939cf502` · [plan](plans/task-20261003-nil-alert-count.md) · [state](tasks/task-state-20261003-nil-alert-count.json).

[engine/internal/pipeline/nil_alert_metrics_test.go](../engine/internal/pipeline/nil_alert_metrics_test.go):

- `TestWorkerNilAlertMetricsMatchRealStoreRows`
- `TestWorkerNilAlertMetricsPreserveFailureAndExportGates`

[engine/internal/stats/nil_alerts_test.go](../engine/internal/stats/nil_alerts_test.go):

- `TestObserveAlertsCountsOnlyNonNilEntries`
- `TestObserveAlertsConcurrentNilBatchesAfterJoin`

### R90-155: reject negative duration observations

Feature `5c28c131c6432136e8453da169ffee7add9c4299` · [plan](plans/task-20261003-negative-durations.md) · [state](tasks/task-state-20261003-negative-durations.json).

[engine/internal/stats/negative_duration_test.go](../engine/internal/stats/negative_duration_test.go):

- `TestNegativeDurationObservationsPreserveSnapshotAndMetrics`
- `TestNonnegativeDurationObservationsRetainEveryBucketBoundary`
- `TestConcurrentDurationObservationsAfterWritersJoin`

### R90-156: count unpublished snapshots as empty

Feature `adcc5e792b371007e9cae8d124e499f76bf5c6c4` · [plan](plans/task-20261003-zero-rule-count.md) · [state](tasks/task-state-20261003-zero-rule-count.json).

[engine/internal/api/zero_rule_count_test.go](../engine/internal/api/zero_rule_count_test.go):

- `TestAPIZeroValueRuleEngineAcrossReloadLifecycle`

[engine/internal/rule/zero_count_test.go](../engine/internal/rule/zero_count_test.go):

- `TestZeroValueEngineCountAcrossReloadLifecycle`

### R90-157: isolate compiled matcher pattern snapshots

Feature `bb98065a2a232459c3e9cfd19787cda803b3654e` · [plan](plans/task-20261003-pattern-snapshots.md) · [state](tasks/task-state-20261003-pattern-snapshots.json).

[engine/internal/rule/ahocorasick/pattern_snapshot_test.go](../engine/internal/rule/ahocorasick/pattern_snapshot_test.go):

- `TestPatternsSnapshotsAreIndependentlyOwned`
- `TestConcurrentPatternSnapshotsRemainIndependent`

[engine/internal/rule/candidate_set_test.go](../engine/internal/rule/candidate_set_test.go):

- `TestPayloadCandidateSetPreservesPerRuleAlertSelection`

### R90-158: preserve invalid calendar shard files

Feature `f78e2ce1386d4f3c20c68bfd9eb5e0ea5a32941e` · [plan](plans/task-20261003-shard-calendar.md) · [state](tasks/task-state-20261003-shard-calendar.json).

[engine/internal/alert/shard_calendar_test.go](../engine/internal/alert/shard_calendar_test.go):

- `TestShardCleanupPreservesInvalidCalendarFiles`
- `TestShardCalendarCleanupDisabledAndCanceledPreserveFiles`

### R90-159: exclude invalid calendar names from shard discovery

Feature `48c26dc4b67b175b0be8ed06716c347c0411ec80` · [plan](plans/task-20261003-shard-discovery-calendar.md) · [state](tasks/task-state-20261003-shard-discovery-calendar.json).

[engine/internal/alert/shard_discovery_calendar_test.go](../engine/internal/alert/shard_discovery_calendar_test.go):

- `TestDailyShardReadsIgnoreInvalidCalendarsWithoutModification`
- `TestDailyShardCalendarDiscoveryStillRejectsValidDateCorruption`

### R90-160: reject overflowing duration settings

Feature `0608c1dfdf95be119927d666e5fa131182e5c6b2` · [plan](plans/task-20261003-duration-config.md) · [state](tasks/task-state-20261003-duration-config.json).

[engine/internal/config/duration_bounds_test.go](../engine/internal/config/duration_bounds_test.go):

- `TestLoadDurationSettingsPreserveRepresentableValuesAndRejectOverflow`
- `TestLoadDurationDefaultsAndCombinedDiagnostics`
- `TestLoadExpandedDurationSettingsUseTheSameBounds`

### R90-161: initialize zero-value alert severity counters

Feature `b233859ef0fc48a5c973e7d5d3963716d993a326` · [plan](plans/task-20261003-zero-stats-alerts.md) · [state](tasks/task-state-20261003-zero-stats-alerts.json).

[engine/internal/pipeline/zero_stats_test.go](../engine/internal/pipeline/zero_stats_test.go):

- `TestWorkerZeroStatsCompletesAfterRealStoreWrite`
- `TestWorkerZeroStatsPreservesNoAlertAndFailedWriteGates`

[engine/internal/stats/zero_alerts_test.go](../engine/internal/stats/zero_alerts_test.go):

- `TestZeroAndConstructedStatsObserveAlertsPreserveSnapshotAndInputs`
- `TestZeroStatsConcurrentFirstAlertObservationAfterJoin`

### R90-162: reject IP blacklists without compiled addresses

Feature `c6a22b1c3e59ea452dbbc34c4e45c55dba9ba3a3` · [plan](plans/task-20261003-empty-ip-blacklist.md) · [state](tasks/task-state-20261003-empty-ip-blacklist.json).

[engine/internal/rule/empty_ip_blacklist_test.go](../engine/internal/rule/empty_ip_blacklist_test.go):

- `TestEmptyIPBlacklistReloadRejectsAndPreservesSnapshot`
- `TestNonemptyIPBlacklistRetainsMatchingFiltersAndOwnership`
- `TestLoadedIPBlacklistValidationPreservesFileAndPublishedState`

### R90-163: reject suppressions without compiled prefixes

Feature `6f3781a36b5997f2c11a35e5d7f2ee379627c080` · [plan](plans/task-20261003-empty-suppressions.md) · [state](tasks/task-state-20261003-empty-suppressions.json).

[engine/internal/alert/empty_suppressions_test.go](../engine/internal/alert/empty_suppressions_test.go):

- `TestSuppressionConstructorsRejectNoCompiledPrefixesWithoutChangingInputs`
- `TestSuppressionConstructorsPreserveEmptySetsAndDisabledPrefixSkipping`
- `TestSuppressionMixedEmptyAndValidPrefixesRetainMatchingAndScoping`
- `TestSuppressionInvalidPrefixesKeepPriorDiagnosticOrder`
- `TestFileBackedSuppressionMutationsRejectEmptyCompiledPrefixesAndPermitRetry`

### R90-164: redact JSON credentials truncated at preview end

Feature `e4563171ae8795a226f202c0187d49ddd69d06a6` · [plan](plans/task-20261003-truncated-json-redaction.md) · [state](tasks/task-state-20261003-truncated-json-redaction.json).

[engine/internal/alert/redactor_json_test.go](../engine/internal/alert/redactor_json_test.go):

- `TestRedactJSONCredentialValuesWithEscapes`
- `TestRedactJSONValueQuoteParityFormattingAndMultipleFields`
- `TestRedactEscapedJSONBatchPreservesAlertMetadataAndNilEntries`
- `TestJSONRedactionPreservesExistingHeaderPairAndUnrelatedValues`
- `TestRedactJSONCredentialValuesAtPreviewEnd`
- `TestRedactTruncatedJSONBatchPreservesMetadata`

[engine/internal/pipeline/redaction_truncation_test.go](../engine/internal/pipeline/redaction_truncation_test.go):

- `TestWorkerRedactsActualEngineTruncatedJSONBeforeWriter`

### R90-165: preserve legacy field decode failures

Feature `0930bc7bd25c394a6bac72e1afedf347b9956d57` · [plan](plans/task-20261003-legacy-rule-decode.md) · [state](tasks/task-state-20261003-legacy-rule-decode.json).

[engine/internal/api/rule_reload_decode_test.go](../engine/internal/api/rule_reload_decode_test.go):

- `TestHTTPRuleReloadPreservesLegacyDecodeFailureAndAllowsRepair`

[engine/internal/rule/loader_legacy_decode_test.go](../engine/internal/rule/loader_legacy_decode_test.go):

- `TestLoadRulesRetainsMalformedLegacyWrappedDecodeError`
- `TestLoadRulesPreservesValidLegacyAndCanonicalFormats`
- `TestLoadRulesKeepsExistingEmptyNullAndSyntaxBoundaries`

### R90-166: reject canceled context before store startup

Feature `2dcd6d46bb38d5807ff41dff94a2e9086aae4f34` · [plan](plans/task-20261003-store-open-cancellation.md) · [state](tasks/task-state-20261003-store-open-cancellation.json).

[engine/internal/alert/store_open_cancellation_test.go](../engine/internal/alert/store_open_cancellation_test.go):

- `TestOpenAlreadyDoneContextPreservesInputs`
- `TestOpenAlreadyDoneContextPrecedesDurableOptionValidation`
- `TestOpenLiveContextCreatesReadableStore`

### R90-167: preserve question marks in SQLite filenames

Feature `029af21905d9b4fd6cd4d8a86495dc1ed1d2a528` · [plan](plans/task-20261003-store-path-encoding.md) · [state](tasks/task-state-20261003-store-path-encoding.json).

[engine/internal/alert/store_path_encoding_test.go](../engine/internal/alert/store_path_encoding_test.go):

- `TestStoreWritableQuestionMarkPrimaryPath`
- `TestStoreWritableQuestionMarkDailyShardPaths`
- `TestStoreWritableQuestionMarkRejectedRecoveryPreservesTree`

### R90-168: preserve relative SQLite filesystem paths

Feature `35ff8c067acfd58be17697952ab3e11c3bc4dfa7` · [plan](plans/task-20261003-store-file-prefix.md) · [state](tasks/task-state-20261003-store-file-prefix.json).

[engine/internal/alert/store_file_prefix_test.go](../engine/internal/alert/store_file_prefix_test.go):

- `TestStoreLiteralFilePrefixPrimaryPathsPreserveAlternateTarget`
- `TestStoreLiteralFilePrefixDailyShardPaths`
- `TestStoreLiteralFilePrefixRejectedInputsPreserveTree`

### R90-169: persist literal memory-sentinel filenames

Feature `5df8af0637bb64dbc1de9d12ef29e9fa9822b819` · [plan](plans/task-20261003-store-memory-filename.md) · [state](tasks/task-state-20261003-store-memory-filename.json).

[engine/internal/alert/store_memory_filename_test.go](../engine/internal/alert/store_memory_filename_test.go):

- `TestStoreLiteralMemoryFilenamePersistsAcrossReopen`
- `TestStoreLiteralMemoryFilenameUsesExistingFile`
- `TestStoreLiteralMemoryFilenameRejectedInputsPreserveTree`

### R90-170: match equivalent exact IP blacklist addresses

Feature `c7e6a56b67f724ebc7a36252df9f0c790780c9a4` · [plan](plans/task-20261003-ip-identity.md) · [state](tasks/task-state-20261003-ip-identity.json).

[engine/internal/rule/ip_identity_test.go](../engine/internal/rule/ip_identity_test.go):

- `TestExactIPBlacklistMatchesEquivalentAddressSpellings`
- `TestExactIPIdentityRetainsFiltersAndCIDRPrecedence`
- `TestExactIPIdentityRetainsRuleScopeOrderingAndRejectedReload`
- `TestExactIPIdentityFileRoundTripPreservesRuleSpelling`

### R90-171: reject overflowing SQLite busy timeouts

Feature `3a11443ca30e426699ec0e85de08b6b0bb4e5bb8` · [plan](plans/task-20261003-busy-timeout-bounds.md) · [state](tasks/task-state-20261003-busy-timeout-bounds.json).

[engine/internal/alert/store_busy_timeout_bounds_test.go](../engine/internal/alert/store_busy_timeout_bounds_test.go):

- `TestOpenBusyTimeoutOverflowPreservesPersistentInputs`
- `TestOpenBusyTimeoutBoundsPreserveValidationPrecedence`
- `TestOpenBusyTimeoutRepresentableValuesRetainEffectivePragma`

[engine/internal/config/busy_timeout_bounds_test.go](../engine/internal/config/busy_timeout_bounds_test.go):

- `TestLoadBusyTimeoutRepresentabilityAndDefaults`
- `TestLoadExpandedBusyTimeoutUsesTheSameBoundary`
- `TestLoadBusyTimeoutDiagnosticOrdering`

### R90-172: validate journal modes before startup side effects

Feature `a217c1d1db05b573cdc7656bf3f6a016f5b8e46b` · [plan](plans/task-20261003-journal-mode-preflight.md) · [state](tasks/task-state-20261003-journal-mode-preflight.json).

[engine/internal/alert/store_journal_preflight_test.go](../engine/internal/alert/store_journal_preflight_test.go):

- `TestOpenInvalidJournalModePreservesPersistentInputs`
- `TestOpenJournalPreflightPreservesEarlierDiagnostics`
- `TestOpenSupportedJournalModesRetainPragmaAndReopen`

## Independent SLO acceptance remains outstanding

R90-75 still requires the full [formal acceptance contract](performance-slo.md)
and [departmental runbook](slo-runbook.md): isolated declared run context, actual
profile resources, correlated live-arrival-to-durable latency, offered versus
completed loss, missing expected alerts retained as failures, extended-tail
coverage, full raw artifacts and reviewed results. These native correctness
regressions cannot establish that outcome. No qualifying profile measurements
are supplied with this handoff.


## Follow-up: R90-174 severity label encoding

The original R90-144..172 inventory and its counts above remain historical.
This separate follow-up adds two source files and three declarations:

- [Renderer regressions](../engine/internal/stats/label_encoding_test.go):
  `TestPrometheusSeverityLabelEncoding` and
  `TestPrometheusSeverityLabelsPreserveSortAndCounts`. Review twelve exact-byte
  cases (canonical, escape, control, Unicode and mixed-injection forms), a reader
  accepting only the three legal Prometheus escapes, raw identity/counts/input/
  snapshot preservation, and sorting across colliding-looking escaped values.
- [HTTP regression](../engine/internal/api/metrics_label_encoding_test.go):
  `TestMetricsSeverityLabelEncodingPreservesHealthIdentity`. Real Stats and
  HTTP Handler with existing store/queue/rule fixtures; exact mixed label bytes,
  canonical lines, total, 200/content type, repeated label lines and raw label
  identity plus existing health JSON field set. This is not a socket/network or
  production Prometheus scraper test.
- [Plan](plans/task-20261004-prometheus-labels.md) and
  [state](tasks/task-state-20261004-prometheus-labels.json) preserve R90-174 scope,
  authority, validation and exact delivery evidence. Resolve its feature SHA
  from those records and Git after delivery.

The package commands above already include stats and API. Execution remains
**not run; delegated by user**. Compile-only success does not establish parser,
HTTP, race or SLO acceptance. Verify the actual returned lines and health JSON
against the plan before departmental execution; invalid UTF-8 policy is outside
scope. Independent R90-75 requirements remain unchanged.

## Follow-up: R90-175 explicit suppression rule scope

This follow-up preserves the original R90-144..172 inventory and R90-174 record.
Five new declarations in two files require departmental execution:

- [Suppressor and manager regressions](../engine/internal/alert/suppression_rule_scope_test.go):
  `TestSuppressionConstructorsRejectExplicitEmptyRuleScope`,
  `TestSuppressionRuleScopePreservesAcceptedSemantics`,
  `TestSuppressionEmptyRuleScopePreservesPrefixDiagnosticPrecedence`, and
  `TestFileBackedSuppressionRuleScopeRejectionPreservesStateAndPermitsRetry`.
  Check one/multiple empty IDs, source/destination/any ranges, no partial result,
  nil/empty all-rule scopes, mixed/duplicate/exact whitespace IDs and disabled
  skip. Real Add/Update/Reload must preserve file bytes/mode/directory membership,
  published List/Filter and caller inputs, then accept a valid same-operation
  retry. Raw structural load acceptance differs from compilation acceptance.
- [HTTP regression](../engine/internal/api/suppression_rule_scope_test.go):
  `TestSuppressionHTTPRejectsExplicitEmptyRuleScopeAndPermitsRetry` reaches
  Handler with a real file-backed manager. POST/PUT must return 400
  VALIDATION_ERROR; reload retains 500 INTERNAL_ERROR. Exact details/request ID,
  file/list/filter preservation and valid retry responses/persistence are asserted.
- [Plan](plans/task-20261004-suppression-rule-scope.md) and
  [state](tasks/task-state-20261004-suppression-rule-scope.json) retain scope,
  authority, validation and exact delivery evidence. Resolve the delivered SHA
  through those records and Git.

Existing owning-package commands above include alert/API. Execution remains
**not run; delegated by user**; compilation is no runtime, HTTP, race or SLO
acceptance result. Prefix diagnostics retain precedence; whitespace IDs are
literal and are not newly rejected or normalized. R90-75 remains independent.

## Follow-up: R90-176 explicit protocol wildcard

The historical inventory and previous follow-ups remain unchanged. Four new
declarations in two files require departmental execution:

- [Engine regressions](../engine/internal/rule/protocol_wildcard_test.go):
  `TestExplicitAnyProtocolPreservesWildcardAcrossRuleTypes`,
  `TestAnyProtocolPreservesOtherRuleGates`, and
  `TestAnyProtocolRejectsInvalidEntriesAndPreservesSnapshot`. Review payload,
  IP and port rules; wildcard order/case/space/duplicates; protocols 0/1/6/17/255;
  complete alerts and input/config preservation. Check nil/empty/blank/named
  compatibility, other filter gates and disabled rules. Invalid entries before,
  after and between wildcards must retain exact errors and prior snapshots,
  including disabled candidates, followed by valid retry. Earlier direction
  and payload-window diagnostics retain precedence.
- [HTTP regression](../engine/internal/api/protocol_wildcard_test.go):
  `TestRuleHTTPProtocolWildcardPersistsRejectsAndPermitsRetry` uses a real
  Engine and seed file for POST/PUT/reload across all three types. Invalid
  entries on either side of `any` retain 400 VALIDATION_ERROR, operation-specific
  message/details/request ID and file bytes/mode/membership plus matching/count/
  rules. Valid same-operation retry checks status/response, canonical loaded
  persistence and original protocol list, every protocol and alert contents.
  Reload preserves the operator-supplied file.
- [Plan](plans/task-20261004-protocol-wildcard.md) and
  [state](tasks/task-state-20261004-protocol-wildcard.json) retain scope, authority,
  compile/static evidence and exact delivery history. Resolve the feature SHA
  through those records and Git.

The existing owning-package commands include rule/API. Execution is **not run;
delegated by user**. Compile-only success does not establish runtime, HTTP,
race or SLO acceptance. Explicit `any` broadens mixed lists as requested by the
wildcard; blank entries continue to be ignored. R90-75 remains independent.

## R90-177: active daily-shard pathname aliases

This supplement leaves the frozen R90-144..172 inventory unchanged.
[Plan](plans/task-20261004-shard-path-alias.md) ·
[state](tasks/task-state-20261004-shard-path-alias.json).

[engine/internal/alert/shard_path_alias_test.go](../engine/internal/alert/shard_path_alias_test.go):

- `TestDailyShardLexicalActivePathAliasesCountRowsOnce`
- `TestDailyShardAliasPreservesMissingDirectoryAndHistoricalErrors`

[engine/internal/api/shard_path_alias_test.go](../engine/internal/api/shard_path_alias_test.go):

- `TestDailyShardAliasHTTPListHealthAndMetricsUseActualCounts`

Review the actual public Options.Path/Dir spellings: relative directory with
absolute active file, absolute directory with relative active file, two relative
spellings with dot components, absolute dot spelling and identical control.
Each main active-store fixture uses WAL and DELETE and seeds two current rows
plus one DELETE historical row through separate non-daily stores. Read-only observers use
independently encoded absolute paths and open before query/count/list calls;
reuse them to assert real row counts. Compare complete rows, totals, filters,
ordering, page offsets, original Path, caller inputs and full persistent
bytes/modes/membership. Preserve missing-directory fallback, invalid-calendar
skip and actual corrupt historical errors. The HTTP test uses the real daily
store and checks list pages, ordinary/verbose health totals and the current
alert gauge across repeated exports, retaining file bytes/mode/membership.

Three declarations in two files are authored and compile-reviewed; behavioral,
race, full-suite, scanner, knowledge, traffic and acceptance checks are
**not run; delegated by user**. Compile success does not establish read-only,
HTTP, durability or race behavior. Execute owning alert/API package correctness
and the established serial full/knowledge sequence under the department's
existing authority; preserve exact revision, commands, logs, skips and failures.
No new SLO profile acceptance or publication authority is implied.


## R90-178: preserve the active database during file retention

This supplement preserves the original frozen inventory and prior follow-ups.
[Plan](plans/task-20261004-shard-active-retention.md) ·
[state](tasks/task-state-20261004-shard-active-retention.json).

[engine/internal/alert/shard_active_retention_test.go](../engine/internal/alert/shard_active_retention_test.go):

- `TestExpiredActiveShardSurvivesStartupAndLexicalCleanup`
- `TestLongLivedActiveShardCleanupPreservesFileAndSeparateRowTTL`
- `TestActiveShardCleanupGuardsAndOtherDirectory`

Review twenty public startup/direct fixtures spanning WAL/DELETE, primary/daily
and five identical/relative/absolute/dot pathname spellings. Primary mode passes
Options.Path directly. Daily mode ignores Path: the first public Now call chooses
Sep 1, later clock calls return Oct 4 before startup retention; expected Path is
derived from Dir/Sep 1 and direct cleanup supplies its own directory alias.
Reopen resets this initial path clock. The expired active filename contains a
current durable alert. Startup asserts base inode/mode/Path
and all alerts/events columns through an independently encoded read-only handle;
ordinary expired controls disappear while cutoff and invalid-calendar controls
retain bytes/modes. Writable initialization does not promise byte preservation.
Direct cleanup observers are established before removal and reused; actual WAL
sidecars must exist. Verify exact active bytes/modes/artifact identities and tree
membership, complete public List/Query/Count, full independent durable rows,
three-artifact ordinary deletion then repeated zero, original input/Path/health,
continued primary/daily routing and close/reopen plus fresh observer counts/rows.
The independent raw-column reader and existing public fixture/tree/encoded
observer helpers are declared in the owning alert test package.

Clock advancement expires a generated initially current pathname; file cleanup
must preserve it, while a separate public PruneExpired still removes its expired
row. Missing-directory, disabled and pre-canceled calls preserve artifacts; an
identical basename in another cleanup directory remains removable. No
symlink/hardlink identity or concurrent independent-store ownership is promised.

Three declarations are authored and compile-reviewed. Behavioral/race/full/
scanner/knowledge/traffic/acceptance **not run; delegated by user**. Compilation
is no deletion, preservation, durability, race or SLO acceptance result. Execute
under the department's existing authority, retain exact SHA/commands/logs/skips/
failures, and compare every asserted boundary with the plan. R90-75 independent
full acceptance and publication boundaries are unchanged.

Closeout source review caught the original daily Options.Path assumption after
its first feature push; corrected declarations must be reviewed at the final
checkout. R90-177's earlier daily alias fixtures share that source-proven
assumption and remain separate R90-179 repair debt. They do not establish the
original claimed direct boundary. No executed test result was changed.
