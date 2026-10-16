package alert_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/decline-llc/netsentry/internal/alert"
	"github.com/decline-llc/netsentry/pkg/model"
)

func TestSuppressionErrorCausesPreserveMessagesAndState(t *testing.T) {
	causes := []error{alert.ErrSuppressionNotFound, alert.ErrSuppressionAlreadyExists, alert.ErrSuppressionsFileNotConfigured}
	for _, id := range []string{"not found", "already exists", "not configured"} {
		t.Run(id, func(t *testing.T) {
			original := alert.Suppression{ID: id, Enabled: true, AnyCIDRs: []string{"192.0.2.0/24"}}
			input := []alert.Suppression{original}
			beforeInput, err := json.Marshal(input)
			if err != nil {
				t.Fatal(err)
			}
			manager, err := alert.NewSuppressionManager(input)
			if err != nil {
				t.Fatal(err)
			}
			probes := []*model.Alert{{SrcIP: "192.0.2.9"}, {SrcIP: "198.51.100.9"}}
			probeBefore := []model.Alert{*probes[0], *probes[1]}
			path := filepath.Join(t.TempDir(), "not found already exists not configured", "suppressions.json")
			assertCause := func(err, want error, message string) {
				t.Helper()
				if err == nil || err.Error() != message {
					t.Fatalf("error=%v, want %q", err, message)
				}
				for _, wrapped := range []error{err, fmt.Errorf("outer: %w", err), errors.Join(errors.New("not found already exists not configured"), err)} {
					for _, cause := range causes {
						if errors.Is(wrapped, cause) != (cause == want) {
							t.Fatalf("cause=%v in %v, want %v", cause, wrapped, want)
						}
					}
				}
				for _, cause := range causes {
					if errors.Is(errors.New(message), cause) {
						t.Fatal("same-text error acquired sentinel identity")
					}
				}
			}
			duplicate := fmt.Sprintf("suppression %q already exists", id)
			assertCause(manager.Add(original), alert.ErrSuppressionAlreadyExists, duplicate)
			missing := "missing " + id
			assertCause(manager.Update(missing, original), alert.ErrSuppressionNotFound, fmt.Sprintf("suppression %q not found", missing))
			assertCause(manager.Delete(missing), alert.ErrSuppressionNotFound, fmt.Sprintf("suppression %q not found", missing))
			assertCause(manager.ReloadFromFile(), alert.ErrSuppressionsFileNotConfigured, "suppressions file is not configured")
			duplicates := []alert.Suppression{original, original}
			if rejected, err := alert.NewSuppressionManager(duplicates); rejected != nil {
				t.Fatal("duplicate constructor returned manager")
			} else {
				assertCause(err, alert.ErrSuppressionAlreadyExists, duplicate)
			}
			if rejected, err := alert.NewSuppressionManagerWithFile(duplicates, path); rejected != nil {
				t.Fatal("duplicate file constructor returned manager")
			} else {
				assertCause(err, alert.ErrSuppressionAlreadyExists, duplicate)
			}
			assertCause(alert.SaveSuppressionsToFile(path, duplicates), alert.ErrSuppressionAlreadyExists, duplicate)
			if _, err := os.Stat(filepath.Dir(path)); !errors.Is(err, os.ErrNotExist) {
				t.Fatalf("rejected constructor/save created parent: %v", err)
			}
			invalid := original
			invalid.AnyCIDRs = []string{"bad-cidr"}
			assertCause(manager.Update(id, invalid), nil, fmt.Sprintf("suppression %s any cidrs: invalid IP or CIDR %q", id, "bad-cidr"))
			afterInput, err := json.Marshal(input)
			if err != nil || string(afterInput) != string(beforeInput) || !reflect.DeepEqual(manager.List(), input) || !reflect.DeepEqual(manager.Filter(probes), probes[1:]) || *probes[0] != probeBefore[0] || *probes[1] != probeBefore[1] {
				t.Fatal("failed operations changed rules, filter or callers")
			}
		})
	}
}
