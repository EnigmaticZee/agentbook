package ui

import "testing"

func TestClassify(t *testing.T) {
	cases := []struct {
		line  string
		first bool
		kind  Kind
		text  string
	}{
		{"Agentbook check for App (layout: single)", true, KTitle, "Agentbook check for App (layout: single)"},
		{"  ✓ workspace: everything committed", false, KOK, "workspace: everything committed"},
		{"  ✗ no session log for today", false, KBad, "no session log for today"},
		{"  - tests skipped", false, KSkip, "tests skipped"},
		{"  ? src/a.go:3: claude", false, KWarn, "src/a.go:3: claude"},
		{"== Big files in history", false, KSection, "Big files in history"},
		{"      FAIL: TestX", false, KDetail, "FAIL: TestX"},
		{"All clear: safe to end the session.", false, KVerdict, "All clear: safe to end the session."},
		{"Existing code repo found: moving it", true, KPlain, "Existing code repo found: moving it"},
		{"", false, KBlank, ""},
	}
	for _, c := range cases {
		k, txt := Classify(c.line, c.first)
		if k != c.kind || txt != c.text {
			t.Errorf("Classify(%q) = %v %q, want %v %q", c.line, k, txt, c.kind, c.text)
		}
	}
}

func TestParseFind(t *testing.T) {
	out := []string{`== "sqlite"`, "history/sessions/2026-11-10.md:6:### Q: SQLite or Postgres? a:b", `== "zebra"`, "  (nothing recorded)"}
	res := ParseFind(out)
	if len(res["sqlite"]) != 1 || res["sqlite"][0].Line != "6" || res["sqlite"][0].Text != "### Q: SQLite or Postgres? a:b" {
		t.Fatalf("sqlite hits wrong: %+v", res["sqlite"])
	}
	if hits, ok := res["zebra"]; !ok || len(hits) != 0 {
		t.Fatalf("zebra should be present with no hits: %+v", res)
	}
}

func TestTruncate(t *testing.T) {
	if got := Truncate("hello world", 6); got != "hello…" {
		t.Fatalf("got %q", got)
	}
	if got := Truncate("short", 10); got != "short" {
		t.Fatalf("got %q", got)
	}
}
