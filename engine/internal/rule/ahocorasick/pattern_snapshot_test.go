package ahocorasick

import (
	"fmt"
	"reflect"
	"sync"
	"testing"
)

func TestPatternsSnapshotsAreIndependentlyOwned(t *testing.T) {
	for _, insensitive := range []bool{false, true} {
		t.Run(fmt.Sprintf("insensitive_%t", insensitive), func(t *testing.T) {
			input := []string{"SHE", "HE", "HE", "ABSENT"}
			matcher := NewMatcher(input, insensitive)
			want := []string{"SHE", "HE", "HE", "ABSENT"}
			if insensitive {
				want = []string{"she", "he", "he", "absent"}
			}
			assertUnchanged := func() {
				t.Helper()
				if got := matcher.Patterns(); !reflect.DeepEqual(got, want) {
					t.Fatalf("patterns = %v, want %v", got, want)
				}
				if got := matcher.Match([]byte("SHE HE")); !reflect.DeepEqual(got, []int{0, 1, 2}) {
					t.Fatalf("duplicate/suffix hits = %v, want [0 1 2]", got)
				}
				if got := matcher.Match([]byte("nothing")); len(got) != 0 {
					t.Fatalf("unexpected no-hit matches %v", got)
				}
			}
			assertUnchanged()
			input[0], input[1] = "input changed", "input changed"
			assertUnchanged()
			first, second := matcher.Patterns(), matcher.Patterns()
			first[0], first[1] = "snapshot changed", "snapshot changed"
			if !reflect.DeepEqual(second, want) {
				t.Fatalf("first getter mutation changed second snapshot: %v", second)
			}
			assertUnchanged()
			first = first[1:]
			first[0] = "resliced snapshot changed"
			first = append(first, "appended snapshot")
			first[len(first)-1] = "appended snapshot changed"
			second[2] = "second snapshot changed"
			assertUnchanged()
		})
	}
	for _, input := range [][]string{nil, {}} {
		matcher := NewMatcher(input, false)
		if got := matcher.Patterns(); got == nil || len(got) != 0 {
			t.Fatalf("empty getter = %#v, want non-nil empty slice", got)
		}
		if got := matcher.Match([]byte("anything")); len(got) != 0 {
			t.Fatalf("empty matcher hits = %v", got)
		}
	}
}

func TestConcurrentPatternSnapshotsRemainIndependent(t *testing.T) {
	matcher := NewMatcher([]string{"SHE", "HE", "HE", "ABSENT"}, true)
	want := []string{"she", "he", "he", "absent"}
	const writers, observations = 4, 100
	start := make(chan struct{})
	failures := make(chan error, writers)
	var done sync.WaitGroup
	done.Add(writers)
	for writer := 0; writer < writers; writer++ {
		go func() {
			defer done.Done()
			<-start
			for observation := 0; observation < observations; observation++ {
				snapshot := matcher.Patterns()
				if !reflect.DeepEqual(snapshot, want) {
					failures <- fmt.Errorf("concurrent snapshot = %v, want %v", snapshot, want)
					return
				}
				snapshot[0], snapshot[1] = "private edit", "private edit"
				if got := matcher.Match([]byte("sHe HE")); !reflect.DeepEqual(got, []int{0, 1, 2}) {
					failures <- fmt.Errorf("concurrent hits = %v, want [0 1 2]", got)
					return
				}
			}
		}()
	}
	close(start)
	done.Wait()
	close(failures)
	for err := range failures {
		t.Error(err)
	}
	if got := matcher.Patterns(); !reflect.DeepEqual(got, want) {
		t.Fatalf("final snapshot = %v, want %v", got, want)
	}
	if got := matcher.Match([]byte("sHe HE")); !reflect.DeepEqual(got, []int{0, 1, 2}) {
		t.Fatalf("final hits = %v, want [0 1 2]", got)
	}
}
