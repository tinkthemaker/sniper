package snippets

// corpus is the shared Unicode round-trip corpus. Invisible and combining
// characters are written as escapes so editors cannot alter them.
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
