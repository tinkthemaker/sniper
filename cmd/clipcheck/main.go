// Command clipcheck verifies on a real Windows desktop that text copied through
// Sniper's clipboard wrapper reads back byte-for-byte (after line-ending conversion).
package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"

	"github.com/tinkthemaker/sniper/internal/clipboard"
)

var corpus = []string{
	"plain ASCII",
	"line one\nline two\n\nline four",
	"tab\tseparated\tvalues",
	"trailing spaces   \n  leading spaces",
	"emoji \U0001F44B\U0001F3FD family \U0001F468\u200d\U0001F469\u200d\U0001F467 flag \U0001F1EF\U0001F1F5",
	"punctuation “quotes” ‘single’ — en–dash … © ® ™ € £ ¥",
	"日本語のテキスト / 中文文本 / 한국어 텍스트",
	"العربية مع English mixed — עברית",
	"Ελληνικά, Кириллица, हिन्दी, ไทย",
	"combining: e\u0301 n\u0303 (decomposed)",
	"zero width:\u200bjoiner\u200d end",
	`json-ish: {"a": "b\\n"} <tag attr='x'> & %s %d`,
}

func main() {
	a := app.New()
	w := a.NewWindow("clipcheck")
	w.Resize(fyne.NewSize(1, 1))
	failed := false
	a.Lifecycle().SetOnStarted(func() {
		go func() {
			failed = !run(a)
			fyne.Do(a.Quit)
		}()
	})
	w.ShowAndRun()
	if failed {
		os.Exit(1)
	}
}

func run(a fyne.App) bool {
	cb := clipboard.New(a.Clipboard())
	ok := true
	for i, s := range corpus {
		n := i + 1
		fyne.DoAndWait(func() { cb.SetText(s) })
		got, err := readClipboard()
		if err != nil {
			fmt.Printf("FAIL %d: cannot read clipboard: %v\n", n, err)
			ok = false
			continue
		}
		if want := clipboard.ToPlatform(s); got != want {
			fmt.Printf("FAIL %d: want %q got %q\n", n, want, got)
			ok = false
			continue
		}
		fmt.Printf("PASS %d\n", n)
	}
	return ok
}

// readClipboard reads the clipboard through PowerShell and returns it as UTF-8 text.
func readClipboard() (string, error) {
	script := "$t = Get-Clipboard -Raw; [Convert]::ToBase64String([Text.Encoding]::UTF8.GetBytes($t))"
	out, err := exec.Command("powershell", "-NoProfile", "-Command", script).Output()
	if err != nil {
		return "", err
	}
	raw, err := base64.StdEncoding.DecodeString(strings.TrimSpace(string(out)))
	if err != nil {
		return "", err
	}
	return string(raw), nil
}
