package alert

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/decline-llc/netsentry/pkg/model"
)

func TestSuppressionReloadOwnsExclusiveLockAtAuthoritativeRead(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suppressions.json")
	rules := []Suppression{{ID: "loaded", Enabled: true, AnyCIDRs: []string{"192.0.2.0/24"}}}
	if err := SaveSuppressionsToFile(path, rules); err != nil {
		t.Fatal(err)
	}
	manager, err := NewSuppressionManagerWithFile(nil, path)
	if err != nil {
		t.Fatal(err)
	}
	calls := 0
	manager.loadFromFile = func(gotPath string) ([]Suppression, error) {
		calls++
		if gotPath != path {
			return nil, fmt.Errorf("reload path = %q, want %q", gotPath, path)
		}
		if err := suppressionReloadExclusiveRead(manager); err != nil {
			return nil, err
		}
		return LoadSuppressionsFromFile(gotPath)
	}
	if err := manager.ReloadFromFile(); err != nil {
		t.Fatal(err)
	}
	if calls != 1 {
		t.Fatalf("authoritative loads = %d, want 1", calls)
	}
	assertSuppressionReloadState(t, manager, path, rules)
}

func TestSuppressionReloadOverlapsMutationsWithoutLosingCommittedState(t *testing.T) {
	old := Suppression{ID: "target", Enabled: true, AnyCIDRs: []string{"10.0.0.0/24"}}
	loaded := Suppression{ID: "target", Enabled: true, AnyCIDRs: []string{"198.51.100.0/24"}}
	added := Suppression{ID: "added", Enabled: true, AnyCIDRs: []string{"192.0.2.0/24"}}
	updated := Suppression{ID: "target", Enabled: true, AnyCIDRs: []string{"203.0.113.0/24"}}
	cases := []struct {
		name   string
		mutate func(*SuppressionManager) error
		want   []Suppression
	}{
		{"add", func(m *SuppressionManager) error { return m.Add(added) }, []Suppression{loaded, added}},
		{"update", func(m *SuppressionManager) error { return m.Update("target", updated) }, []Suppression{updated}},
		{"delete", func(m *SuppressionManager) error { return m.Delete("target") }, []Suppression{}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "suppression state", "suppressions.json")
			if err := SaveSuppressionsToFile(path, []Suppression{loaded}); err != nil {
				t.Fatal(err)
			}
			manager, err := NewSuppressionManagerWithFile([]Suppression{old}, path)
			if err != nil {
				t.Fatal(err)
			}
			entered, release := make(chan struct{}), make(chan struct{})
			var once sync.Once
			unblock := func() { once.Do(func() { close(release) }) }
			t.Cleanup(unblock)
			manager.loadFromFile = func(path string) ([]Suppression, error) {
				if err := suppressionReloadExclusiveRead(manager); err != nil {
					return nil, err
				}
				rules, err := LoadSuppressionsFromFile(path)
				if err != nil {
					return nil, err
				}
				close(entered)
				<-release // Hold the real disk snapshot until mutation has started.
				return rules, nil
			}
			reloadDone := make(chan error, 1)
			go func() { reloadDone <- manager.ReloadFromFile() }()
			select {
			case <-entered:
			case err := <-reloadDone:
				t.Fatalf("reload returned before its read boundary: %v", err)
			case <-time.After(2 * time.Second):
				t.Fatal("reload did not reach its read boundary")
			}
			started, mutationDone := make(chan struct{}), make(chan error, 1)
			go func() {
				close(started)
				mutationDone <- tc.mutate(manager)
			}()
			<-started
			unblock()
			awaitSuppressionReloadOperation(t, reloadDone, "reload")
			awaitSuppressionReloadOperation(t, mutationDone, tc.name)
			assertSuppressionReloadState(t, manager, path, tc.want)
		})
	}
}

func TestSuppressionReloadErrorsPreserveStateReleaseLockAndAllowMutation(t *testing.T) {
	readErr := errors.New("injected suppression load failure")
	initial := []Suppression{{ID: "old", Enabled: true, AnyCIDRs: []string{"10.0.0.0/24"}}}
	cases := []struct {
		name, input, diagnostic string
		injected                error
	}{
		{"read", `{"suppressions":[]}`, "", readErr},
		{"parse", `{"suppressions":`, "parse suppressions", nil},
		{"set", `{"suppressions":[{"id":"","enabled":true,"any_cidrs":["192.0.2.0/24"]}]}`, "validate suppressions", nil},
		{"compile", `{"suppressions":[{"id":"bad","enabled":true,"any_cidrs":["bad-cidr"]}]}`, "invalid IP or CIDR", nil},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			path := filepath.Join(t.TempDir(), "suppressions.json")
			if err := os.WriteFile(path, []byte(tc.input), 0o600); err != nil {
				t.Fatal(err)
			}
			manager, err := NewSuppressionManagerWithFile(initial, path)
			if err != nil {
				t.Fatal(err)
			}
			manager.loadFromFile = func(path string) ([]Suppression, error) {
				if err := suppressionReloadExclusiveRead(manager); err != nil {
					return nil, err
				}
				if tc.injected != nil {
					return nil, tc.injected
				}
				return LoadSuppressionsFromFile(path)
			}
			err = manager.ReloadFromFile()
			if err == nil || (tc.injected != nil && !errors.Is(err, tc.injected)) || !strings.Contains(err.Error(), tc.diagnostic) {
				t.Fatalf("reload error = %v, want sentinel/diagnostic %q", err, tc.diagnostic)
			}
			if !manager.mu.TryLock() {
				t.Fatal("reload failure retained the mutation lock")
			}
			manager.mu.Unlock()
			if got := manager.List(); !reflect.DeepEqual(got, initial) {
				t.Fatalf("failed reload changed active rules: %+v", got)
			}
			before, err := os.ReadFile(path)
			if err != nil || string(before) != tc.input {
				t.Fatalf("failed reload changed canonical bytes: %q, %v", before, err)
			}
			if len(manager.Filter([]*model.Alert{alertForSuppression("rule", "10.0.0.1", "198.51.100.1")})) != 0 {
				t.Fatal("failed reload changed the old active filter")
			}
			manager.loadFromFile = LoadSuppressionsFromFile
			added := Suppression{ID: "new", Enabled: true, AnyCIDRs: []string{"192.0.2.0/24"}}
			if err := manager.Add(added); err != nil {
				t.Fatalf("mutation after failed reload: %v", err)
			}
			assertSuppressionReloadState(t, manager, path, append(cloneSuppressions(initial), added))
		})
	}
}

