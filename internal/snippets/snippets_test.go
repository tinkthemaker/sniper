package snippets

import (
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strings"
	"testing"
)

func fieldOf(t *testing.T, err error) string {
	t.Helper()
	var ve *ValidationError
	if !errors.As(err, &ve) {
		t.Fatalf("error %v is not a *ValidationError", err)
	}
	return ve.Field
}

func TestNormalize(t *testing.T) {
	title, body := Normalize("  Hello\r\nWorld\rX  ", "a\r\nb\rc\n  d  \t")
	if title != "Hello  World X" {
		t.Errorf("title = %q", title)
	}
	if body != "a\nb\nc\n  d  \t" {
		t.Errorf("body = %q (must not be trimmed)", body)
	}
}

func TestValidateFields(t *testing.T) {
	tests := []struct {
		name        string
		title, body string
		field       string // empty means valid
		msg         string
	}{
		{"ok", "T", "B", "", ""},
		{"empty title", "", "B", "title", "Title is required."},
		{"blank title", " \t", "B", "title", "Title is required."},
		{"200 runes ok", strings.Repeat("日", 200), "B", "", ""},
		{"201 runes", strings.Repeat("日", 201), "B", "title", "Title is too long (201 characters; limit 200)."},
		{"title bad utf8", "a\xffb", "B", "title", ""},
		{"title NUL", "a\x00b", "B", "title", ""},
		{"blank body", "T", " \n\t", "body", "Body is required."},
		{"body 65536 ok", "T", strings.Repeat("a", 65536), "", ""},
		{"body 65537", "T", strings.Repeat("a", 65537), "body", "Body is too long (65,537 bytes; limit 65,536)."},
		{"body bad utf8", "T", "a\xff", "body", ""},
		{"body NUL", "T", "a\x00", "body", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFields(tt.title, tt.body)
			if tt.field == "" {
				if err != nil {
					t.Fatalf("unexpected error: %v", err)
				}
				return
			}
			if err == nil {
				t.Fatal("expected an error")
			}
			if got := fieldOf(t, err); got != tt.field {
				t.Errorf("field = %q, want %q", got, tt.field)
			}
			if tt.msg != "" && err.Error() != tt.msg {
				t.Errorf("message = %q, want %q", err.Error(), tt.msg)
			}
		})
	}
}

func TestValidateLibrary(t *testing.T) {
	good := Library{Version: 1, Snippets: []Snippet{{"a", "T", "B"}, {"b", "T2", "B2"}}}
	if err := ValidateLibrary(good); err != nil {
		t.Fatalf("good library: %v", err)
	}
	if err := ValidateLibrary(NewLibrary()); err != nil {
		t.Fatalf("empty library: %v", err)
	}

	many := make([]Snippet, MaxSnippets+1)
	for i := range many {
		many[i] = Snippet{fmt.Sprint(i), "T", "B"}
	}
	tests := []struct {
		name  string
		lib   Library
		field string
	}{
		{"version", Library{Version: 2}, "library"},
		{"too many", Library{Version: 1, Snippets: many}, "library"},
		{"empty id", Library{Version: 1, Snippets: []Snippet{{"", "T", "B"}}}, "id"},
		{"dup id", Library{Version: 1, Snippets: []Snippet{{"a", "T", "B"}, {"a", "T", "B"}}}, "id"},
		{"bad record", Library{Version: 1, Snippets: []Snippet{{"a", "T", "B"}, {"b", "", "B"}}}, "title"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateLibrary(tt.lib)
			if err == nil {
				t.Fatal("expected an error")
			}
			if got := fieldOf(t, err); got != tt.field {
				t.Errorf("field = %q, want %q", got, tt.field)
			}
		})
	}
	err := ValidateLibrary(tests[4].lib)
	if !strings.HasPrefix(err.Error(), "Snippet 2: ") {
		t.Errorf("record error should name the record: %q", err)
	}
}

func TestNewID(t *testing.T) {
	re := regexp.MustCompile(`^[0-9a-f]{32}$`)
	seen := make(map[string]struct{}, 10000)
	for i := 0; i < 10000; i++ {
		id, err := NewID()
		if err != nil {
			t.Fatal(err)
		}
		if !re.MatchString(id) {
			t.Fatalf("bad id %q", id)
		}
		if _, dup := seen[id]; dup {
			t.Fatalf("duplicate id %q", id)
		}
		seen[id] = struct{}{}
	}
}

func deepCopy(lib Library) Library {
	return Library{Version: lib.Version, Snippets: append([]Snippet{}, lib.Snippets...)}
}

