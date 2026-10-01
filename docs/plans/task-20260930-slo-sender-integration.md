# R90-127: Fresh sender reconstruction in bundle and pair review

## Selection and authority

Clean fetched HEAD/origin/main/FETCH_HEAD baseline is
`7dc6a1f38e80ea128b3e542af08e5af000f9eef6`. R90-126 feature/closure Vault
notes, full-index/MOC links and nine current stable notes are verified. Closure
Vault hash is `f797a21c598b6acbdf60933f6296c91610ce2922bb9cb69c2a10e030d97efd3e`.
September 2–30 history has 30 commits across queue/contract, measurement tools,
bundle/context/reconstruction and runbook/queue deliveries. No missing closure,
stale authority or new acceptance artifact changes priority. All 143 prior task
states parse; 131 unique roadmap rows/Definitions match. R90-127 depends on the
completed R90-126 implementation; R90-59 authority and R90-75 departmental
acceptance remain independent. Sep 30–Dec 28 horizon remains current.

Tests, knowledge suites and acceptance remain not run, delegated by the Sep 25
user instruction. Perform source/schema/data-flow/AST/docs/JSON/diff/sensitive
review and direct Git/remote/Vault verification only. Do not import/execute the
review tools, create measurement fixtures or claim direct regression evidence.

## Interface and schema contract before implementation

Add `reconstruct_sender=False` and `max_sender_reconstruction_bytes=64*1024**3`
to bundle.reconcile and compare.compare; CLI `--reconstruct-sender` and
`--max-sender-reconstruction-bytes`. Positive byte arguments validate regardless
of mode. Existing default and --reconstruct-ledgers behavior/schema remain.
Sender and adapter replay are independent selections; context remains orthogonal.

Add sender_reconstruct.integrate(retained, *, max_bytes) returning summary,
gaps, mismatches, errors. Base bundle checks must all complete first. Each selected
replay then runs from its own retained role directory; neither replay failure
suppresses the other. New nested directory is sender-reconstruction/, with
sender-reconstruction.json and four sender/ copies. Verify every complete nested
file/bytes/hash/rows inventory against the enclosing bundle before source binding.
Never read or accept an old nested receipt as fresh evidence. Partial output is
retained; I/O errors remain distinct from mismatches and gaps.

Sender-enabled bundle schema v3 includes base fields plus errors,
sender_reconstruction_policy=retained_sender_replay_v1,
sender_reconstruction, max_sender_reconstruction_bytes,
sender_reconstruction_source_sha256 and adapter_reconstruction_required boolean.
When the latter is true, include the existing adapter reconstruction fields too.
Sender summary has status, reconstruction_attempted, reconstruction_complete,
source_binding_complete, sender_records_match (true or null), and manifest
(file/bytes/sha256 at sender-reconstruction/sender-reconstruction.json, or null).
Completion requires attempted replay; bound success requires a manifest and
sender_records_match true. Both selected summaries must succeed for a complete
bundle. A shared nonempty errors list corresponds to at least one replay error;
one replay's errors must not make the other summary invalid.

Pair accepts original v1; v2 requires --reconstruct-ledgers; v3 requires
--reconstruct-sender and additionally --reconstruct-ledgers if the original
requires adapter replay. Missing required selection is invalid evidence, without
silent downgrade. Strict root/summary/policy/type/path/diagnostic coherence checks
precede use. Original failures stay failures even after successful current replay.
Only complete, original/current inventory-bound sides provide metrics, conditions
or context-binding identity. Relocate fresh summary references under each side.

Sender-enabled pair schema v4 uses exact_declared_repeatability_with_ followed by
optional context_and_, optional reconstruction_and_, then sender_reconstruction_v1.
It adds sender_reconstruction_required=true, adapter_reconstruction_required,
errors, max_sender_reconstruction_bytes_per_side and each side's sender summary;
existing adapter/context fields remain when selected. Status precedence remains
error > invalid_evidence > incomplete > conditions_differ > review_required.
Budgets are per nested operation/per side, not a total workspace quota. Sync child
entries and publish completion manifests last; late failures require process-exit
review. Physical facts/comparability/regression/SLO remain unasserted.

