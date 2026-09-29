# Supplied SLO run context

R90-120 adds `scripts/slo_context.py`: a versioned declaration validator and
retention API/CLI for hardware, toolchain and isolation information. It binds
supplied declarations and opaque supporting files to exact observation bytes.
It performs no resource discovery, shell execution, network access, service startup
or acceptance run. The implementation has received static review only; all
behavioral and acceptance validation is delegated to the specialist department.

## Department invocation

An unexecuted template, using prepared inputs and a new output directory:

```bash
python3 scripts/slo_context.py \
  --declaration run-context.json \
  --observations retained-run/observations.json \
  --evidence-dir context-evidence \
  --output-dir retained-context
```

`validate(document)` validates the declaration and returns field-coverage gaps;
malformed input raises `EvidenceError`. `retain(output, *, declaration=None,
observations=None, evidence_dir=None, max_evidence_bytes=256*1024**2)` accepts
`pathlib.Path` arguments and publishes the retained context. Missing inputs are
permitted so incomplete handoffs can record explicit gaps. No existing output
directory/file is overwritten and no evidence is deleted automatically.

## Declaration schema v1

Root fields are exactly `schema_version: 1`,
`artifact_kind: "netsentry_slo_run_context"`, `run_id`, `profile`, `started_at`,
`observations_sha256`, `hardware`, `toolchain`, `isolation`, `evidence`.
The run ID uses the reporter's bounded ASCII identifier grammar; profile is
`staging` or `prod`; start is canonical whole-second UTC `YYYY-MM-DDTHH:MM:SSZ`.
The observation SHA-256 is lowercase hex over the exact supplied file bytes.
Identity fields must be supplied; the tool never invents them.

Every field in the following table is mandatory and contains exactly
`{"value": ..., "evidence_ids": [...]}`. Unknown values are JSON `null`.
Known values without evidence references remain explicit gaps. Null may have
references (for example, evidence documenting an unknown); it remains unknown.
A boolean `false` is a known declaration, never silently converted to unknown.

| Section | Field | Non-null value type |
| --- | --- | --- |
| hardware | deployment_type | `kvm_vm` or `physical` |
| hardware | cpu_arch | `x86_64`, `aarch64` or `other` |
| hardware | cpu_model | text |
| hardware | cpu_pinning | text describing the supplied allocation |
| hardware | sut_vcpus | positive integer |
| hardware | sut_memory_bytes | positive integer, bytes |
| hardware | storage_model | text |
| hardware | storage_medium | `ssd`, `nvme`, `hdd` or `other` |
| hardware | storage_capacity_bytes | positive integer, bytes |
| hardware | storage_dedicated_partition | boolean |
| hardware | nic_model | text |
| hardware | nic_mode | `physical`, `sriov_vf`, `virtual` or `other` |
| hardware | nic_speed_bps | positive integer, bits/second |
| hardware | rss_enabled | boolean |
| toolchain | os_release | text |
| toolchain | kernel_release | text |
| toolchain | go_version | text |
| toolchain | c_compiler | text including exact compiler identity/version |
| toolchain | libpcap_version | text |
| toolchain | build_flags | text; explicitly state no extra flags if appropriate |
| isolation | same_host | boolean: generator and SUT share a host |
| isolation | separate_workdir | boolean |
| isolation | separate_process_groups | boolean |
| isolation | fresh_runtime | boolean: no production runtime/state reuse |
| isolation | allocation_notes | text describing supplied resource/isolation facts |

Text is 1–512 UTF-8 bytes, printable and without surrounding whitespace.
Positive integers are actual JSON integers from 1 through `2^63-1`, excluding
booleans/floats. These are schema bounds, not hardware recommendations. Values
such as HDD storage, disabled RSS or false isolation flags are retained as stated;
the tool does not decide whether they satisfy a deployment target. Missing keys,
unknown keys, invalid enum/type values, duplicate JSON members, nonfinite numbers
and malformed UTF-8/JSON are errors, not substituted defaults.

## Evidence catalog and references

`evidence` is an array of at most 64 objects with exactly:

```text
id:          [A-Za-z0-9][A-Za-z0-9_-]{0,63}
description: bounded text explaining the supplied reference
bytes:       positive integer, exact original file byte count
sha256:      64 lowercase hexadecimal characters
```

IDs are case-sensitive and unique. Every field's `evidence_ids` array contains
at most 64 distinct catalog IDs; unknown IDs are invalid. Each catalog entry must
be referenced by at least one field. One reference may support several fields.

The only source filename is `<id>.bin` inside the explicit evidence directory.
No document-supplied path, URI, command, extension or traversal component is
followed. The `.bin` suffix denotes opaque bytes: rename/copy the already prepared
file under that name without changing its content. The tool retains it without
interpreting or executing it. Reviewers determine whether its contents actually
support the referenced declaration. Hash agreement proves byte identity only.
Keep operational evidence local; only use material authorized for the handoff.

For example, an unknown cell is `{"value": null, "evidence_ids": []}`. A known
cell references catalog IDs through `{"value": true, "evidence_ids": ["isolation"]}`.
The latter requires a corresponding catalog entry and supplied `isolation.bin`;
it is only a schema example, not evidence that isolation exists.

