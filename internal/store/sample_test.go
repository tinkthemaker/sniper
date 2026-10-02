package store

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/tinkthemaker/sniper/internal/snippets"
)

var corpusLabels = []string{
	"plain ASCII", "multiline", "tabs", "spaces", "emoji", "punctuation", "CJK",
	"RTL mixed", "Greek, Cyrillic, Hindi, Thai", "combining marks", "zero width", "json-ish",
}

// TestSampleImportFile keeps testdata/sample-import.json equal to the corpus, one
// snippet per string, so a human can import it instead of typing the cases.
// Run with SNIPER_WRITE_SAMPLE=1 to regenerate the file.
func TestSampleImportFile(t *testing.T) {
	lib := snippets.NewLibrary()
	for i, body := range corpus {
		lib.Snippets = append(lib.Snippets, snippets.Snippet{
			ID:    fmt.Sprintf("%032x", i+1),
			Title: fmt.Sprintf("Corpus %02d – %s", i+1, corpusLabels[i]),
			Body:  body,
		})
	}
	var buf bytes.Buffer
	if err := Export(&buf, lib); err != nil {
		t.Fatal(err)
	}
	path := filepath.Join("..", "..", "testdata", "sample-import.json")
	if os.Getenv("SNIPER_WRITE_SAMPLE") == "1" {
		if err := os.WriteFile(path, buf.Bytes(), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	got, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("%v (run with SNIPER_WRITE_SAMPLE=1 to create it)", err)
	}
	// Compare ignoring line endings: git may check the file out with CRLF.
	if !bytes.Equal(bytes.ReplaceAll(got, []byte("\r\n"), []byte("\n")), buf.Bytes()) {
		t.Error("testdata/sample-import.json is stale; run with SNIPER_WRITE_SAMPLE=1")
	}
	if _, err := DecodeImport(bytes.NewReader(got)); err != nil {
		t.Errorf("sample file does not import: %v", err)
	}
}
