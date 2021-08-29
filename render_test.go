package gemini_test

import (
	"bytes"
	"fmt"
	"testing"

	. "github.com/tdemin/gmnhg"
)

var (
	divider = "============================================================"
)

func TestHr(t *testing.T) {
	t.Run("Dash", testRenderer([]byte("---"), []byte("---\n")))
	t.Run("Asterisk", testRenderer([]byte("***"), []byte("---\n")))
	t.Run("Underscore", testRenderer([]byte("___"), []byte("---\n")))
	t.Run("Inline", testRenderer([]byte("Test ---"), []byte("Test ---\n")))
	t.Run("Between paragraphs", testRenderer([]byte("Foo\n\n---\n\nBar"), []byte("Foo\n\n---\n\nBar\n")))
	t.Run("Adjacent", testRenderer([]byte("Foo\n\n---\n\n---\n\nBar"), []byte("Foo\n\n---\n\n---\n\nBar\n")))
}

func testRenderer(markdown []byte, expectedGemini []byte) func(*testing.T) {
	return func(t *testing.T) {
		geminiContent, _, err := RenderMarkdown(markdown, Defaults)
		if err != nil {
			t.Error(fmt.Errorf("Error during rendering: %w", err))
		}
		if !bytes.Equal(geminiContent, expectedGemini) {
			t.Error(fmt.Sprintf("Output does not match expected!\n\nActual output:\n\n%s\n%s\n%s\n\nExpected output:\n\n%s\n%s\n%s", divider, geminiContent, divider, divider, expectedGemini, divider))
		}
	}
}
