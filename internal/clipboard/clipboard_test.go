package clipboard

import (
	"runtime"
	"testing"
)

type fakeClipboard struct{ content string }

func (f *fakeClipboard) Content() string     { return f.content }
func (f *fakeClipboard) SetContent(s string) { f.content = s }

func TestToPlatform(t *testing.T) {
	crlf := runtime.GOOS == "windows"
	nl := "\n"
	if crlf {
		nl = "\r\n"
	}
	tests := []struct{ name, in, want string }{
		{"lf", "a\nb", "a" + nl + "b"},
		{"crlf not doubled", "a\r\nb", "a" + nl + "b"},
		{"no newline", "plain", "plain"},
		{"blank lines", "a\n\nb\n", "a" + nl + nl + "b" + nl},
		{"empty", "", ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := ToPlatform(tt.in); got != tt.want {
				t.Errorf("ToPlatform(%q) = %q, want %q", tt.in, got, tt.want)
			}
		})
	}
}

func TestSetTextConvertsLineEndings(t *testing.T) {
	fc := &fakeClipboard{}
	New(fc).SetText("one\ntwo")
	if want := ToPlatform("one\ntwo"); fc.content != want {
		t.Errorf("clipboard = %q, want %q", fc.content, want)
	}
}
