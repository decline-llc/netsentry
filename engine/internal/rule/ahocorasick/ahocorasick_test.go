package ahocorasick

import "testing"

func TestBasicMatch(t *testing.T) {
	m := NewMatcher([]string{"he", "she", "his", "hers"}, false)
	hits := m.Match([]byte("ushers"))
	// "he" matches at pos 2 (us-he-rs), "she" at pos 1 (u-she-rs), "hers" at pos 2.
	want := map[int]bool{0: true, 1: true, 3: true}
	if len(hits) != len(want) {
		t.Fatalf("got %v, want indices %v", hits, want)
	}
	for _, idx := range hits {
		if !want[idx] {
			t.Errorf("unexpected index %d", idx)
		}
	}
}

func TestCaseInsensitive(t *testing.T) {
	m := NewMatcher([]string{"union select"}, true)
	// Match auto-lowercases input when caseInsensitive=true.
	hits := m.Match([]byte("UNION SELECT * FROM users"))
	if len(hits) == 0 {
		t.Fatal("expected match, got none")
	}
}

func TestNoMatch(t *testing.T) {
	m := NewMatcher([]string{"abc"}, false)
	if hits := m.Match([]byte("xyz")); len(hits) != 0 {
		t.Fatalf("expected no match, got %v", hits)
	}
}

func TestEmptyPatterns(t *testing.T) {
	m := NewMatcher(nil, false)
	if hits := m.Match([]byte("anything")); len(hits) != 0 {
		t.Fatalf("expected no match, got %v", hits)
	}
}

func TestDuplicateMatch(t *testing.T) {
	m := NewMatcher([]string{"ab", "ab"}, false)
	hits := m.Match([]byte("ab"))
	// Both pattern indices should be returned.
	if len(hits) != 2 {
		t.Fatalf("expected 2 hits (one per pattern), got %v", hits)
	}
}
