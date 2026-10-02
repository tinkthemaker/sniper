package snippets

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"unicode/utf8"
)

const (
	SchemaVersion = 1
	MaxTitleRunes = 200
	MaxBodyBytes  = 65536
	MaxSnippets   = 10000
)

// Snippet is one saved piece of text.
type Snippet struct {
	ID    string `json:"id"`
	Title string `json:"title"`
	Body  string `json:"body"`
}

// Library is the whole collection, in creation order.
type Library struct {
	Version  int       `json:"version"`
	Snippets []Snippet `json:"snippets"`
}

// ValidationError describes one rule violation. Field is "title", "body", "id" or "library".
type ValidationError struct {
	Field   string
	Message string // user-facing sentence, e.g. "Title is required."
}

func (e *ValidationError) Error() string { return e.Message }

// ErrNotFound is returned when an ID is not in the library.
var ErrNotFound = errors.New("snippet not found")

// NewLibrary returns an empty library with a non-nil slice.
func NewLibrary() Library {
	return Library{Version: SchemaVersion, Snippets: []Snippet{}}
}

// NewID returns 32 lowercase hex characters from 16 random bytes.
func NewID() (string, error) {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		return "", fmt.Errorf("generate snippet id: %w", err)
	}
	return hex.EncodeToString(b[:]), nil
}

// Normalize applies the title and body rules from the product spec.
func Normalize(title, body string) (string, string) {
	title = strings.NewReplacer("\r", " ", "\n", " ").Replace(title)
	title = strings.TrimSpace(title)
	body = strings.ReplaceAll(body, "\r\n", "\n")
	body = strings.ReplaceAll(body, "\r", "\n")
	return title, body
}

// ValidateFields checks normalized title and body. It returns a *ValidationError.
func ValidateFields(title, body string) error {
	switch {
	case !utf8.ValidString(title) || strings.ContainsRune(title, 0):
		return &ValidationError{"title", "Title contains characters that cannot be saved."}
	case strings.TrimSpace(title) == "":
		return &ValidationError{"title", "Title is required."}
	case utf8.RuneCountInString(title) > MaxTitleRunes:
		return &ValidationError{"title", fmt.Sprintf("Title is too long (%s characters; limit %s).",
			thousands(utf8.RuneCountInString(title)), thousands(MaxTitleRunes))}
	case !utf8.ValidString(body) || strings.ContainsRune(body, 0):
		return &ValidationError{"body", "Body contains characters that cannot be saved."}
	case strings.TrimSpace(body) == "":
		return &ValidationError{"body", "Body is required."}
	case len(body) > MaxBodyBytes:
		return &ValidationError{"body", fmt.Sprintf("Body is too long (%s bytes; limit %s).",
			thousands(len(body)), thousands(MaxBodyBytes))}
	}
	return nil
}

// ValidateLibrary checks the version, the count, every record, and ID uniqueness.
func ValidateLibrary(lib Library) error {
	if lib.Version != SchemaVersion {
		return &ValidationError{"library", fmt.Sprintf("Unsupported library version %d.", lib.Version)}
	}
	if len(lib.Snippets) > MaxSnippets {
		return &ValidationError{"library", libraryFull()}
	}
	seen := make(map[string]struct{}, len(lib.Snippets))
	for i, s := range lib.Snippets {
		if s.ID == "" {
			return &ValidationError{"id", fmt.Sprintf("Snippet %d has no ID.", i+1)}
		}
		if _, dup := seen[s.ID]; dup {
			return &ValidationError{"id", fmt.Sprintf("Snippet %d has a duplicate ID.", i+1)}
		}
		seen[s.ID] = struct{}{}
		if err := ValidateFields(s.Title, s.Body); err != nil {
			var ve *ValidationError
			errors.As(err, &ve)
			return &ValidationError{ve.Field, fmt.Sprintf("Snippet %d: %s", i+1, ve.Message)}
		}
	}
	return nil
}

// Add appends a new snippet. The input library is not modified.
func Add(lib Library, title, body string) (Library, Snippet, error) {
	title, body = Normalize(title, body)
	if err := ValidateFields(title, body); err != nil {
		return lib, Snippet{}, err
	}
	if len(lib.Snippets) >= MaxSnippets {
		return lib, Snippet{}, &ValidationError{"library", libraryFull()}
	}
	id, err := freshID(lib.Snippets)
	if err != nil {
		return lib, Snippet{}, err
	}
	item := Snippet{ID: id, Title: title, Body: body}
	return withSnippets(lib, append(clone(lib.Snippets, 1), item)), item, nil
}

// Update replaces the title and body of the snippet with the given ID.
func Update(lib Library, id, title, body string) (Library, Snippet, error) {
	title, body = Normalize(title, body)
	if err := ValidateFields(title, body); err != nil {
		return lib, Snippet{}, err
	}
	i := indexOf(lib.Snippets, id)
	if i < 0 {
		return lib, Snippet{}, ErrNotFound
	}
	next := clone(lib.Snippets, 0)
	next[i] = Snippet{ID: id, Title: title, Body: body}
	return withSnippets(lib, next), next[i], nil
}

// Delete removes the snippet with the given ID.
func Delete(lib Library, id string) (Library, error) {
	i := indexOf(lib.Snippets, id)
	if i < 0 {
		return lib, ErrNotFound
	}
	next := clone(lib.Snippets, 0)
	next = append(next[:i], next[i+1:]...)
	return withSnippets(lib, next), nil
}

// Append adds items in order with fresh IDs. Titles and bodies are normalized and
// validated; the input library is not modified.
func Append(lib Library, items []Snippet) (Library, error) {
	if len(lib.Snippets)+len(items) > MaxSnippets {
		return lib, &ValidationError{"library", libraryFull()}
	}
	next := clone(lib.Snippets, len(items))
	for i, it := range items {
		title, body := Normalize(it.Title, it.Body)
		if err := ValidateFields(title, body); err != nil {
			var ve *ValidationError
			errors.As(err, &ve)
			return lib, &ValidationError{ve.Field, fmt.Sprintf("record %d: %s", i+1, ve.Message)}
		}
		id, err := freshID(next)
		if err != nil {
			return lib, err
		}
		next = append(next, Snippet{ID: id, Title: title, Body: body})
	}
	return withSnippets(lib, next), nil
}

func libraryFull() string {
	return fmt.Sprintf("The library is full (limit %s snippets).", thousands(MaxSnippets))
}

func withSnippets(lib Library, s []Snippet) Library {
	lib.Snippets = s
	return lib
}

// clone copies s into a fresh slice with room for extra more elements.
func clone(s []Snippet, extra int) []Snippet {
	out := make([]Snippet, len(s), len(s)+extra)
	copy(out, s)
	return out
}

func indexOf(s []Snippet, id string) int {
	for i := range s {
		if s[i].ID == id {
			return i
		}
	}
	return -1
}

func freshID(existing []Snippet) (string, error) {
	for {
		id, err := NewID()
		if err != nil {
			return "", err
		}
		if indexOf(existing, id) < 0 {
			return id, nil
		}
	}
}

// thousands formats n with comma separators, e.g. 65536 -> "65,536".
func thousands(n int) string {
	s := fmt.Sprint(n)
	for i := len(s) - 3; i > 0; i -= 3 {
		s = s[:i] + "," + s[i:]
	}
	return s
}
