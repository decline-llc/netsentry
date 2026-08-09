package alert

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/decline-llc/netsentry/pkg/model"
)

func TestSuppressorMatchesSourceDestinationAndAnyCIDRs(t *testing.T) {
	suppressor, err := NewSuppressor([]Suppression{
		{ID: "src", Enabled: true, RuleIDs: []string{"rule-src"}, SrcCIDRs: []string{"10.0.0.0/24"}},
		{ID: "dst", Enabled: true, RuleIDs: []string{"rule-dst"}, DstCIDRs: []string{"192.0.2.10"}},
		{ID: "any", Enabled: true, AnyCIDRs: []string{"203.0.113.0/24"}},
	})
	if err != nil {
		t.Fatalf("new suppressor: %v", err)
	}

	cases := []struct {
		name  string
		alert *model.Alert
	}{
		{name: "source", alert: alertForSuppression("rule-src", "10.0.0.55", "198.51.100.1")},
		{name: "destination", alert: alertForSuppression("rule-dst", "198.51.100.1", "192.0.2.10")},
		{name: "any source", alert: alertForSuppression("rule-other", "203.0.113.20", "198.51.100.1")},
		{name: "any destination", alert: alertForSuppression("rule-other", "198.51.100.1", "203.0.113.20")},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if !suppressor.Suppressed(tc.alert) {
				t.Fatalf("expected alert to be suppressed: %+v", tc.alert)
			}
		})
	}
}

func TestSuppressorHonorsRuleIDAndEnabled(t *testing.T) {
	suppressor, err := NewSuppressor([]Suppression{
		{ID: "disabled", Enabled: false, SrcCIDRs: []string{"10.0.0.0/24"}},
		{ID: "scoped", Enabled: true, RuleIDs: []string{"rule-1"}, SrcCIDRs: []string{"10.0.0.0/24"}},
	})
	if err != nil {
		t.Fatalf("new suppressor: %v", err)
	}
	if suppressor.Suppressed(alertForSuppression("rule-2", "10.0.0.5", "198.51.100.1")) {
		t.Fatal("expected different rule ID to avoid suppression")
	}
	if !suppressor.Suppressed(alertForSuppression("rule-1", "10.0.0.5", "198.51.100.1")) {
		t.Fatal("expected matching rule ID to suppress alert")
	}
}

func TestSuppressorFilter(t *testing.T) {
	suppressor, err := NewSuppressor([]Suppression{{ID: "src", Enabled: true, SrcCIDRs: []string{"10.0.0.0/24"}}})
	if err != nil {
		t.Fatalf("new suppressor: %v", err)
	}
	alerts := []*model.Alert{
		alertForSuppression("rule-1", "10.0.0.5", "198.51.100.1"),
		alertForSuppression("rule-2", "198.51.100.2", "198.51.100.1"),
	}
	filtered := suppressor.Filter(alerts)
	if len(filtered) != 1 || filtered[0].RuleID != "rule-2" {
		t.Fatalf("unexpected filtered alerts: %+v", filtered)
	}
}

func TestSuppressorRejectsInvalidCIDR(t *testing.T) {
	if _, err := NewSuppressor([]Suppression{{ID: "bad", Enabled: true, SrcCIDRs: []string{"not-a-cidr"}}}); err == nil {
		t.Fatal("expected invalid CIDR error")
	}
}

func TestSuppressionManagerAddsAndFilters(t *testing.T) {
	manager, err := NewSuppressionManager(nil)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	if err := manager.Add(Suppression{ID: "src", Enabled: true, SrcCIDRs: []string{"10.0.0.0/24"}}); err != nil {
		t.Fatalf("add suppression: %v", err)
	}
	rules := manager.List()
	if len(rules) != 1 || rules[0].ID != "src" {
		t.Fatalf("unexpected rules: %+v", rules)
	}
	filtered := manager.Filter([]*model.Alert{
		alertForSuppression("rule-1", "10.0.0.10", "198.51.100.1"),
		alertForSuppression("rule-2", "198.51.100.2", "198.51.100.1"),
	})
	if len(filtered) != 1 || filtered[0].RuleID != "rule-2" {
		t.Fatalf("unexpected filtered alerts: %+v", filtered)
	}
}

