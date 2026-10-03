package alert

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestShardCleanupPreservesInvalidCalendarFiles(t *testing.T) {
	for _, startup := range []bool{false, true} {
		t.Run(fmt.Sprintf("startup_%t", startup), func(t *testing.T) {
			dir := t.TempDir()
			now := time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC)
			retained := make(map[string][]byte)
			var removed []string
			createFixtures := func() {
				for _, name := range []string{
					"netsentry-2025-02-29.db", "netsentry-2025-02-30.db",
					"netsentry-2025-04-31.db", "netsentry-2025-00-01.db",
					"netsentry-2025-13-01.db", "netsentry-2025-01-00.db",
					"netsentry-2025-01-32.db", "netsentry-2025-2-01.db",
					"notes-2025-01-01.db", "netsentry-2026-06-20.db",
					"netsentry-2026-06-21.db",
				} {
					for path, contents := range calendarFixtureSet(t, dir, name) {
						retained[path] = contents
					}
				}
				for _, name := range []string{
					"netsentry-2024-02-29.db", "netsentry-2026-06-19.db",
					"netsentry-0000-01-01.db",
				} {
					for path := range calendarFixtureSet(t, dir, name) {
						removed = append(removed, path)
					}
				}
				// Orphan sidecars are outside the base-file driven cleanup contract.
				for _, suffix := range []string{"-wal", "-shm"} {
					path := filepath.Join(dir, "netsentry-2020-01-01.db"+suffix)
					retained[path] = calendarFixtureFile(t, path)
				}
				folder := filepath.Join(dir, "netsentry-2020-01-02.db")
				if err := os.Mkdir(folder, 0o700); err != nil {
					t.Fatalf("create date-shaped directory: %v", err)
				}
				path := filepath.Join(folder, "nested")
				retained[path] = calendarFixtureFile(t, path)
			}
			if startup {
				createFixtures()
			}
			store, err := Open(context.Background(), Options{
				Dir: dir, DailyShard: true, JournalMode: "WAL", BusyTimeoutMS: 1000,
				AggregationWindow: time.Minute, RetentionDays: 7,
				Now: func() time.Time { return now },
			})
			if err != nil {
				t.Fatalf("open daily store: %v", err)
			}
			t.Cleanup(func() { _ = store.Close() })
			if !startup {
				createFixtures()
			}
			if len(removed) != 9 {
				t.Fatalf("expired fixture count = %d, want 9", len(removed))
			}
			deleted, err := store.PruneExpiredShardFiles(context.Background(), dir)
			want := 9
			if startup {
				want = 0 // Open already performed the same public cleanup.
			}
			if err != nil || deleted != want {
				t.Fatalf("cleanup = %d, %v, want %d, nil", deleted, err, want)
			}
			for _, path := range removed {
				assertFileDoesNotExist(t, path)
			}
			for path, contents := range retained {
				assertFileBytes(t, path, contents)
			}
			deleted, err = store.PruneExpiredShardFiles(context.Background(), dir)
			if err != nil || deleted != 0 {
				t.Fatalf("repeated cleanup = %d, %v, want 0, nil", deleted, err)
			}
			if _, err := os.Stat(store.Path()); err != nil {
				t.Fatalf("current primary missing: %v", err)
			}
			if err := store.Close(); err != nil {
				t.Fatalf("close daily store: %v", err)
			}
			for path, contents := range retained {
				assertFileBytes(t, path, contents)
			}
		})
	}
}

func TestShardCalendarCleanupDisabledAndCanceledPreserveFiles(t *testing.T) {
	for _, canceled := range []bool{false, true} {
		t.Run(fmt.Sprintf("canceled_%t", canceled), func(t *testing.T) {
			dir := t.TempDir()
			retention := 0
			if canceled {
				retention = 7
			}
			store, err := Open(context.Background(), Options{
				Dir: dir, DailyShard: true, JournalMode: "WAL", BusyTimeoutMS: 1000,
				AggregationWindow: time.Minute, RetentionDays: retention,
				Now: func() time.Time { return time.Date(2026, 6, 27, 12, 0, 0, 0, time.UTC) },
			})
			if err != nil {
				t.Fatalf("open daily store: %v", err)
			}
			defer store.Close()
			retained := calendarFixtureSet(t, dir, "netsentry-2024-02-29.db")
			for path, contents := range calendarFixtureSet(t, dir, "netsentry-2025-02-29.db") {
				retained[path] = contents
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			if canceled {
				cancel()
			}
			deleted, err := store.PruneExpiredShardFiles(ctx, dir)
			if deleted != 0 || (canceled && !errors.Is(err, context.Canceled)) || (!canceled && err != nil) {
				t.Fatalf("cleanup = %d, %v; canceled=%t", deleted, err, canceled)
			}
			for path, contents := range retained {
				assertFileBytes(t, path, contents)
			}
		})
	}
}

func calendarFixtureSet(t *testing.T, dir, name string) map[string][]byte {
	t.Helper()
	files := make(map[string][]byte)
	for _, suffix := range []string{"", "-wal", "-shm"} {
		path := filepath.Join(dir, name+suffix)
		files[path] = calendarFixtureFile(t, path)
	}
	return files
}

func calendarFixtureFile(t *testing.T, path string) []byte {
	t.Helper()
	contents := append([]byte("arbitrary preserved bytes: "+filepath.Base(path)), 0, 0xff, '\n')
	if err := os.WriteFile(path, contents, 0o600); err != nil {
		t.Fatalf("write fixture %s: %v", path, err)
	}
	return contents
}
