package store

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"sniper/internal/snippets"
)

const (
	FileName     = "snippets.json"
	BackupName   = "snippets.bak.json"
	LockName     = "sniper.lock"
	MaxFileBytes = 16 << 20
)

var (
	ErrTooLarge           = errors.New("file is larger than 16 MiB")
	ErrUnsupportedVersion = errors.New("file was created by an unsupported version of Sniper")
)

var utf8BOM = []byte{0xEF, 0xBB, 0xBF}

// Encode returns the schema-v1 file bytes: 2-space indent, readable <, > and &,
// and a trailing newline.
func Encode(lib snippets.Library) ([]byte, error) {
	if lib.Snippets == nil {
		lib.Snippets = []snippets.Snippet{}
	}
	var buf bytes.Buffer
	enc := json.NewEncoder(&buf)
	enc.SetIndent("", "  ")
	enc.SetEscapeHTML(false)
	if err := enc.Encode(lib); err != nil {
		return nil, fmt.Errorf("encode library: %w", err)
	}
	return buf.Bytes(), nil
}

// Decode reads a library file strictly and validates it. Stored data is not normalized.
func Decode(r io.Reader) (snippets.Library, error) {
	lib, err := decodeLibrary(r)
	if err != nil {
		return snippets.Library{}, err
	}
	if err := snippets.ValidateLibrary(lib); err != nil {
		return snippets.Library{}, err
	}
	return lib, nil
}

// DecodeImport reads a schema-v1 file for import. IDs are ignored (returned empty),
// and every record is normalized and validated.
func DecodeImport(r io.Reader) ([]snippets.Snippet, error) {
	lib, err := decodeLibrary(r)
	if err != nil {
		return nil, err
	}
	items := make([]snippets.Snippet, len(lib.Snippets))
	for i, s := range lib.Snippets {
		title, body := snippets.Normalize(s.Title, s.Body)
		if err := snippets.ValidateFields(title, body); err != nil {
			return nil, fmt.Errorf("record %d: %w", i+1, err)
		}
		items[i] = snippets.Snippet{Title: title, Body: body}
	}
	return items, nil
}

// Export writes the full library in the schema-v1 format.
func Export(w io.Writer, lib snippets.Library) error {
	data, err := Encode(lib)
	if err != nil {
		return err
	}
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("write export: %w", err)
	}
	return nil
}

// decodeLibrary applies the shared decoding rules without validating records.
func decodeLibrary(r io.Reader) (snippets.Library, error) {
	data, err := io.ReadAll(io.LimitReader(r, MaxFileBytes+1))
	if err != nil {
		return snippets.Library{}, fmt.Errorf("read file: %w", err)
	}
	if len(data) > MaxFileBytes {
		return snippets.Library{}, ErrTooLarge
	}
	data = bytes.TrimPrefix(data, utf8BOM)

	var head struct{ Version *int }
	if err := json.Unmarshal(data, &head); err != nil {
		return snippets.Library{}, fmt.Errorf("invalid JSON: %w", err)
	}
	if head.Version == nil {
		return snippets.Library{}, fmt.Errorf("%w (no version field)", ErrUnsupportedVersion)
	}
	if *head.Version != snippets.SchemaVersion {
		return snippets.Library{}, fmt.Errorf("%w (version %d)", ErrUnsupportedVersion, *head.Version)
	}

	dec := json.NewDecoder(bytes.NewReader(data))
	dec.DisallowUnknownFields()
	var lib snippets.Library
	if err := dec.Decode(&lib); err != nil {
		return snippets.Library{}, fmt.Errorf("invalid library file: %w", err)
	}
	var extra json.RawMessage
	if err := dec.Decode(&extra); err != io.EOF {
		return snippets.Library{}, errors.New("invalid library file: unexpected data after the library")
	}
	if lib.Snippets == nil {
		lib.Snippets = []snippets.Snippet{}
	}
	return lib, nil
}