func TestSuppressionManagerRejectsDuplicateID(t *testing.T) {
	manager, err := NewSuppressionManager([]Suppression{{ID: "dup", Enabled: true, AnyCIDRs: []string{"10.0.0.0/24"}}})
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	if err := manager.Add(Suppression{ID: "dup", Enabled: true, AnyCIDRs: []string{"192.0.2.0/24"}}); err == nil {
		t.Fatal("expected duplicate suppression ID error")
	}
}

func TestSuppressionManagerRejectsEnabledRuleWithoutCIDR(t *testing.T) {
	if _, err := NewSuppressionManager([]Suppression{{ID: "empty", Enabled: true}}); err == nil {
		t.Fatal("expected missing CIDR error")
	}
}

func TestLoadAndSaveSuppressionsFile(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suppressions.json")
	rules := []Suppression{{ID: "src", Enabled: true, RuleIDs: []string{"rule-1"}, SrcCIDRs: []string{"10.0.0.0/24"}}}
	if err := SaveSuppressionsToFile(path, rules); err != nil {
		t.Fatalf("save suppressions: %v", err)
	}
	loaded, err := LoadSuppressionsFromFile(path)
	if err != nil {
		t.Fatalf("load suppressions: %v", err)
	}
	if len(loaded) != 1 || loaded[0].ID != "src" || loaded[0].SrcCIDRs[0] != "10.0.0.0/24" {
		t.Fatalf("unexpected suppressions: %+v", loaded)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read suppressions file: %v", err)
	}
	if !strings.Contains(string(raw), `"suppressions"`) {
		t.Fatalf("expected canonical wrapper, got %s", string(raw))
	}
}

func TestRepositorySuppressionsFileUsesCanonicalWrappedSchema(t *testing.T) {
	path := filepath.Join("..", "..", "..", "configs", "suppressions.json")
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	var wrapped struct {
		Suppressions []Suppression `json:"suppressions"`
	}
	if err := json.Unmarshal(raw, &wrapped); err != nil {
		t.Fatalf("decode repository suppressions: %v", err)
	}
	if wrapped.Suppressions == nil {
		t.Fatal("repository suppressions file must use canonical suppressions wrapper")
	}
	loaded, err := LoadSuppressionsFromFile(path)
	if err != nil {
		t.Fatalf("load repository suppressions: %v", err)
	}
	if _, err := NewSuppressionManager(loaded); err != nil {
		t.Fatalf("compile repository suppressions: %v", err)
	}
}

func TestLoadSuppressionsFromMissingFileReturnsEmpty(t *testing.T) {
	loaded, err := LoadSuppressionsFromFile(filepath.Join(t.TempDir(), "missing.json"))
	if err != nil {
		t.Fatalf("load missing suppressions: %v", err)
	}
	if len(loaded) != 0 {
		t.Fatalf("expected empty suppressions, got %+v", loaded)
	}
}

func TestSaveSuppressionsCreatesMissingParentDirectory(t *testing.T) {
	path := filepath.Join(t.TempDir(), "nested", "suppression state", "suppressions.json")
	rules := []Suppression{{ID: "nested", Enabled: true, AnyCIDRs: []string{"192.0.2.0/24"}}}
	if err := SaveSuppressionsToFile(path, rules); err != nil {
		t.Fatalf("save suppressions: %v", err)
	}
	loaded, err := LoadSuppressionsFromFile(path)
	if err != nil {
		t.Fatalf("load suppressions: %v", err)
	}
	if len(loaded) != 1 || loaded[0].ID != "nested" {
		t.Fatalf("unexpected suppressions: %+v", loaded)
	}
}

func TestPersistentSuppressionManagerWritesOnAdd(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suppressions.json")
	manager, err := NewSuppressionManagerWithFile(nil, path)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	if err := manager.Add(Suppression{ID: "persisted", Enabled: true, AnyCIDRs: []string{"10.0.0.0/24"}}); err != nil {
		t.Fatalf("add suppression: %v", err)
	}
	loaded, err := LoadSuppressionsFromFile(path)
	if err != nil {
		t.Fatalf("load suppressions: %v", err)
	}
	if len(loaded) != 1 || loaded[0].ID != "persisted" {
		t.Fatalf("unexpected persisted suppressions: %+v", loaded)
	}
}

