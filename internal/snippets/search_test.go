package snippets

import (
	"fmt"
	"reflect"
	"strings"
	"testing"
)

func titles(items []Snippet) []string {
	out := make([]string, len(items))
	for i, s := range items {
		out[i] = s.Title
	}
	return out
}

func TestSearch(t *testing.T) {
	items := []Snippet{
		{"1", "Zebra", "stripes and apple"},
		{"2", "apple pie", "dessert"},
		{"3", "Banana", "contains apple here"},
		{"4", "same", "first"},
		{"5", "Same", "second"},
		{"6", "日本語", "テキスト"},
		{"7", "Party 🎉", "confetti"},
	}
	tests := []struct {
		name, query string
		want        []string
	}{
		{"empty returns all sorted, stable for equal titles", "", []string{"apple pie", "Banana", "Party 🎉", "same", "Same", "Zebra", "日本語"}},
		{"whitespace query", "  \t ", []string{"apple pie", "Banana", "Party 🎉", "same", "Same", "Zebra", "日本語"}},
		{"title tier before body tier", "apple", []string{"apple pie", "Banana", "Zebra"}},
		{"case insensitive", "APPLE", []string{"apple pie", "Banana", "Zebra"}},
		{"AND across fields", "apple dessert", []string{"apple pie"}},
		{"AND across title and body", "banana apple", []string{"Banana"}},
		{"title-tier needs all terms in title", "apple pie", []string{"apple pie"}},
		{"equal titles keep file order", "same", []string{"same", "Same"}},
		{"non-Latin", "日本", []string{"日本語"}},
		{"non-Latin body", "テキ", []string{"日本語"}},
		{"emoji", "🎉", []string{"Party 🎉"}},
		{"no match", "zzz", []string{}},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Search(items, tt.query)
			if got == nil {
				t.Fatal("Search returned nil")
			}
			if !reflect.DeepEqual(titles(got), tt.want) {
				t.Errorf("got %v, want %v", titles(got), tt.want)
			}
		})
	}
	if got := Search(nil, "x"); got == nil || len(got) != 0 {
		t.Errorf("nil input: %v", got)
	}
	if first := Search(items, "same"); first[0].ID != "4" || first[1].ID != "5" {
		t.Errorf("file order not kept: %+v", first)
	}
}

func TestSearchReturnsNewSlice(t *testing.T) {
	items := []Snippet{{"1", "b", "x"}, {"2", "a", "x"}}
	got := Search(items, "")
	got[0].Title = "changed"
	if items[1].Title != "a" || items[0].Title != "b" {
		t.Errorf("Search aliased its input: %+v", items)
	}
}

func TestExcerpt(t *testing.T) {
	tests := []struct {
		name, body string
		max        int
		want       string
	}{
		{"skips blank lines", "\n  \n\t\n  hello  \nworld", 80, "hello"},
		{"short unchanged", "hello", 80, "hello"},
		{"cut with ellipsis", "abcdefghij", 5, "abcd…"},
		{"exact fit", "abcde", 5, "abcde"},
		{"rune boundary", "日本語のテキスト", 4, "日本語…"},
		{"emoji rune", "😀😀😀😀", 3, "😀😀…"},
		{"all blank", " \n \n", 10, ""},
		{"zero max", "abc", 0, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := Excerpt(tt.body, tt.max); got != tt.want {
				t.Errorf("got %q, want %q", got, tt.want)
			}
		})
	}
}

func BenchmarkSearch(b *testing.B) {
	items := make([]Snippet, 1000)
	filler := strings.Repeat("The Quick Brown Fox jumps over the lazy dog. ", 50)[:2048]
	for i := range items {
		items[i] = Snippet{
			ID:    fmt.Sprint(i),
			Title: fmt.Sprintf("Snippet number %d", i),
			Body:  filler + fmt.Sprintf(" tail%d", i),
		}
	}
	b.ReportAllocs()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if got := Search(items, "lazy tail5"); len(got) == 0 {
			b.Fatal("no results")
		}
	}
}
