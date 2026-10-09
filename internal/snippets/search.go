package snippets

import (
	"sort"
	"strings"
)

type hit struct {
	index int
	tier  int
	title string // lowercased
}

// Search returns the snippets matching query, ranked per the product spec.
// The result is a new, non-nil slice.
func Search(items []Snippet, query string) []Snippet {
	terms := strings.Fields(strings.ToLower(query))
	hits := make([]hit, 0, len(items))
	for i, s := range items {
		title := strings.ToLower(s.Title)
		tier, ok := rank(title, s.Body, terms)
		if ok {
			hits = append(hits, hit{i, tier, title})
		}
	}
	sort.Slice(hits, func(a, b int) bool {
		x, y := hits[a], hits[b]
		if x.tier != y.tier {
			return x.tier < y.tier
		}
		if x.title != y.title {
			return x.title < y.title
		}
		return x.index < y.index
	})
	out := make([]Snippet, len(hits))
	for i, h := range hits {
		out[i] = items[h.index]
	}
	return out
}

// rank returns tier 0 when every term is in the title, tier 1 when the rest are
// found in the body, and ok=false when some term appears in neither.
func rank(lowerTitle, body string, terms []string) (tier int, ok bool) {
	var missing []string
	for _, t := range terms {
		if !strings.Contains(lowerTitle, t) {
			missing = append(missing, t)
		}
	}
	if len(missing) == 0 {
		return 0, true
	}
	lowerBody := strings.ToLower(body)
	for _, t := range missing {
		if !strings.Contains(lowerBody, t) {
			return 0, false
		}
	}
	return 1, true
}
