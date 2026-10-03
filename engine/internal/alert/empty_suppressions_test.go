package alert_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strconv"
	"testing"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/pkg/model"
)

type emptySuppressionCase struct {
	name string
	rule alert.Suppression
}

func emptySuppressionCases() []emptySuppressionCase {
	cases := []emptySuppressionCase{
		{"nil_lists", alert.Suppression{}},
		{"empty_lists", alert.Suppression{SrcCIDRs: []string{}, DstCIDRs: []string{}, AnyCIDRs: []string{}}},
	}
	for mask := 1; mask < 8; mask++ {
		rule := alert.Suppression{}
		if mask&1 != 0 {
			rule.SrcCIDRs = []string{"", ""}
		}
		if mask&2 != 0 {
			rule.DstCIDRs = []string{"", ""}
		}
		if mask&4 != 0 {
			rule.AnyCIDRs = []string{"", ""}
		}
		cases = append(cases, emptySuppressionCase{fmt.Sprintf("empty_entries_%d", mask), rule})
	}
	return cases
}

func copyEmptySuppressionRules(rules []alert.Suppression) []alert.Suppression {
	if rules == nil {
		return nil
	}
	out := make([]alert.Suppression, len(rules))
	copy(out, rules)
	clone := func(values []string) []string {
		if values == nil {
			return nil
		}
		result := make([]string, len(values))
		copy(result, values)
		return result
	}
	for i := range out {
		out[i].RuleIDs = clone(rules[i].RuleIDs)
		out[i].SrcCIDRs = clone(rules[i].SrcCIDRs)
		out[i].DstCIDRs = clone(rules[i].DstCIDRs)
		out[i].AnyCIDRs = clone(rules[i].AnyCIDRs)
	}
	return out
}

func assertEmptySuppressionError(t *testing.T, err error, id string) {
	t.Helper()
	want := "suppression " + strconv.Quote(id) + " must include at least one CIDR"
	if err == nil || err.Error() != want {
		t.Fatalf("error = %v, want %q", err, want)
	}
}