func TestSuppressionManagerUpdateDeleteAndReload(t *testing.T) {
	path := filepath.Join(t.TempDir(), "suppressions.json")
	initial := []Suppression{{ID: "s1", Enabled: true, SrcCIDRs: []string{"10.0.0.0/24"}}}
	if err := SaveSuppressionsToFile(path, initial); err != nil {
		t.Fatalf("save initial suppressions: %v", err)
	}
	manager, err := NewSuppressionManagerWithFile(initial, path)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}

	if err := manager.Update("s1", Suppression{ID: "s1", Enabled: true, DstCIDRs: []string{"192.0.2.0/24"}}); err != nil {
		t.Fatalf("update suppression: %v", err)
	}
	if !manager.suppressor.Suppressed(alertForSuppression("rule-1", "198.51.100.1", "192.0.2.10")) {
		t.Fatal("expected updated destination suppression to be active")
	}

	loaded, err := LoadSuppressionsFromFile(path)
	if err != nil {
		t.Fatalf("load updated suppressions: %v", err)
	}
	if len(loaded) != 1 || len(loaded[0].DstCIDRs) != 1 || loaded[0].DstCIDRs[0] != "192.0.2.0/24" {
		t.Fatalf("unexpected updated file suppressions: %+v", loaded)
	}

	if err := manager.Delete("s1"); err != nil {
		t.Fatalf("delete suppression: %v", err)
	}
	if listed := manager.List(); len(listed) != 0 {
		t.Fatalf("expected delete to clear suppression, got %+v", listed)
	}

	reloadedRules := []Suppression{{ID: "disk", Enabled: true, AnyCIDRs: []string{"203.0.113.0/24"}}}
	if err := SaveSuppressionsToFile(path, reloadedRules); err != nil {
		t.Fatalf("save reload suppressions: %v", err)
	}
	if err := manager.ReloadFromFile(); err != nil {
		t.Fatalf("reload suppressions: %v", err)
	}
	listed := manager.List()
	if len(listed) != 1 || listed[0].ID != "disk" {
		t.Fatalf("unexpected reloaded suppressions: %+v", listed)
	}
}

func TestSuppressionManagerUpdateDeleteMissingID(t *testing.T) {
	manager, err := NewSuppressionManager(nil)
	if err != nil {
		t.Fatalf("new manager: %v", err)
	}
	if err := manager.Update("missing", Suppression{Enabled: true, AnyCIDRs: []string{"10.0.0.0/24"}}); err == nil {
		t.Fatal("expected missing update error")
	}
	if err := manager.Delete("missing"); err == nil {
		t.Fatal("expected missing delete error")
	}
}

var errInjectedSuppressionReplacement = errors.New("injected suppression replacement failure")

type faultSuppressionReplacementFile struct {
	suppressionReplacementFile
	phase string
}

func (f *faultSuppressionReplacementFile) Write(data []byte) (int, error) {
	switch f.phase {
	case "short-write":
		if len(data) == 0 {
			return 0, nil
		}
		n, err := f.suppressionReplacementFile.Write(data[:len(data)-1])
		if err != nil {
			return n, err
		}
		return n, nil
	case "write":
		return 0, errInjectedSuppressionReplacement
	default:
		return f.suppressionReplacementFile.Write(data)
	}
}

func (f *faultSuppressionReplacementFile) Chmod(mode os.FileMode) error {
	if f.phase == "chmod" {
		return errInjectedSuppressionReplacement
	}
	return f.suppressionReplacementFile.Chmod(mode)
}

func (f *faultSuppressionReplacementFile) Sync() error {
	if f.phase == "file-sync" {
		return errInjectedSuppressionReplacement
	}
	return f.suppressionReplacementFile.Sync()
}

func (f *faultSuppressionReplacementFile) Close() error {
	err := f.suppressionReplacementFile.Close()
	if f.phase == "file-close" {
		return errInjectedSuppressionReplacement
	}
	return err
}

type faultSuppressionReplacementDirectory struct {
	suppressionReplacementDirectory
	phase string
}

func (d *faultSuppressionReplacementDirectory) Sync() error {
	if d.phase == "directory-sync" {
		return errInjectedSuppressionReplacement
	}
	return d.suppressionReplacementDirectory.Sync()
}