## Acceptance and evidence map

| Acceptance | Static source boundary | Departmental cases, explicitly not run |
| --- | --- | --- |
| Fresh sender binding | integrate uses only retained sender role, fresh manifest and exact four-file inventories | old receipt substitution, changed/incomplete/missing source and binding drift |
| Mode compatibility | bundle options/v1-v3 and pair options/v1-v4 contracts | default, adapter-only, sender-only, both, context combinations and mixed input versions |
| Status preservation | independent replay calls; combined diagnostics; strict original manifest validation | each original/current gap/mismatch/error, two different failures, asymmetric sides and invalid flags/types/policy/path |
| Comparison eligibility | original/current complete binding before conditions/context identity | no fallback after sender failure; unchanged failed/inconclusive measurement metrics |
| Provenance and resource failures | fixed nested references/digests, separate positive budgets, existing no-overwrite and fsync pattern | budget exhaustion, read/write/close/fsync/interruption, partial manifest, resource scale |
| Delivery | AST/docs/JSON/roadmap/history/links/diff/sensitive review, verified push and exact Vault ranges | tests and knowledge suites remain delegated |

## Scope and stop boundaries

Ten intended paths: scripts/slo_sender_reconstruct.py, scripts/slo_bundle.py,
scripts/slo_compare.py; docs/slo-sender-reconstruct.md, docs/slo-bundle.md,
docs/slo-compare.md, docs/slo-runbook.md; roadmap, this plan and matching state.
No standalone sender algorithm change, traffic, new protocol, oracle generation,
private data, dependency, CI, thresholds or release action. Stop for undocumented
formats or new product/private/external authority. Refresh future scope without
starting it. One feature commit and one docs-only delivery closure, each pushed,
fetched and synchronized to the sole local Vault; preserve immutable notes.

## Static acceptance checkpoint

All ten intended paths are implemented and reviewed against the interface above.
The four-source binder calls fresh reconstruction only on the enclosing sender
snapshots; original nested receipts never select inputs. Both replay calls finish
before diagnostic aggregation, preserving independent outcomes. The shared
summary validator retains v2 semantics and validates selected v3 operation/error
coherence. Pair eligibility remains behind original/current success and inventory
binding; current summary paths are relocated under the relevant side.

AST/python-check, docs-check, 144 task-state JSONs, 131 unique roadmap multisets,
ordered history, local links/fences, exact scope/diff and sensitive-information
review pass. Every acceptance criterion has source-level evidence and explicit
unexecuted departmental cases; no test or runtime correctness claim is made.
No scope deviation or reusable skill update was needed. Once delivered, no local
ready increment remains; the next trigger must audit the forward queue without
repeating this delivery. R90-59/R90-75 retain their independent outstanding
contracts. Feature delivery and the single docs-only closure remain.

## Verified delivery and resume boundary

Feature `0b6283e7e6474ef6a1e8f68ef751de35832733c5` contains exactly ten intended paths. Push without
force/tags and fresh fetch verified clean HEAD/origin/main/FETCH_HEAD equality.
Exact range `7dc6a1f38e80ea128b3e542af08e5af000f9eef6..0b6283e7e6474ef6a1e8f68ef751de35832733c5`
was synchronized to the sole local Vault. Note/index/MOC and nine stable notes
are verified; pre-existing immutable iteration bytes are preserved. Identical-
range replay retained Vault Markdown hash
`66feca85909ab71e63eec079562c5fa50b982fa2db8c1100628b5dd2a419f023`.

Every acceptance row is covered by source review and documented unexecuted
departmental cases. No scope deviation, behavioral proof, acceptance outcome or
skill change is claimed. R90-127 is complete as implementation delivery under
the standing testing split. No local ready increment remains; R90-59 authority
and R90-75 departmental acceptance retain their full independent contracts.

This single docs-only closure receives its own verified push/fetch/exact Vault
range. The next trigger must verify that closure against fetched origin/main and
audit the forward queue before selecting work. Do not repeat this feature's
commit/push/sync, start unplanned consumer work or execute delegated tests.