func TestSuppressionReloadGuardsAndMissingFile(t *testing.T) {
	var absent *SuppressionManager
	if err := absent.ReloadFromFile(); err == nil || err.Error() != "suppression manager is not configured" {
		t.Fatalf("nil-manager error = %v", err)
	}
	manager, err := NewSuppressionManager(nil)
	if err != nil {
		t.Fatal(err)
	}
	manager.loadFromFile = func(string) ([]Suppression, error) {
		t.Fatal("unconfigured reload invoked its loader")
		return nil, nil
	}
	if err := manager.ReloadFromFile(); err == nil || err.Error() != "suppressions file is not configured" {
		t.Fatalf("unconfigured error = %v", err)
	}
	path := filepath.Join(t.TempDir(), "missing.json")
	manager, err = NewSuppressionManagerWithFile([]Suppression{{ID: "old", Enabled: true, AnyCIDRs: []string{"10.0.0.0/24"}}}, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.ReloadFromFile(); err != nil {
		t.Fatal(err)
	}
	if len(manager.List()) != 0 || len(manager.Filter([]*model.Alert{alertForSuppression("rule", "10.0.0.1", "198.51.100.1")})) != 1 {
		t.Fatal("missing-file reload did not clear rules and filter")
	}
	if _, err := os.Stat(path); !errors.Is(err, os.ErrNotExist) {
		t.Fatalf("missing-file reload created or modified file: %v", err)
	}
}

func TestSuppressionReloadRetainsDefensiveListSlices(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suppressions.json")
	rules := []Suppression{{ID: "owned", Enabled: true, RuleIDs: []string{"rule"}, SrcCIDRs: []string{"192.0.2.0/24"}, DstCIDRs: []string{"203.0.113.0/24"}, AnyCIDRs: []string{"198.51.100.0/24"}}}
	if err := SaveSuppressionsToFile(path, rules); err != nil {
		t.Fatal(err)
	}
	manager, err := NewSuppressionManagerWithFile(nil, path)
	if err != nil {
		t.Fatal(err)
	}
	if err := manager.ReloadFromFile(); err != nil {
		t.Fatal(err)
	}
	listed := manager.List()
	listed[0].RuleIDs[0] = "changed"
	listed[0].SrcCIDRs[0] = "10.0.0.0/24"
	listed[0].DstCIDRs[0] = "10.0.0.0/24"
	listed[0].AnyCIDRs[0] = "10.0.0.0/24"
	listed[0].ID, listed[0].Enabled = "changed", false
	assertSuppressionReloadState(t, manager, path, rules)
}

// This seam is called by public ReloadFromFile before the real file read.
// With no other reader yet started, TryRLock detects a missing exclusive lock.
func suppressionReloadExclusiveRead(manager *SuppressionManager) error {
	if manager.mu.TryRLock() {
		manager.mu.RUnlock()
		return errors.New("authoritative reload read was not exclusively locked")
	}
	return nil
}

func awaitSuppressionReloadOperation(t *testing.T, done <-chan error, operation string) {
	t.Helper()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("%s: %v", operation, err)
		}
	case <-time.After(2 * time.Second):
		t.Fatalf("%s did not finish", operation)
	}
}

func assertSuppressionReloadState(t *testing.T, manager *SuppressionManager, path string, want []Suppression) {
	t.Helper()
	listed := manager.List()
	persisted, err := LoadSuppressionsFromFile(path)
	if err != nil || !reflect.DeepEqual(listed, want) || !reflect.DeepEqual(persisted, want) {
		t.Fatalf("final state: active=%+v disk=%+v want=%+v error=%v", listed, persisted, want, err)
	}
	raw, err := os.ReadFile(path)
	if err != nil || string(raw) != string(canonicalSuppressionBytes(t, want)) {
		t.Fatalf("canonical bytes disagree with committed state: %q, %v", raw, err)
	}
	suppressor, err := NewSuppressor(want)
	if err != nil {
		t.Fatal(err)
	}
	for _, ruleID := range []string{"rule", "other"} {
		for _, addr := range []string{"10.0.0.1", "192.0.2.1", "198.51.100.1", "203.0.113.1"} {
			for _, sample := range []*model.Alert{alertForSuppression(ruleID, addr, "100.64.0.1"), alertForSuppression(ruleID, "100.64.0.1", addr)} {
				filtered := manager.Filter([]*model.Alert{sample})
				wantKept := !suppressor.Suppressed(sample)
				if (len(filtered) == 1) != wantKept || (wantKept && filtered[0] != sample) {
					t.Fatalf("final filter disagrees for %+v: kept=%v want=%v", sample, len(filtered), wantKept)
				}
			}
		}
	}
}