func TestAddUpdateDelete(t *testing.T) {
	// Spare capacity makes accidental backing-array sharing visible.
	base := Library{Version: 1, Snippets: make([]Snippet, 1, 10)}
	base.Snippets[0] = Snippet{"one", "One", "1"}
	orig := deepCopy(base)

	added, item, err := Add(base, "  Two\n", "2\r\n2")
	if err != nil {
		t.Fatal(err)
	}
	if item.Title != "Two" || item.Body != "2\n2" || len(item.ID) != 32 {
		t.Errorf("added item = %+v", item)
	}
	if len(added.Snippets) != 2 || added.Snippets[1] != item {
		t.Errorf("added = %+v", added)
	}
	if !reflect.DeepEqual(base, orig) {
		t.Errorf("Add modified its input")
	}
	_ = append(added.Snippets, Snippet{ID: "x"}) // must not leak into base's array
	if base.Snippets[:2][1].ID != "" {
		t.Errorf("Add shares the backing array with its input")
	}

	updated, up, err := Update(added, item.ID, "Two!", "22")
	if err != nil {
		t.Fatal(err)
	}
	if up.ID != item.ID || updated.Snippets[1].Title != "Two!" || updated.Snippets[1].Body != "22" {
		t.Errorf("updated = %+v", updated)
	}
	if added.Snippets[1].Title != "Two" {
		t.Errorf("Update modified its input")
	}
	if _, _, err := Update(added, "nope", "T", "B"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Update unknown id: %v", err)
	}
	if _, _, err := Update(added, item.ID, "", "B"); fieldOf(t, err) != "title" {
		t.Errorf("Update invalid: %v", err)
	}

	snapshot := deepCopy(updated)
	deleted, err := Delete(updated, "one")
	if err != nil {
		t.Fatal(err)
	}
	if len(deleted.Snippets) != 1 || deleted.Snippets[0].ID != item.ID {
		t.Errorf("deleted = %+v", deleted)
	}
	if !reflect.DeepEqual(updated, snapshot) {
		t.Errorf("Delete modified its input")
	}
	if _, err := Delete(updated, "nope"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Delete unknown id: %v", err)
	}
	if _, _, err := Add(base, "", "B"); fieldOf(t, err) != "title" {
		t.Errorf("Add invalid: %v", err)
	}

	full := NewLibrary()
	for i := 0; i < MaxSnippets; i++ {
		full.Snippets = append(full.Snippets, Snippet{fmt.Sprint(i), "T", "B"})
	}
	if _, _, err := Add(full, "T", "B"); fieldOf(t, err) != "library" {
		t.Errorf("Add to full library: %v", err)
	}
}

func TestAppend(t *testing.T) {
	base := Library{Version: 1, Snippets: []Snippet{{"keep", "Keep", "k"}}}
	orig := deepCopy(base)
	in := []Snippet{{"keep", "A\r\nB", "a\r\nb"}, {"", "C", "c"}}
	got, err := Append(base, in)
	if err != nil {
		t.Fatal(err)
	}
	if len(got.Snippets) != 3 || got.Snippets[0] != base.Snippets[0] {
		t.Fatalf("got = %+v", got)
	}
	if got.Snippets[1].ID == "keep" || got.Snippets[1].ID == "" || got.Snippets[2].ID == "" {
		t.Errorf("IDs not fresh: %+v", got.Snippets)
	}
	if got.Snippets[1].Title != "A  B" || got.Snippets[1].Body != "a\nb" || got.Snippets[2].Title != "C" {
		t.Errorf("order or normalization wrong: %+v", got.Snippets)
	}
	if !reflect.DeepEqual(base, orig) {
		t.Errorf("Append modified its input")
	}

	big := make([]Snippet, MaxSnippets)
	for i := range big {
		big[i] = Snippet{"", "T", "B"}
	}
	if _, err := Append(base, big); fieldOf(t, err) != "library" {
		t.Errorf("limit not enforced: %v", err)
	}
	if _, err := Append(base, []Snippet{{"", "ok", "ok"}, {"", "", "x"}}); err == nil ||
		!strings.Contains(err.Error(), "record 2:") {
		t.Errorf("invalid record error = %v", err)
	}
	if !reflect.DeepEqual(base, orig) {
		t.Errorf("failed Append modified its input")
	}
}

func TestCorpusSurvivesAdd(t *testing.T) {
	for _, s := range corpus {
		title := strings.ReplaceAll(s, "\n", " ")
		_, item, err := Add(NewLibrary(), title, s)
		if err != nil {
			t.Errorf("%q: %v", s, err)
			continue
		}
		if item.Body != s {
			t.Errorf("body changed: %q -> %q", s, item.Body)
		}
	}
}
