package alert_test

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/pkg/model"
)

func TestSuppressionConstructorsRejectExplicitEmptyRuleScope(t *testing.T) {
	for _, ids := range [][]string{{""}, {"", ""}} {
		for _, rangeField := range []string{"src", "dst", "any"} {
			t.Run(fmt.Sprintf("ids=%d/%s", len(ids), rangeField), func(t *testing.T) {
				candidate := alert.Suppression{ID: "empty\"scope", Enabled: true, RuleIDs: ids}
				switch rangeField {
				case "src":
					candidate.SrcCIDRs = []string{"192.0.2.9"}
				case "dst":
					candidate.DstCIDRs = []string{"198.51.100.0/24"}
				case "any":
					candidate.AnyCIDRs = []string{"203.0.113.0/24"}
				}
				// A valid prefix candidate must not escape as a partial result.
				rules := []alert.Suppression{{ID: "valid", Enabled: true, AnyCIDRs: []string{"192.0.2.0/24"}}, candidate}
				before := copyEmptySuppressionRules(rules)
				want := fmt.Sprintf("suppression %q rule_ids must include at least one nonempty rule ID", candidate.ID)
				if got, err := alert.NewSuppressor(rules); got != nil || err == nil || err.Error() != want {
					t.Fatalf("NewSuppressor = %+v, %v; want nil, %q", got, err, want)
				}
				if got, err := alert.NewSuppressionManager(rules); got != nil || err == nil || err.Error() != want {
					t.Fatalf("NewSuppressionManager = %+v, %v; want nil, %q", got, err, want)
				}
				if !reflect.DeepEqual(rules, before) {
					t.Fatal("rejection changed caller rules")
				}
			})
		}
	}
}

func TestSuppressionRuleScopePreservesAcceptedSemantics(t *testing.T) {
	cases := []struct {
		name         string
		ids          []string
		enabled      bool
		matchingID   string
		otherMatches bool
	}{
		{"nil_unrestricted", nil, true, "rule-1", true},
		{"empty_unrestricted", []string{}, true, "rule-1", true},
		{"mixed_exact_duplicates", []string{"", "rule-1", "rule-1", ""}, true, "rule-1", false},
		{"literal_spaces", []string{"", " rule-1 "}, true, " rule-1 ", false},
		{"literal_whitespace", []string{" \t"}, true, " \t", false},
		{"disabled_empty_only", []string{"", ""}, false, "rule-1", false},
	}
	for _, tc := range cases {
		for _, field := range []string{"src", "dst", "any"} {
			t.Run(tc.name+"/"+field, func(t *testing.T) {
				rule := alert.Suppression{ID: "scope", Enabled: tc.enabled, RuleIDs: tc.ids}
				switch field {
				case "src":
					rule.SrcCIDRs = []string{"192.0.2.0/24"}
				case "dst":
					rule.DstCIDRs = []string{"192.0.2.0/24"}
				case "any":
					rule.AnyCIDRs = []string{"192.0.2.0/24"}
				}
				rules := []alert.Suppression{rule}
				before := copyEmptySuppressionRules(rules)
				suppressor, err := alert.NewSuppressor(rules)
				if err != nil {
					t.Fatal(err)
				}
				for _, sourceHit := range []bool{true, false} {
					src, dst := "192.0.2.9", "198.51.100.7"
					if !sourceHit {
						src, dst = dst, src
					}
					for _, id := range []string{tc.matchingID, "unrelated-rule", "RULE-1", "rule-1"} {
						probe := &model.Alert{RuleID: id, SrcIP: src, DstIP: dst}
						probeBefore := *probe
						input := []*model.Alert{nil, probe}
						want := tc.enabled && (id == tc.matchingID || tc.otherMatches) && (field == "any" || (field == "src") == sourceHit)
						if got := suppressor.Suppressed(probe); got != want {
							t.Fatalf("id=%q src=%s dst=%s suppressed=%t want=%t", id, src, dst, got, want)
						}
						filtered := suppressor.Filter(input)
						wantFiltered := []*model.Alert{nil, probe}
						if want {
							wantFiltered = []*model.Alert{nil}
						}
						if !reflect.DeepEqual(filtered, wantFiltered) || *probe != probeBefore || input[0] != nil || input[1] != probe {
							t.Fatal("filter changed identity/input or returned incorrect contents")
						}
					}
				}
				if suppressor.Suppressed(&model.Alert{RuleID: tc.matchingID, SrcIP: "203.0.113.1", DstIP: "203.0.113.2"}) || !reflect.DeepEqual(rules, before) {
					t.Fatal("range rejection or rule preservation changed")
				}
			})
		}
	}
}

func TestSuppressionEmptyRuleScopePreservesPrefixDiagnosticPrecedence(t *testing.T) {
	for _, field := range []string{"missing", "src", "dst", "any"} {
		t.Run(field, func(t *testing.T) {
			rule := alert.Suppression{ID: "scope", Enabled: true, RuleIDs: []string{""}}
			want := `suppression "scope" must include at least one CIDR`
			switch field {
			case "src":
				rule.SrcCIDRs = []string{"bad"}
			case "dst":
				rule.SrcCIDRs, rule.DstCIDRs = []string{"192.0.2.9"}, []string{"bad"}
			case "any":
				rule.SrcCIDRs, rule.AnyCIDRs = []string{"192.0.2.9"}, []string{"bad"}
			}
			if field != "missing" {
				want = fmt.Sprintf("suppression scope %s cidrs: invalid IP or CIDR %q", field, "bad")
			}
			if got, err := alert.NewSuppressor([]alert.Suppression{rule}); got != nil || err == nil || err.Error() != want {
				t.Fatalf("NewSuppressor = %+v, %v; want %q", got, err, want)
			}
			if got, err := alert.NewSuppressionManager([]alert.Suppression{rule}); got != nil || err == nil || err.Error() != want {
				t.Fatalf("NewSuppressionManager = %+v, %v; want %q", got, err, want)
			}
		})
	}
}