func (d *faultSuppressionReplacementDirectory) Close() error {
	err := d.suppressionReplacementDirectory.Close()
	if d.phase == "directory-close" {
		return errInjectedSuppressionReplacement
	}
	return err
}

func suppressionReplacementOpsWithFault(phase string) suppressionReplacementOps {
	ops := osSuppressionReplacementOps
	baseStat := ops.stat
	ops.stat = func(path string) (os.FileInfo, error) {
		if phase == "stat" {
			return nil, errInjectedSuppressionReplacement
		}
		return baseStat(path)
	}
	baseMkdirAll := ops.mkdirAll
	ops.mkdirAll = func(path string, mode os.FileMode) error {
		if phase == "directory-create" {
			return errInjectedSuppressionReplacement
		}
		return baseMkdirAll(path, mode)
	}
	baseCreateTemp := ops.createTemp
	ops.createTemp = func(dir, pattern string) (suppressionReplacementFile, error) {
		if phase == "create" {
			return nil, errInjectedSuppressionReplacement
		}
		file, err := baseCreateTemp(dir, pattern)
		if err != nil {
			return nil, err
		}
		return &faultSuppressionReplacementFile{suppressionReplacementFile: file, phase: phase}, nil
	}
	baseRename := ops.rename
	ops.rename = func(oldPath, newPath string) error {
		if phase == "rename" {
			return errInjectedSuppressionReplacement
		}
		return baseRename(oldPath, newPath)
	}
	baseOpenDir := ops.openDir
	ops.openDir = func(path string) (suppressionReplacementDirectory, error) {
		if phase == "directory-open" {
			return nil, errInjectedSuppressionReplacement
		}
		dir, err := baseOpenDir(path)
		if err != nil {
			return nil, err
		}
		return &faultSuppressionReplacementDirectory{suppressionReplacementDirectory: dir, phase: phase}, nil
	}
	return ops
}

func TestSaveSuppressionsReplacementLifecyclePreservesPriorFileBeforeRename(t *testing.T) {
	for _, phase := range []string{
		"stat",
		"directory-create",
		"create",
		"short-write",
		"write",
		"chmod",
		"file-sync",
		"file-close",
		"rename",
	} {
		t.Run(phase, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "suppressions.json")
			before := []byte("{\"suppressions\":[]}\n")
			if err := os.WriteFile(path, before, 0o600); err != nil {
				t.Fatal(err)
			}

			rules := []Suppression{{ID: "new", Enabled: true, AnyCIDRs: []string{"192.0.2.0/24"}}}
			err := saveSuppressionsToFile(path, rules, suppressionReplacementOpsWithFault(phase))
			wantErr := errInjectedSuppressionReplacement
			if phase == "short-write" {
				wantErr = io.ErrShortWrite
			}
			if !errors.Is(err, wantErr) {
				t.Fatalf("error = %v, want %v", err, wantErr)
			}
			if SuppressionReplacementCommitted(err) {
				t.Fatalf("phase %s incorrectly reported a committed replacement: %v", phase, err)
			}
			after, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if !slices.Equal(after, before) {
				t.Fatalf("phase %s changed prior file: got %q want %q", phase, after, before)
			}
			assertNoSuppressionTemporaryFiles(t, dir)
		})
	}
}

func TestSaveSuppressionsPostRenameFailuresReportCommittedReplacement(t *testing.T) {
	for _, phase := range []string{"directory-open", "directory-sync", "directory-close"} {
		t.Run(phase, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "suppressions.json")
			if err := os.WriteFile(path, []byte("{\"suppressions\":[]}\n"), 0o600); err != nil {
				t.Fatal(err)
			}
			rules := []Suppression{{ID: "committed", Enabled: true, AnyCIDRs: []string{"192.0.2.0/24"}}}
			want := canonicalSuppressionBytes(t, rules)

			err := saveSuppressionsToFile(path, rules, suppressionReplacementOpsWithFault(phase))
			if !errors.Is(err, errInjectedSuppressionReplacement) || !SuppressionReplacementCommitted(err) {
				t.Fatalf("error = %v, want committed injected failure", err)
			}
			after, readErr := os.ReadFile(path)
			if readErr != nil {
				t.Fatal(readErr)
			}
			if !slices.Equal(after, want) {
				t.Fatalf("phase %s canonical bytes = %q, want %q", phase, after, want)
			}
			assertNoSuppressionTemporaryFiles(t, dir)
		})
	}
}

