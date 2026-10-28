package ui

import (
	"strings"
	"testing"
)

func TestWrapKeepsWordsWhole(t *testing.T) {
	s := "the quick brown fox jumps over the lazy dog"
	lines := Wrap(s, 80)
	if len(lines) < 2 {
		t.Fatalf("expected wrapping, got %v", lines)
	}
	if strings.Join(lines, " ") != s {
		t.Fatalf("words lost or reordered: %v", lines)
	}
	for _, l := range lines {
		if Width(l) > 80 && strings.Contains(l, " ") {
			t.Fatalf("line too wide: %q", l)
		}
	}
}

func TestWrapHonoursNewlines(t *testing.T) {
	lines := Wrap("a\n\nb", 1000)
	if len(lines) != 3 || lines[1] != "" {
		t.Fatalf("got %q", lines)
	}
}

func TestTextBoxReveal(t *testing.T) {
	tb := NewTextBox()
	tb.CharsPerSecond = 10
	tb.Set("Old wizard", "hello world", 1000)
	if tb.Done() {
		t.Fatal("should start hidden")
	}
	tb.Advance(0.5)
	if tb.Done() {
		t.Fatal("5 of 11 chars should not be done")
	}
	tb.Advance(10)
	if !tb.Done() {
		t.Fatal("should be done after plenty of time")
	}
	tb.Set("", "x", 1000)
	tb.Finish()
	if !tb.Done() {
		t.Fatal("Finish should complete the page")
	}
}
