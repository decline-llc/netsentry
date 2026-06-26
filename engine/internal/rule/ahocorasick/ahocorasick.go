// Package ahocorasick provides a simple Aho-Corasick multi-pattern matcher.
// It is a self-contained implementation with no external dependencies.
package ahocorasick

import "strings"

type node struct {
	children [256]*node
	fail     *node
	output   []int // indices into the pattern slice
}

// Matcher is a compiled Aho-Corasick automaton.
type Matcher struct {
	root            *node
	patterns        []string
	caseInsensitive bool
}

// NewMatcher builds an Aho-Corasick automaton from the given patterns.
// If caseInsensitive is true, all patterns and inputs are lower-cased before matching.
func NewMatcher(patterns []string, caseInsensitive bool) *Matcher {
	root := &node{}
	normalized := make([]string, len(patterns))
	for i, p := range patterns {
		if caseInsensitive {
			p = strings.ToLower(p)
		}
		normalized[i] = p
	}

	// Build trie.
	for i, p := range normalized {
		cur := root
		for j := 0; j < len(p); j++ {
			c := p[j]
			if cur.children[c] == nil {
				cur.children[c] = &node{}
			}
			cur = cur.children[c]
		}
		cur.output = append(cur.output, i)
	}

	// Build failure links via BFS.
	queue := make([]*node, 0, 64)
	for _, child := range root.children {
		if child != nil {
			child.fail = root
			queue = append(queue, child)
		}
	}
	for len(queue) > 0 {
		cur := queue[0]
		queue = queue[1:]
		for c, child := range cur.children {
			if child == nil {
				continue
			}
			// Walk failure chain to find longest proper suffix.
			fail := cur.fail
			for fail != nil && fail.children[c] == nil {
				fail = fail.fail
			}
			if fail == nil {
				child.fail = root
			} else {
				child.fail = fail.children[c]
				if child.fail == child {
					child.fail = root
				}
			}
			// Merge outputs from failure link.
			child.output = append(child.output, child.fail.output...)
			queue = append(queue, child)
		}
	}

	return &Matcher{root: root, patterns: normalized, caseInsensitive: caseInsensitive}
}

// Match returns the indices of all patterns found in text.
// When the matcher was built with caseInsensitive=true, text is automatically
// lower-cased before matching; the caller does not need to pre-lowercase.
func (m *Matcher) Match(text []byte) []int {
	if len(m.patterns) == 0 {
		return nil
	}
	if m.caseInsensitive {
		text = []byte(strings.ToLower(string(text)))
	}
	seen := make(map[int]struct{})
	var result []int
	cur := m.root
	for i := 0; i < len(text); i++ {
		c := text[i]
		for cur != m.root && cur.children[c] == nil {
			cur = cur.fail
		}
		if cur.children[c] != nil {
			cur = cur.children[c]
		}
		for _, idx := range cur.output {
			if _, ok := seen[idx]; !ok {
				seen[idx] = struct{}{}
				result = append(result, idx)
			}
		}
	}
	return result
}

// Patterns returns the (possibly normalised) patterns this matcher was built from.
func (m *Matcher) Patterns() []string {
	return m.patterns
}