## Retained package and status

| Output | Meaning |
| --- | --- |
| `declaration.json` | Exact input bytes, or an explicitly incomplete prefix |
| `observations.json` | Exact supplied observation bytes, or an explicitly incomplete prefix |
| `evidence/<id>.bin` | Opaque referenced source bytes, retained only from a structurally valid declaration |
| `context.json` | Published last: status, run binding, normalized declared fields, inventory/digests, gaps, mismatches and completed checks |

Observations must satisfy the existing [reporter schema](slo-report.md).
Run ID, profile, start time and raw observation digest must match the declaration.
Known `hardware.sut_vcpus` and `hardware.sut_memory_bytes` must also agree with
observation resource values; unknown values stay gaps and are not inferred from
the observations. No live packet reconstruction or measurement recomputation is
performed. The original observation file remains available for separate review.

| Status / CLI exit | Meaning |
| --- | --- |
| `review_required` / 0 | Valid declaration, known values with references, complete matching retained evidence and observation binding; department review required |
| `mismatch` / 1 | At least one invalid schema, changed source, digest/byte mismatch or inconsistent run/resource binding |
| error / 2 | Invalid arguments or output I/O failure; retain partial output and inspect process status |
| `incomplete` / 3 | No detected mismatch, but unknown values, unreferenced known values, missing sources or byte limits leave gaps |

Mismatches take precedence over gaps; both lists remain in the receipt.
`context_complete` is true only for `review_required`, meaning retention and
implemented checks completed. `observation_binding_complete` only describes the
observation identity/digest/allocation check; supporting evidence may still be
missing. `declaration_schema_valid` does not imply field coverage or factual truth.
Every result keeps `execution_complete`, `facts_verified`,
`comparability_established` and `slo_compliance_asserted` false, with departmental
review required. Even complete packages can contain declarations that violate a
proposed target; no acceptance or actual-environment verdict is made.

## Bounds and failure behavior

Declaration JSON is capped at 1 MiB and observations at 64 MiB. Each supporting
file is capped at 64 MiB, with a default 256 MiB combined supporting-file budget;
`--max-evidence-bytes` changes the combined budget only. Metadata/observations
have their own limits and are additional disk use, as is the small final receipt.
At most 64 references are retained in catalog order. Opaque copying streams bounded
chunks; JSON parsing uses memory proportional to these finite JSON limits.

Inputs must be regular files; direct file symlinks and special files are not read.
Input size/mtime/ctime/identity changes during copying become mismatches. Use
quiescent trusted paths; this is not an adversarial filesystem guarantee. A missing
source, read failure or exceeded byte budget retains any copied prefix with
`complete: false` and records a gap. An invalid declaration remains retained but
cannot select reference files for copying. Files use mode 0600 and directories
0700. Files and evidence directories are synced before `context.json` is
exclusively published, then the output parent is synced. Partial directories or a
visible receipt can remain after interruptions/late sync failures; inspect exit
status and retry to a fresh destination. No scale or performance is demonstrated.

## Comparison consumer contract

R90-121 adds optional context consumption to [the pair comparator](slo-compare.md).
Providing a context package or requiring context enables its version-2 policy;
the default without context options keeps the version-1 path. A consumer must
retain/revalidate the declaration and
references, verify the receipt against its snapshots, bind exact run/profile/start
and observation SHA-256 to the corresponding bundle, and compare declared values
while preserving null/unsupported/mismatched evidence. Evidence IDs and raw file
digests are source identity, not equality conditions for machine properties.
Differing evidence files can legitimately support equal declarations.

Unknown fields cannot count as a match that establishes comparability. Even known
matching declarations with matching retained references remain subject to factual
review. Binding must never replace live clock/durability, packet-oracle, hardware
or isolation qualification. The consumer rechecks original/fresh inventories and binds each context to its
exact bundle observations. Unknown, unsupported and unavailable fields remain
explicit, and null values cannot count as a match. Known values may be compared
individually within an incomplete context, while all original gaps remain.

## Required departmental validation — not executed

| Boundary | Required cases |
| --- | --- |
| Schema | Every field/type/enum, null and false, unknown/missing keys, UTF-8/control/length bounds, huge integers/nonfinite numbers, duplicate members, invalid identity/digest |
| References | Duplicate/unknown/unused IDs, duplicate field references, limits, empty catalogs, known values without references, unknown values with references, opaque non-JSON contents, byte/hash drift |
| Binding | Run/profile/start/digest drift, known allocation mismatch, unknown allocation remains unknown, invalid observation schema, missing observation input |
| Filesystem | Regular/symlink/FIFO/directory, source mutation, exact/exceeded per-file/combined limits, missing/read errors, output preservation/modes, disk/write/fsync/close/interruption faults and partial receipts |
| Semantics | False isolation/resource-target violations remain declarations, irrelevant but digest-correct evidence does not certify facts, no discovery/command execution, consumer identity binding and unknown handling |

No context package or qualifying measurement was generated by the agent. Both
profiles' actual acceptance remains outstanding under the [SLO contract](performance-slo.md).