func readEmptySuppressionFile(t *testing.T, path string) []byte {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func writeEmptySuppressionFile(t *testing.T, path string, rules []alert.Suppression) []byte {
	t.Helper()
	data, err := json.MarshalIndent(struct {
		Suppressions []alert.Suppression `json:"suppressions"`
	}{Suppressions: rules}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	data = append(data, '\n')
	if err := os.WriteFile(path, data, 0o600); err != nil {
		t.Fatal(err)
	}
	return data
}

func emptySuppressionPath(t *testing.T) string {
	t.Helper()
	dir := filepath.Join(t.TempDir(), "suppression state with spaces")
	if err := os.Mkdir(dir, 0o700); err != nil {
		t.Fatal(err)
	}
	return filepath.Join(dir, "suppression rules.json")
}

func TestSuppressionConstructorsRejectNoCompiledPrefixesWithoutChangingInputs(t *testing.T) {
	for _, tc := range emptySuppressionCases() {
		t.Run(tc.name, func(t *testing.T) {
			candidate := tc.rule
			candidate.ID, candidate.Enabled = "empty\"candidate", true
			candidate.RuleIDs = []string{"", "scoped-rule"}
			rules := []alert.Suppression{candidate}
			before := copyEmptySuppressionRules(rules)
			if got, err := alert.NewSuppressor(rules); got != nil {
				t.Fatalf("rejected suppressor = %+v, error = %v", got, err)
			} else {
				assertEmptySuppressionError(t, err, candidate.ID)
			}
			if got, err := alert.NewSuppressionManager(rules); got != nil {
				t.Fatalf("rejected manager = %+v, error = %v", got, err)
			} else {
				assertEmptySuppressionError(t, err, candidate.ID)
			}
			path := emptySuppressionPath(t)
			fileBefore := writeEmptySuppressionFile(t, path, []alert.Suppression{{ID: "existing", Enabled: true, AnyCIDRs: []string{"192.0.2.0/24"}}})
			if got, err := alert.NewSuppressionManagerWithFile(rules, path); got != nil {
				t.Fatalf("rejected file manager = %+v, error = %v", got, err)
			} else {
				assertEmptySuppressionError(t, err, candidate.ID)
			}
			if !reflect.DeepEqual(rules, before) || !bytes.Equal(readEmptySuppressionFile(t, path), fileBefore) {
				t.Fatal("constructor rejection changed caller slices or existing file bytes")
			}
		})
	}
}

func TestSuppressionConstructorsPreserveEmptySetsAndDisabledPrefixSkipping(t *testing.T) {
	cases := []struct {
		name  string
		rules []alert.Suppression
	}{
		{"nil_set", nil}, {"empty_set", []alert.Suppression{}},
		{"disabled_nil", []alert.Suppression{{ID: "disabled"}}},
		{"disabled_empty", []alert.Suppression{{ID: "disabled", SrcCIDRs: []string{}, DstCIDRs: []string{}, AnyCIDRs: []string{}}}},
		{"disabled_empty_entries", []alert.Suppression{{ID: "disabled", SrcCIDRs: []string{""}, DstCIDRs: []string{""}, AnyCIDRs: []string{""}}}},
		{"disabled_invalid", []alert.Suppression{{ID: "disabled", SrcCIDRs: []string{"bad-src"}, DstCIDRs: []string{"bad-dst"}, AnyCIDRs: []string{" \t"}}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			before := copyEmptySuppressionRules(tc.rules)
			suppressor, err := alert.NewSuppressor(tc.rules)
			if err != nil || suppressor == nil {
				t.Fatalf("NewSuppressor = %+v, %v", suppressor, err)
			}
			manager, err := alert.NewSuppressionManager(tc.rules)
			if err != nil || manager == nil || len(manager.List()) != len(tc.rules) {
				t.Fatalf("NewSuppressionManager = %+v, %v", manager, err)
			}
			path := emptySuppressionPath(t)
			fileBefore := writeEmptySuppressionFile(t, path, tc.rules)
			fileManager, err := alert.NewSuppressionManagerWithFile(tc.rules, path)
			if err != nil || fileManager == nil || len(fileManager.List()) != len(tc.rules) {
				t.Fatalf("NewSuppressionManagerWithFile = %+v, %v", fileManager, err)
			}
			probe := &model.Alert{ID: "probe", RuleID: "scoped-rule", SrcIP: "192.0.2.9", DstIP: "198.51.100.7", PayloadPreview: "public-marker"}
			probeBefore := *probe
			input := []*model.Alert{nil, probe}
			if suppressor.Suppressed(nil) || suppressor.Suppressed(probe) || !reflect.DeepEqual(suppressor.Filter(input), input) || !reflect.DeepEqual(manager.Filter(input), input) || !reflect.DeepEqual(fileManager.Filter(input), input) {
				t.Fatal("empty or disabled rules unexpectedly suppressed alerts")
			}
			if !reflect.DeepEqual(tc.rules, before) || *probe != probeBefore || !bytes.Equal(readEmptySuppressionFile(t, path), fileBefore) {
				t.Fatal("accepted empty/disabled construction changed inputs or file")
			}
		})
	}
}

func TestSuppressionMixedEmptyAndValidPrefixesRetainMatchingAndScoping(t *testing.T) {
	prefixCases := []struct {
		name, value, match, outside string
	}{
		{"ipv4_exact", "192.0.2.9", "192.0.2.9", "192.0.2.10"},
		{"ipv4_masked_cidr", "198.51.100.42/24", "198.51.100.77", "198.51.101.77"},
		{"ipv6_exact", "2001:db8:1::9", "2001:db8:1::9", "2001:db8:1::a"},
		{"ipv6_masked_cidr", "2001:db8:2::abcd/64", "2001:db8:2::7", "2001:db8:3::7"},
	}
	for _, direction := range []string{"src", "dst", "any"} {
		for _, tc := range prefixCases {
			t.Run(direction+"/"+tc.name, func(t *testing.T) {
				candidate := alert.Suppression{ID: "mixed", Enabled: true, RuleIDs: []string{"", "scoped-rule", "scoped-rule"}}
				values := []string{"", tc.value, tc.value, ""}
				switch direction {
				case "src":
					candidate.SrcCIDRs = values
				case "dst":
					candidate.DstCIDRs = values
				case "any":
					candidate.AnyCIDRs = values
				}
				rules := []alert.Suppression{candidate, {ID: "disabled", SrcCIDRs: []string{"not-an-ip"}}}
				before := copyEmptySuppressionRules(rules)
				suppressor, err := alert.NewSuppressor(rules)
				if err != nil {
					t.Fatal(err)
				}
				manager, err := alert.NewSuppressionManager(rules)
				if err != nil {
					t.Fatal(err)
				}
				for _, placement := range []string{"src", "dst"} {
					positive := &model.Alert{RuleID: "scoped-rule", SrcIP: tc.outside, DstIP: tc.outside, PayloadPreview: "public-marker"}
					if placement == "src" {
						positive.SrcIP = tc.match
					} else {
						positive.DstIP = tc.match
					}
					positiveBefore := *positive
					outside := &model.Alert{RuleID: "scoped-rule", SrcIP: tc.outside, DstIP: tc.outside}
					wrongRule := *positive
					wrongRule.RuleID = "other-rule"
					invalidIPs := &model.Alert{RuleID: "scoped-rule", SrcIP: "bad-src", DstIP: "bad-dst"}
					input := []*model.Alert{nil, positive, outside, &wrongRule, invalidIPs}
					wantMatch := direction == "any" || direction == placement
					want := input
					if wantMatch {
						want = []*model.Alert{nil, outside, &wrongRule, invalidIPs}
					}
					if suppressor.Suppressed(positive) != wantMatch || suppressor.Suppressed(outside) || suppressor.Suppressed(&wrongRule) || suppressor.Suppressed(invalidIPs) || !reflect.DeepEqual(suppressor.Filter(input), want) || !reflect.DeepEqual(manager.Filter(input), want) {
						t.Fatalf("matching changed for %s placement; suppressor=%+v manager=%+v", placement, suppressor.Filter(input), manager.Filter(input))
					}
					if *positive != positiveBefore || input[0] != nil || input[1] != positive || input[2] != outside || input[3] != &wrongRule || input[4] != invalidIPs {
						t.Fatal("matching changed caller alert or input slice")
					}
				}
				if !reflect.DeepEqual(rules, before) {
					t.Fatal("compilation or matching changed caller prefix or rule-ID slices")
				}
			})
		}
	}
	// Empty rule IDs still mean every detection rule, and valid prefixes in one
	// direction remain useful when other directions contain only empty entries.
	rules := []alert.Suppression{{ID: "all-rules", Enabled: true, RuleIDs: []string{""}, SrcCIDRs: []string{""}, DstCIDRs: []string{""}, AnyCIDRs: []string{"", "203.0.113.0/24"}}}
	before := copyEmptySuppressionRules(rules)
	suppressor, err := alert.NewSuppressor(rules)
	if err != nil {
		t.Fatal(err)
	}
	for _, id := range []string{"", "rule-one", "rule-two"} {
		if !suppressor.Suppressed(&model.Alert{RuleID: id, SrcIP: "198.51.100.1", DstIP: "203.0.113.7"}) {
			t.Fatalf("empty rule-ID scope did not match %q", id)
		}
	}
	if !reflect.DeepEqual(rules, before) {
		t.Fatal("unscoped compilation changed input")
	}
	mixed := []alert.Suppression{{ID: "exact-and-cidr", Enabled: true, SrcCIDRs: []string{"", "192.0.2.9", "198.51.100.42/24", ""}}}
	mixedBefore := copyEmptySuppressionRules(mixed)
	mixedSuppressor, err := alert.NewSuppressor(mixed)
	if err != nil {
		t.Fatal(err)
	}
	for _, source := range []string{"192.0.2.9", "198.51.100.77"} {
		probe := &model.Alert{SrcIP: source, DstIP: "203.0.113.7"}
		probeBefore := *probe
		if !mixedSuppressor.Suppressed(probe) || len(mixedSuppressor.Filter([]*model.Alert{probe})) != 0 || *probe != probeBefore {
			t.Fatalf("mixed exact/CIDR list did not suppress %q without changing it", source)
		}
	}
	if mixedSuppressor.Suppressed(&model.Alert{SrcIP: "192.0.2.10", DstIP: "198.51.100.77"}) || !reflect.DeepEqual(mixed, mixedBefore) {
		t.Fatal("mixed source filter matched destination or changed caller input")
	}
}

func TestSuppressionInvalidPrefixesKeepPriorDiagnosticOrder(t *testing.T) {
	cases := []struct {
		name, field, invalid string
		rule                 alert.Suppression
	}{
		{"src_before_dst_and_any", "src", "bad-src", alert.Suppression{SrcCIDRs: []string{"", "bad-src"}, DstCIDRs: []string{"bad-dst"}, AnyCIDRs: []string{"bad-any"}}},
		{"dst_before_any", "dst", "bad-dst", alert.Suppression{SrcCIDRs: []string{""}, DstCIDRs: []string{"", "bad-dst"}, AnyCIDRs: []string{"bad-any"}}},
		{"any_after_empty_src_dst", "any", "bad-any", alert.Suppression{SrcCIDRs: []string{""}, DstCIDRs: []string{""}, AnyCIDRs: []string{"", "bad-any"}}},
		{"src_whitespace", "src", " \t", alert.Suppression{SrcCIDRs: []string{"", " \t"}}},
		{"dst_whitespace", "dst", "\u2003", alert.Suppression{DstCIDRs: []string{"", "\u2003"}}},
		{"any_whitespace", "any", "\n", alert.Suppression{AnyCIDRs: []string{"", "\n"}}},
		{"src_invalid_cidr", "src", "192.0.2.9/33", alert.Suppression{SrcCIDRs: []string{"", "192.0.2.9/33"}}},
		{"dst_invalid_cidr", "dst", "2001:db8::9/129", alert.Suppression{DstCIDRs: []string{"", "2001:db8::9/129"}}},
		{"any_invalid_cidr", "any", "203.0.113.9/no", alert.Suppression{AnyCIDRs: []string{"", "203.0.113.9/no"}}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			tc.rule.ID, tc.rule.Enabled = "invalid", true
			rules := []alert.Suppression{tc.rule}
			before := copyEmptySuppressionRules(rules)
			want := "suppression invalid " + tc.field + " cidrs: invalid IP or CIDR " + strconv.Quote(tc.invalid)
			if result, err := alert.NewSuppressor(rules); result != nil || err == nil || err.Error() != want {
				t.Fatalf("NewSuppressor = %+v, %v; want %q", result, err, want)
			}
			if result, err := alert.NewSuppressionManager(rules); result != nil || err == nil || err.Error() != want {
				t.Fatalf("NewSuppressionManager = %+v, %v; want %q", result, err, want)
			}
			if !reflect.DeepEqual(rules, before) {
				t.Fatal("prefix rejection changed caller input")
			}
		})
	}
}

func TestFileBackedSuppressionMutationsRejectEmptyCompiledPrefixesAndPermitRetry(t *testing.T) {
	for _, operation := range []string{"add", "update", "reload"} {
		for _, tc := range emptySuppressionCases() {
			t.Run(operation+"/"+tc.name, func(t *testing.T) {
				path := emptySuppressionPath(t)
				initial := []alert.Suppression{{ID: "original", Enabled: true, RuleIDs: []string{"scoped-rule"}, SrcCIDRs: []string{"192.0.2.9"}}}
				writeEmptySuppressionFile(t, path, initial)
				loaded, err := alert.LoadSuppressionsFromFile(path)
				if err != nil {
					t.Fatal(err)
				}
				manager, err := alert.NewSuppressionManagerWithFile(loaded, path)
				if err != nil {
					t.Fatal(err)
				}
				oldMatch := &model.Alert{ID: "old", RuleID: "scoped-rule", SrcIP: "192.0.2.9", DstIP: "198.51.100.1"}
				newMatch := &model.Alert{ID: "new", RuleID: "scoped-rule", SrcIP: "203.0.113.7", DstIP: "198.51.100.1"}
				other := &model.Alert{ID: "outside", RuleID: "other-rule", SrcIP: "192.0.2.9", DstIP: "203.0.113.7"}
				probes := []*model.Alert{nil, oldMatch, newMatch, other}
				probeBefore := []model.Alert{*oldMatch, *newMatch, *other}
				wantBefore := []*model.Alert{nil, newMatch, other}
				listBefore := manager.List()
				if !reflect.DeepEqual(listBefore, initial) || !reflect.DeepEqual(manager.Filter(probes), wantBefore) {
					t.Fatal("initial real-file manager did not publish expected rules/filter")
				}
				candidate := tc.rule
				candidate.ID, candidate.Enabled = "candidate", true
				candidate.RuleIDs = []string{"scoped-rule"}
				candidateBefore := copyEmptySuppressionRules([]alert.Suppression{candidate})[0]
				wantID := candidate.ID
				if operation == "reload" {
					writeEmptySuppressionFile(t, path, []alert.Suppression{candidate})
					// Nonempty raw lists pass the loader's structural check. Actual
					// ReloadFromFile must reject them at compilation, using real I/O.
					if len(candidate.SrcCIDRs)+len(candidate.DstCIDRs)+len(candidate.AnyCIDRs) > 0 {
						parsed, err := alert.LoadSuppressionsFromFile(path)
						if err != nil || !reflect.DeepEqual(parsed, []alert.Suppression{candidate}) {
							t.Fatalf("structural load = %+v, %v", parsed, err)
						}
					}
				}
				fileBefore := readEmptySuppressionFile(t, path)
				switch operation {
				case "add":
					err = manager.Add(candidate)
				case "update":
					wantID = "original"
					err = manager.Update(wantID, candidate)
				case "reload":
					err = manager.ReloadFromFile()
				}
				if operation == "reload" && len(candidate.SrcCIDRs)+len(candidate.DstCIDRs)+len(candidate.AnyCIDRs) == 0 {
					want := "validate suppressions " + path + ": suppression " + strconv.Quote(wantID) + " must include at least one CIDR"
					if err == nil || err.Error() != want {
						t.Fatalf("structural reload error = %v, want %q", err, want)
					}
				} else {
					assertEmptySuppressionError(t, err, wantID)
				}
				if !reflect.DeepEqual(manager.List(), listBefore) || !reflect.DeepEqual(manager.Filter(probes), wantBefore) || !bytes.Equal(readEmptySuppressionFile(t, path), fileBefore) {
					t.Fatal("rejected mutation changed published list/filter or complete existing file bytes")
				}
				if !reflect.DeepEqual(candidate, candidateBefore) || *oldMatch != probeBefore[0] || *newMatch != probeBefore[1] || *other != probeBefore[2] || probes[0] != nil || probes[1] != oldMatch || probes[2] != newMatch || probes[3] != other {
					t.Fatal("rejected mutation changed caller candidate, alert or input slice")
				}
				valid := alert.Suppression{ID: "candidate", Enabled: true, RuleIDs: []string{"scoped-rule"}, AnyCIDRs: []string{"", "203.0.113.41/24", ""}}
				validBefore := copyEmptySuppressionRules([]alert.Suppression{valid})[0]
				wantRules := []alert.Suppression{valid}
				wantFiltered := []*model.Alert{nil, oldMatch, other}
				var retryFile []byte
				switch operation {
				case "add":
					err = manager.Add(valid)
					wantRules = append(copyEmptySuppressionRules(initial), valid)
					wantFiltered = []*model.Alert{nil, other}
				case "update":
					err = manager.Update("original", valid)
					wantRules[0].ID = "original"
				case "reload":
					retryFile = writeEmptySuppressionFile(t, path, wantRules)
					err = manager.ReloadFromFile()
				}
				if err != nil || !reflect.DeepEqual(manager.List(), wantRules) || !reflect.DeepEqual(manager.Filter(probes), wantFiltered) {
					t.Fatalf("valid retry error=%v list=%+v filter=%+v", err, manager.List(), manager.Filter(probes))
				}
				persisted, err := alert.LoadSuppressionsFromFile(path)
				if err != nil || !reflect.DeepEqual(persisted, wantRules) {
					t.Fatalf("retry file rules=%+v error=%v", persisted, err)
				}
				if operation == "reload" && !bytes.Equal(readEmptySuppressionFile(t, path), retryFile) {
					t.Fatal("successful reload rewrote supplied canonical file")
				}
				if !reflect.DeepEqual(valid, validBefore) || *oldMatch != probeBefore[0] || *newMatch != probeBefore[1] || *other != probeBefore[2] {
					t.Fatal("valid retry changed caller candidate or alerts")
				}
			})
		}
	}
}