type recordingSuppressionReplacementFile struct {
	suppressionReplacementFile
	events *[]string
}

func (f *recordingSuppressionReplacementFile) Write(data []byte) (int, error) {
	*f.events = append(*f.events, "write")
	return f.suppressionReplacementFile.Write(data)
}

func (f *recordingSuppressionReplacementFile) Chmod(mode os.FileMode) error {
	*f.events = append(*f.events, "chmod")
	return f.suppressionReplacementFile.Chmod(mode)
}

func (f *recordingSuppressionReplacementFile) Sync() error {
	*f.events = append(*f.events, "file-sync")
	return f.suppressionReplacementFile.Sync()
}

func (f *recordingSuppressionReplacementFile) Close() error {
	*f.events = append(*f.events, "file-close")
	return f.suppressionReplacementFile.Close()
}

type recordingSuppressionReplacementDirectory struct {
	suppressionReplacementDirectory
	events *[]string
}

func (d *recordingSuppressionReplacementDirectory) Sync() error {
	*d.events = append(*d.events, "directory-sync")
	return d.suppressionReplacementDirectory.Sync()
}

func (d *recordingSuppressionReplacementDirectory) Close() error {
	*d.events = append(*d.events, "directory-close")
	return d.suppressionReplacementDirectory.Close()
}

func TestSaveSuppressionsRunsCompleteDurabilityLifecycleInOrder(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "suppressions.json")
	if err := os.WriteFile(path, []byte("{\"suppressions\":[]}\n"), 0o600); err != nil {
		t.Fatal(err)
	}

	events := make([]string, 0, 11)
	ops := osSuppressionReplacementOps
	baseStat := ops.stat
	ops.stat = func(path string) (os.FileInfo, error) {
		events = append(events, "stat")
		return baseStat(path)
	}
	baseMkdirAll := ops.mkdirAll
	ops.mkdirAll = func(path string, mode os.FileMode) error {
		events = append(events, "directory-create")
		return baseMkdirAll(path, mode)
	}
	baseCreateTemp := ops.createTemp
	ops.createTemp = func(dir, pattern string) (suppressionReplacementFile, error) {
		events = append(events, "create")
		file, err := baseCreateTemp(dir, pattern)
		if err != nil {
			return nil, err
		}
		return &recordingSuppressionReplacementFile{suppressionReplacementFile: file, events: &events}, nil
	}
	baseRename := ops.rename
	ops.rename = func(oldPath, newPath string) error {
		events = append(events, "rename")
		return baseRename(oldPath, newPath)
	}
	baseOpenDir := ops.openDir
	ops.openDir = func(path string) (suppressionReplacementDirectory, error) {
		events = append(events, "directory-open")
		dir, err := baseOpenDir(path)
		if err != nil {
			return nil, err
		}
		return &recordingSuppressionReplacementDirectory{suppressionReplacementDirectory: dir, events: &events}, nil
	}

	rules := []Suppression{{ID: "saved", Enabled: true, AnyCIDRs: []string{"192.0.2.0/24"}}}
	if err := saveSuppressionsToFile(path, rules, ops); err != nil {
		t.Fatal(err)
	}
	wantEvents := []string{
		"stat",
		"directory-create",
		"create",
		"write",
		"chmod",
		"file-sync",
		"file-close",
		"rename",
		"directory-open",
		"directory-sync",
		"directory-close",
	}
	if !slices.Equal(events, wantEvents) {
		t.Fatalf("events = %v, want %v", events, wantEvents)
	}
	info, err := os.Stat(path)
	if err != nil {
		t.Fatal(err)
	}
	if got := info.Mode().Perm(); got != 0o600 {
		t.Fatalf("mode = %v, want 0600", got)
	}
	assertNoSuppressionTemporaryFiles(t, dir)
}