func TestFileBackedSuppressionRuleScopeRejectionPreservesStateAndPermitsRetry(t *testing.T) {
	for _, operation := range []string{"add", "update", "reload"} {
		for _, ids := range [][]string{{""}, {"", ""}} {
			t.Run(fmt.Sprintf("%s/ids=%d", operation, len(ids)), func(t *testing.T) {
				path := emptySuppressionPath(t)
				initial := []alert.Suppression{{ID: "original", Enabled: true, RuleIDs: []string{"old-rule"}, AnyCIDRs: []string{"192.0.2.0/24"}}}
				writeEmptySuppressionFile(t, path, initial)
				manager, err := alert.NewSuppressionManagerWithFile(initial, path)
				if err != nil {
					t.Fatal(err)
				}
				old := &model.Alert{RuleID: "old-rule", SrcIP: "192.0.2.9", DstIP: "198.51.100.7"}
				newHit := &model.Alert{RuleID: "new-rule", SrcIP: "203.0.113.9", DstIP: "198.51.100.7"}
				other := &model.Alert{RuleID: "unrelated-rule", SrcIP: "203.0.113.9", DstIP: "192.0.2.9"}
				probes := []*model.Alert{nil, old, newHit, other}
				probeBefore := []model.Alert{*old, *newHit, *other}
				listBefore := manager.List()
				wantBefore := []*model.Alert{nil, newHit, other}
				if !reflect.DeepEqual(manager.Filter(probes), wantBefore) {
					t.Fatal("initial filter fixture invalid")
				}
				candidate := alert.Suppression{ID: "candidate", Enabled: true, RuleIDs: ids, AnyCIDRs: []string{"203.0.113.0/24"}}
				candidateBefore := copyEmptySuppressionRules([]alert.Suppression{candidate})[0]
				wantID := candidate.ID
				if operation == "reload" {
					writeEmptySuppressionFile(t, path, []alert.Suppression{candidate})
					loaded, err := alert.LoadSuppressionsFromFile(path)
					if err != nil || !reflect.DeepEqual(loaded, []alert.Suppression{candidate}) {
						t.Fatalf("structural load = %+v, %v", loaded, err)
					}
				}
				bytesBefore := readEmptySuppressionFile(t, path)
				modeBefore, err := os.Stat(path)
				if err != nil {
					t.Fatal(err)
				}
				entriesBefore, err := os.ReadDir(filepath.Dir(path))
				if err != nil {
					t.Fatal(err)
				}
				switch operation {
				case "add":
					err = manager.Add(candidate)
				case "update":
					wantID = "original"
					err = manager.Update(wantID, candidate)
				case "reload":
					err = manager.ReloadFromFile()
				}
				want := fmt.Sprintf("suppression %q rule_ids must include at least one nonempty rule ID", wantID)
				if err == nil || err.Error() != want {
					t.Fatalf("rejection = %v, want %q", err, want)
				}
				modeAfter, err := os.Stat(path)
				if err != nil {
					t.Fatal(err)
				}
				entriesAfter, err := os.ReadDir(filepath.Dir(path))
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(readEmptySuppressionFile(t, path), bytesBefore) || modeAfter.Mode() != modeBefore.Mode() || !reflect.DeepEqual(entryNames(entriesAfter), entryNames(entriesBefore)) || !reflect.DeepEqual(manager.List(), listBefore) || !reflect.DeepEqual(manager.Filter(probes), wantBefore) {
					t.Fatal("rejection changed file bytes/mode/membership or published list/filter")
				}
				if !reflect.DeepEqual(candidate, candidateBefore) {
					t.Fatal("rejection changed caller candidate")
				}
				valid := candidate
				valid.RuleIDs = []string{"", "new-rule", "new-rule", ""}
				validBefore := copyEmptySuppressionRules([]alert.Suppression{valid})[0]
				wantRules := []alert.Suppression{valid}
				wantFiltered := []*model.Alert{nil, old, other}
				var reloadBytes []byte
				switch operation {
				case "add":
					err = manager.Add(valid)
					wantRules = append(copyEmptySuppressionRules(initial), valid)
					wantFiltered = []*model.Alert{nil, other}
				case "update":
					err = manager.Update("original", valid)
					wantRules[0].ID = "original"
				case "reload":
					reloadBytes = writeEmptySuppressionFile(t, path, wantRules)
					err = manager.ReloadFromFile()
				}
				if err != nil || !reflect.DeepEqual(manager.List(), wantRules) || !reflect.DeepEqual(manager.Filter(probes), wantFiltered) {
					t.Fatalf("retry = %v; list=%+v filter=%+v", err, manager.List(), manager.Filter(probes))
				}
				loaded, err := alert.LoadSuppressionsFromFile(path)
				if err != nil || !reflect.DeepEqual(loaded, wantRules) || !reflect.DeepEqual(valid, validBefore) {
					t.Fatal("retry file/caller identity changed")
				}
				if operation == "reload" && !bytes.Equal(readEmptySuppressionFile(t, path), reloadBytes) {
					t.Fatal("reload rewrote supplied file")
				}
				if *old != probeBefore[0] || *newHit != probeBefore[1] || *other != probeBefore[2] || probes[0] != nil || probes[1] != old || probes[2] != newHit || probes[3] != other {
					t.Fatal("mutation changed caller alerts/slice")
				}
			})
		}
	}
}

func entryNames(entries []os.DirEntry) []string {
	names := make([]string, len(entries))
	for i, entry := range entries {
		names[i] = entry.Name()
	}
	return names
}
