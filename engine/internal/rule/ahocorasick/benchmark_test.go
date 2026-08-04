package ahocorasick

import (
	"bytes"
	"runtime"
	"testing"
)

func BenchmarkMatcherMatch(b *testing.B) {
	patterns := []string{
		"union select",
		"../",
		"/etc/passwd",
		"powershell",
		"cmd.exe",
		"/bin/sh",
		"wget ",
		"curl ",
		"authorization:",
		"user-agent:",
		"drop table",
		"or 1=1",
		"javascript:",
		"<script",
		"suspicious-marker",
		"malware-marker",
	}
	matcher := NewMatcher(patterns, true)

	cases := []struct {
		name    string
		payload []byte
		want    map[int]struct{}
	}{
		{
			name:    "no_hit",
			payload: bytes.Repeat([]byte("ordinary request body segment "), 32),
			want:    map[int]struct{}{},
		},
		{
			name: "multi_hit",
			payload: append(
				bytes.Repeat([]byte("ordinary request body segment "), 32),
				[]byte("UNION SELECT followed by ../etc/passwd and /bin/sh")...,
			),
			want: map[int]struct{}{0: {}, 1: {}, 2: {}, 5: {}},
		},
	}

	for _, benchmark := range cases {
		b.Run(benchmark.name, func(b *testing.B) {
			hits := matcher.Match(benchmark.payload)
			assertBenchmarkPatternHits(b, hits, benchmark.want)

			b.ReportAllocs()
			b.SetBytes(int64(len(benchmark.payload)))
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				hits = matcher.Match(benchmark.payload)
			}
			b.StopTimer()

			runtime.KeepAlive(hits)
			assertBenchmarkPatternHits(b, hits, benchmark.want)
		})
	}
}

func assertBenchmarkPatternHits(b *testing.B, got []int, want map[int]struct{}) {
	b.Helper()
	if len(got) != len(want) {
		b.Fatalf("got pattern indices %v, want %v", got, want)
	}
	seen := make(map[int]struct{}, len(got))
	for _, index := range got {
		if _, ok := want[index]; !ok {
			b.Fatalf("unexpected pattern index %d in %v", index, got)
		}
		if _, duplicate := seen[index]; duplicate {
			b.Fatalf("duplicate pattern index %d in %v", index, got)
		}
		seen[index] = struct{}{}
	}
}