func TestSuppressionManagerPreRenameFailurePreservesStateAndAllowsRetry(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "suppressions.json")
	initial := []Suppression{{ID: "old", Enabled: true, AnyCIDRs: []string{"10.0.0.0/24"}}}
	if err := SaveSuppressionsToFile(path, initial); err != nil {
		t.Fatal(err)
	}
	before, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	manager, err := NewSuppressionManagerWithFile(initial, path)
	if err != nil {
		t.Fatal(err)
	}
	manager.replacementOps = suppressionReplacementOpsWithFault("file-sync")
	candidate := Suppression{ID: "new", Enabled: true, AnyCIDRs: []string{"192.0.2.0/24"}}

	err = manager.Add(candidate)
	if !errors.Is(err, errInjectedSuppressionReplacement) || SuppressionReplacementCommitted(err) {
		t.Fatalf("error = %v, want uncommitted injected failure", err)
	}
	after, readErr := os.ReadFile(path)
	if readErr != nil {
		t.Fatal(readErr)
	}
	if !slices.Equal(after, before) {
		t.Fatalf("pre-rename failure changed canonical bytes: got %q want %q", after, before)
	}
	if listed := manager.List(); len(listed) != 1 || listed[0].ID != "old" {
		t.Fatalf("pre-rename failure changed active rules: %+v", listed)
	}
	if manager.suppressor.Suppressed(alertForSuppression("rule", "192.0.2.1", "198.51.100.1")) {
		t.Fatal("pre-rename failure changed active filter")
	}
	assertNoSuppressionTemporaryFiles(t, dir)

	manager.replacementOps = osSuppressionReplacementOps
	if err := manager.Add(candidate); err != nil {
		t.Fatalf("retry add suppression: %v", err)
	}
	if listed := manager.List(); len(listed) != 2 || listed[1].ID != "new" {
		t.Fatalf("retry active rules = %+v", listed)
	}
}

func TestSuppressionManagerPostRenameFailurePublishesCommittedState(t *testing.T) {
	for _, phase := range []string{"directory-open", "directory-sync", "directory-close"} {
		t.Run(phase, func(t *testing.T) {
			dir := t.TempDir()
			path := filepath.Join(dir, "suppressions.json")
			initial := []Suppression{{ID: "old", Enabled: true, AnyCIDRs: []string{"10.0.0.0/24"}}}
			if err := SaveSuppressionsToFile(path, initial); err != nil {
				t.Fatal(err)
			}
			manager, err := NewSuppressionManagerWithFile(initial, path)
			if err != nil {
				t.Fatal(err)
			}
			manager.replacementOps = suppressionReplacementOpsWithFault(phase)
			candidate := Suppression{ID: "new", Enabled: true, AnyCIDRs: []string{"192.0.2.0/24"}}

			err = manager.Add(candidate)
			if !errors.Is(err, errInjectedSuppressionReplacement) || !SuppressionReplacementCommitted(err) {
				t.Fatalf("error = %v, want committed injected failure", err)
			}
			persisted, loadErr := LoadSuppressionsFromFile(path)
			if loadErr != nil {
				t.Fatal(loadErr)
			}
			listed := manager.List()
			if len(persisted) != 2 || len(listed) != 2 || persisted[1].ID != "new" || listed[1].ID != "new" {
				t.Fatalf("committed state disagrees: persisted=%+v active=%+v", persisted, listed)
			}
			if !manager.suppressor.Suppressed(alertForSuppression("rule", "192.0.2.1", "198.51.100.1")) {
				t.Fatal("committed candidate was not published to active filter")
			}
			assertNoSuppressionTemporaryFiles(t, dir)
		})
	}
}

func canonicalSuppressionBytes(t *testing.T, rules []Suppression) []byte {
	t.Helper()
	data, err := json.MarshalIndent(suppressionsFile{Suppressions: cloneSuppressions(rules)}, "", "  ")
	if err != nil {
		t.Fatal(err)
	}
	return append(data, '\n')
}

func assertNoSuppressionTemporaryFiles(t *testing.T, dir string) {
	t.Helper()
	matches, err := filepath.Glob(filepath.Join(dir, ".suppressions-*.json"))
	if err != nil {
		t.Fatal(err)
	}
	if len(matches) != 0 {
		t.Fatalf("temporary suppression files remain: %v", matches)
	}
}

func alertForSuppression(ruleID, srcIP, dstIP string) *model.Alert {
	return &model.Alert{RuleID: ruleID, SrcIP: srcIP, DstIP: dstIP}
}
